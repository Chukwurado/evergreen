package model

import (
	"net/http"

	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/model/host"
	"github.com/evergreen-ci/evergreen/model/task"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/pkg/errors"
)

type (
	apiTaskKey    int
	apiHostKey    int
	apiProjectKey int
)

const (
	ApiTaskKey    apiTaskKey    = 0
	ApiHostKey    apiHostKey    = 0
	ApiProjectKey apiProjectKey = 0
)

// ValidateTask ensures that a task ID is set and corresponds to a task in the
// database. If checkSecret is true, it also validates the task's secret. It
// returns a task, http status code, and error.
func ValidateTask(taskId string, checkSecret bool, r *http.Request) (*task.Task, int, error) {
	if taskId == "" {
		return nil, http.StatusBadRequest, errors.New("missing task id")
	}
	t, err := task.FindOneId(taskId)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if t == nil {
		return nil, http.StatusNotFound, errors.New("task not found")
	}
	if checkSecret {
		secret := r.Header.Get(evergreen.TaskSecretHeader)
		if secret != t.Secret {
			return nil, http.StatusConflict, errors.Errorf("Wrong secret sent for task '%s'", taskId)
		}
	}
	return t, http.StatusOK, nil
}

// ValidateHost ensures that the host exists in the database and that, if a
// secret is provided, it matches the secret in the database. If a task ID is
// provided, it ensures that the host should be running this task. It returns a
// host, http status code, and error.
func ValidateHost(hostId string, r *http.Request) (*host.Host, int, error) {
	if hostId == "" {
		// fall back to the host header if host ids are not part of the path
		hostId = r.Header.Get(evergreen.HostHeader)
		if hostId == "" {
			return nil, http.StatusBadRequest, errors.Errorf("Request %s is missing host information", r.URL)
		}
	}
	secret := r.Header.Get(evergreen.HostSecretHeader)
	if secret == "" {
		return nil, http.StatusBadRequest, errors.Errorf("Missing host secret for host '%s'", hostId)
	}

	// If the host was provisioned through user data, the host will be started
	// with the intent host ID instead of the _id.
	h, err := host.FindOneByIdOrTag(hostId)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "finding host '%s'", hostId)
	}
	if h == nil {
		return nil, http.StatusNotFound, errors.Errorf("host '%s' not found", hostId)
	}
	if secret != h.Secret {
		return nil, http.StatusUnauthorized, errors.Errorf("Invalid host secret for host '%s'", hostId)
	}

	// if the task is attached to the context, check host-task relationship
	var t *task.Task
	if rv := r.Context().Value(ApiTaskKey); rv != nil {
		if rvTask, ok := rv.(*task.Task); ok {
			t = rvTask
		}
	}
	if badHostTaskRelationship(h, t) {
		return nil, http.StatusConflict, errors.Errorf("Host '%s' should be running '%s', not '%s'", hostId, h.RunningTask, t.Id)
	}
	return h, http.StatusOK, nil
}

func badHostTaskRelationship(h *host.Host, t *task.Task) bool {
	if t == nil {
		return false
	}
	if t.Id == h.RunningTask {
		return false
	}
	if t.Id == h.LastTask {
		if h.RunningTask == "" {
			return false
		}
		nextTask, err := task.FindOneId(h.RunningTask)
		if err != nil {
			grip.Error(message.WrapError(err, message.Fields{
				"message": "problem finding task",
				"task":    h.RunningTask,
			}))
		}
		// If the next task has not been marked started, allow logs to be posted for post group.
		if nextTask == nil || nextTask.Status == evergreen.TaskDispatched {
			return false
		}
	}
	return true
}
