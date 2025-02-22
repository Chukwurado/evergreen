package units

import (
	"context"
	"fmt"
	"time"

	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/apimodels"
	"github.com/evergreen-ci/evergreen/model"
	"github.com/evergreen-ci/evergreen/model/host"
	"github.com/evergreen-ci/evergreen/model/task"
	"github.com/evergreen-ci/utility"
	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/job"
	"github.com/mongodb/amboy/registry"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/mongodb/grip/sometimes"
	"github.com/pkg/errors"
)

const (
	heartbeatTimeoutThreshold             = 7 * time.Minute
	taskExecutionTimeoutJobName           = "task-execution-timeout"
	taskExecutionTimeoutPopulationJobName = "task-execution-timeout-populate"
	maxAttempts                           = 10
)

func init() {
	registry.AddJobType(taskExecutionTimeoutJobName, func() amboy.Job {
		return makeTaskExecutionTimeoutMonitorJob()
	})
	registry.AddJobType(taskExecutionTimeoutPopulationJobName, func() amboy.Job {
		return makeTaskExecutionTimeoutPopulateJob()
	})
}

type taskExecutionTimeoutJob struct {
	Task      string `bson:"task_id"`
	Execution int    `bson:"execution"`
	Attempt   int    `bson:"attempt"`

	successful bool
	job.Base   `bson:"metadata" json:"metadata" yaml:"metadata"`
}

func makeTaskExecutionTimeoutMonitorJob() *taskExecutionTimeoutJob {
	j := &taskExecutionTimeoutJob{
		Base: job.Base{
			JobType: amboy.JobType{
				Name:    taskExecutionTimeoutJobName,
				Version: 0,
			},
		},
	}
	return j
}

func NewTaskExecutionMonitorJob(taskID string, execution int, attempt int, ts string) amboy.Job {
	j := makeTaskExecutionTimeoutMonitorJob()
	j.Task = taskID
	j.Execution = execution
	j.Attempt = attempt
	j.SetID(fmt.Sprintf("%s.%s.%d.attempt-%d.%s", taskExecutionTimeoutJobName, taskID, execution, attempt, ts))
	return j
}

func (j *taskExecutionTimeoutJob) Run(ctx context.Context) {
	defer j.MarkComplete()

	flags, err := evergreen.GetServiceFlags()
	if err != nil {
		j.AddError(err)
		return
	}
	env := evergreen.GetEnvironment()

	if flags.MonitorDisabled {
		grip.InfoWhen(sometimes.Percent(evergreen.DegradedLoggingPercent), message.Fields{
			"message":   "monitor is disabled",
			"operation": j.Type().Name,
			"impact":    "skipping task heartbeat cleanup job",
			"mode":      "degraded",
		})
		return
	}
	defer j.tryRequeue(ctx, env)

	t, err := task.FindOneId(j.Task)
	if err != nil {
		j.AddError(errors.Wrap(err, "error finding task"))
		return
	}
	if t == nil {
		j.AddError(errors.New("no task found"))
		return
	}

	// if the task has heartbeat since this job was queued, let it run
	if t.LastHeartbeat.Add(heartbeatTimeoutThreshold).After(time.Now()) {
		j.successful = true
		return
	}

	msg := message.Fields{
		"operation": j.Type().Name,
		"id":        j.ID(),
		"task":      t.Id,
		"host_id":   t.HostId,
	}

	err = cleanUpTimedOutTask(ctx, env, j.ID(), t)
	if err != nil {
		grip.Warning(message.WrapError(err, msg))
		j.AddError(err)
	} else {
		j.successful = true
	}
	grip.Debug(msg)
}

func (j *taskExecutionTimeoutJob) tryRequeue(ctx context.Context, env evergreen.Environment) {
	if j.successful || j.Attempt >= maxAttempts {
		return
	}
	ts := utility.RoundPartOfHour(15)
	newJob := NewTaskExecutionMonitorJob(j.Task, j.Execution, j.Attempt+1, ts.Format(TSFormat))
	newJob.UpdateTimeInfo(amboy.JobTimeInfo{
		WaitUntil: time.Now().Add(time.Minute),
	})
	err := env.RemoteQueue().Put(ctx, newJob)
	grip.Error(message.WrapError(err, message.Fields{
		"message":  "failed to requeue task timeout job",
		"task":     j.Task,
		"job":      j.ID(),
		"attempts": j.Attempt,
	}))
	j.AddError(err)
}

// function to clean up a single task
func cleanUpTimedOutTask(ctx context.Context, env evergreen.Environment, id string, t *task.Task) error {
	// get the host for the task
	host, err := host.FindOne(host.ById(t.HostId))
	if err != nil {
		return errors.Wrapf(err, "error finding host %s for task %s",
			t.HostId, t.Id)
	}

	// if there's no relevant host and the task is not a display task, something went wrong
	if host == nil {
		if !t.DisplayOnly {
			grip.Error(message.Fields{
				"message":   "no entry found for host",
				"task":      t.Id,
				"host_id":   t.HostId,
				"operation": "cleanup timed out task",
			})
		}
		return errors.WithStack(t.MarkUnscheduled())
	}

	// For a single-host task group, if a task fails, block and dequeue later tasks in that group.
	if t.IsPartOfSingleHostTaskGroup() && t.Status != evergreen.TaskSucceeded {
		if err = model.BlockTaskGroupTasks(t.Id); err != nil {
			grip.Error(message.WrapError(err, message.Fields{
				"message": "problem blocking task group tasks",
				"task_id": t.Id,
			}))
		}
		grip.Debug(message.Fields{
			"message": "blocked task group tasks for task",
			"task_id": t.Id,
		})
	}

	// if the host still has the task as its running task, clear it.
	if host.RunningTask == t.Id {
		// Check if the host was externally terminated. When the running task is
		// cleared on the host, an agent or agent monitor deploy might run,
		// which updates the LCT and prevents detection of external termination
		// until the deploy job runs out of retries.
		var terminated bool
		terminated, err = handleExternallyTerminatedHost(ctx, id, env, host)
		if err != nil {
			return errors.Wrap(err, "could not check host with timed out task for external termination")
		}
		if terminated {
			return nil
		}
		if err = host.ClearRunningAndSetLastTask(t); err != nil {
			return errors.Wrapf(err, "error clearing running task %s from host %s", t.Id, host.Id)
		}
	}

	detail := &apimodels.TaskEndDetail{
		Description: evergreen.TaskDescriptionHeartbeat,
		Type:        evergreen.CommandTypeSystem,
		TimedOut:    true,
		Status:      evergreen.TaskFailed,
	}

	// try to reset the task
	if t.IsPartOfDisplay() {
		dt, err := t.GetDisplayTask()
		if err != nil {
			return errors.Wrapf(err, "error getting display task")
		}
		if err = dt.SetResetWhenFinished(); err != nil {
			return errors.Wrap(err, "can't mark display task for reset")
		}
		return errors.Wrap(model.MarkEnd(t, "monitor", time.Now(), detail, false), "error marking execution task ended")
	}

	return errors.Wrapf(model.TryResetTask(t.Id, "", "monitor", detail), "error trying to reset task %s", t.Id)
}

////////////////////////////////////////////////////////////////////////
//
// Population Job

type taskExecutionTimeoutPopulationJob struct {
	job.Base `bson:"metadata" json:"metadata" yaml:"metadata"`
}

func makeTaskExecutionTimeoutPopulateJob() *taskExecutionTimeoutPopulationJob {
	j := &taskExecutionTimeoutPopulationJob{
		Base: job.Base{
			JobType: amboy.JobType{
				Name:    taskExecutionTimeoutPopulationJobName,
				Version: 0,
			},
		},
	}
	return j
}

func NewTaskExecutionMonitorPopulateJob(id string) amboy.Job {
	j := makeTaskExecutionTimeoutPopulateJob()
	j.SetID(fmt.Sprintf("%s.%s", j.Type().Name, id))
	return j
}

func (j *taskExecutionTimeoutPopulationJob) Run(ctx context.Context) {
	defer j.MarkComplete()

	flags, err := evergreen.GetServiceFlags()
	if err != nil {
		j.AddError(err)
		return
	}
	if flags.MonitorDisabled {
		grip.InfoWhen(sometimes.Percent(evergreen.DegradedLoggingPercent), message.Fields{
			"message":   "monitor is disabled",
			"operation": j.Type().Name,
			"impact":    "skipping task heartbeat cleanup job",
			"mode":      "degraded",
		})
		return
	}

	queue := evergreen.GetEnvironment().RemoteQueue()

	taskIDs := map[string]int{}
	tasks, err := host.FindStaleRunningTasks(heartbeatTimeoutThreshold, host.TaskHeartbeatPastCutoff)
	if err != nil {
		j.AddError(errors.Wrap(err, "error finding tasks with timed-out or stale heartbeats"))
		return
	}
	for _, t := range tasks {
		taskIDs[t.Id] = t.Execution
	}
	j.logTasks(tasks, "heartbeat past cutoff, on running host")
	tasks, err = host.FindStaleRunningTasks(heartbeatTimeoutThreshold, host.TaskNoHeartbeatSinceDispatch)
	if err != nil {
		j.AddError(errors.Wrap(err, "error finding tasks with timed-out or stale heartbeats"))
		return
	}
	for _, t := range tasks {
		taskIDs[t.Id] = t.Execution
	}
	j.logTasks(tasks, "no heartbeat since dispatch, on running host")
	tasks, err = host.FindStaleRunningTasks(heartbeatTimeoutThreshold, host.TaskUndispatchedHasHeartbeat)
	if err != nil {
		j.AddError(errors.Wrap(err, "error finding tasks with timed-out or stale heartbeats"))
		return
	}
	for _, t := range tasks {
		taskIDs[t.Id] = t.Execution
	}
	j.logTasks(tasks, "undispatched task has a heartbeat, on running host")

	tasks, err = task.Find(task.ByStaleRunningTask(heartbeatTimeoutThreshold, task.HeartbeatPastCutoff).WithFields(task.IdKey, task.ExecutionKey))
	if err != nil {
		j.AddError(errors.Wrap(err, "error finding tasks with timed-out or stale heartbeats"))
		return
	}
	for _, t := range tasks {
		taskIDs[t.Id] = t.Execution
	}
	j.logTasks(tasks, "heartbeat past cutoff")
	tasks, err = task.Find(task.ByStaleRunningTask(heartbeatTimeoutThreshold, task.NoHeartbeatSinceDispatch).WithFields(task.IdKey, task.ExecutionKey))
	if err != nil {
		j.AddError(errors.Wrap(err, "error finding tasks with timed-out or stale heartbeats"))
		return
	}
	for _, t := range tasks {
		taskIDs[t.Id] = t.Execution
	}
	j.logTasks(tasks, "no heartbeat since dispatch")

	for id, execution := range taskIDs {
		ts := utility.RoundPartOfHour(15)
		j.AddError(queue.Put(ctx, NewTaskExecutionMonitorJob(id, execution, 1, ts.Format(TSFormat))))
	}
	grip.Info(message.Fields{
		"operation": "task-execution-timeout-populate",
		"num_tasks": len(tasks),
		"errors":    j.HasErrors(),
	})
}

func (j *taskExecutionTimeoutPopulationJob) logTasks(tasks []task.Task, reason string) {
	taskIds := []string{}
	for _, t := range tasks {
		taskIds = append(taskIds, t.Id)
	}
	grip.Info(message.Fields{
		"message":   "found stale tasks",
		"reason":    reason,
		"tasks":     taskIds,
		"operation": j.Type().Name,
		"job_id":    j.ID(),
	})
}
