package units

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/model"
	"github.com/evergreen-ci/evergreen/model/event"
	"github.com/evergreen-ci/evergreen/repotracker"
	"github.com/evergreen-ci/evergreen/thirdparty"
	"github.com/evergreen-ci/utility"
	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/job"
	"github.com/mongodb/amboy/registry"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/pkg/errors"
)

const periodicBuildJobName = "periodic-build"

func init() {
	registry.AddJobType(periodicBuildJobName, func() amboy.Job {
		return makePeriodicBuildsJob()
	})
}

type periodicBuildJob struct {
	ProjectID    string `bson:"project_id"`
	DefinitionID string `bson:"def_id"`

	project *model.ProjectRef
	env     evergreen.Environment
	job.Base
}

func makePeriodicBuildsJob() *periodicBuildJob {
	j := &periodicBuildJob{
		Base: job.Base{
			JobType: amboy.JobType{
				Name:    periodicBuildJobName,
				Version: 0,
			},
		},
	}
	return j
}

func NewPeriodicBuildJob(projectID, definitionID string) amboy.Job {
	j := makePeriodicBuildsJob()
	j.ProjectID = projectID
	j.DefinitionID = definitionID
	ts := utility.RoundPartOfHour(15)
	j.SetID(fmt.Sprintf("%s-%s-%s-%s", periodicBuildJobName, projectID, definitionID, ts))
	j.SetScopes([]string{fmt.Sprintf("%s.%s.%s", periodicBuildJobName, projectID, definitionID)})
	j.SetEnqueueAllScopes(true)
	j.UpdateTimeInfo(amboy.JobTimeInfo{WaitUntil: ts})

	return j
}

func (j *periodicBuildJob) Run(ctx context.Context) {
	if j.env == nil {
		j.env = evergreen.GetEnvironment()
	}
	var err error
	j.project, err = model.FindMergedProjectRef(j.ProjectID, "", true)
	if err != nil {
		j.AddError(errors.Wrap(err, "error finding project"))
		return
	}
	var definition *model.PeriodicBuildDefinition
	for _, d := range j.project.PeriodicBuilds {
		if d.ID == j.DefinitionID {
			definition = &d
			break
		}
	}
	if definition == nil {
		j.AddError(errors.New("no definition ID found"))
		return
	}
	defer func() {
		baseTime := definition.NextRunTime
		if utility.IsZeroTime(baseTime) {
			baseTime = time.Now()
		}
		err = j.project.UpdateNextPeriodicBuild(definition.ID, baseTime.Add(time.Duration(definition.IntervalHours)*time.Hour))
		grip.Error(message.WrapError(err, message.Fields{
			"message":    "unable to set next periodic build job time",
			"project":    j.ProjectID,
			"definition": j.DefinitionID,
		}))
	}()
	versionID, versionError := j.addVersion(ctx, *definition)

	if versionError != nil {
		// if the version fails to be added, create a stub version and
		// log an event so users can get notified when notifications are configured
		metadata := model.VersionMetadata{
			IsAdHoc:         true,
			Message:         definition.Message,
			PeriodicBuildID: definition.ID,
			Alias:           definition.Alias,
		}
		stubVersion, dbErr := repotracker.ShellVersionFromRevision(ctx, j.project, metadata)
		if dbErr != nil {
			grip.Error(message.WrapError(dbErr, message.Fields{
				"message":            "error creating stub version for periodic build",
				"runner":             periodicBuildJobName,
				"project":            j.project,
				"project_identifier": j.project.Identifier,
				"definitionID":       j.DefinitionID,
			}))
		}
		if stubVersion == nil {
			j.AddError(versionError)
			return
		}
		stubVersion.Errors = []string{versionError.Error()}
		insertError := stubVersion.Insert()
		if err != nil {
			grip.Error(message.WrapError(insertError, message.Fields{
				"message":            "error inserting stub version for periodic build",
				"runner":             periodicBuildJobName,
				"project":            j.project,
				"project_identifier": j.project.Identifier,
				"definitionID":       j.DefinitionID,
			}))
		}
		event.LogVersionStateChangeEvent(stubVersion.Id, evergreen.VersionFailed)

		j.AddError(versionError)
		return
	}

	err = model.SetVersionActivation(versionID, true, evergreen.User)
	if err != nil {
		// if the version fails to activate, log an event so users
		// can get notified when notifications are configured
		event.LogVersionStateChangeEvent(versionID, evergreen.VersionFailed)
		j.AddError(err)
		return
	}

}

func (j *periodicBuildJob) addVersion(ctx context.Context, definition model.PeriodicBuildDefinition) (string, error) {
	token, err := j.env.Settings().GetGithubOauthToken()
	if err != nil {
		return "", errors.Wrap(err, "error getting github token")
	}
	configFile, err := thirdparty.GetGithubFile(ctx, token, j.project.Owner, j.project.Repo, definition.ConfigFile, j.project.Branch)
	if err != nil {
		return "", errors.Wrap(err, "error getting config file from github")
	}
	configBytes, err := base64.StdEncoding.DecodeString(*configFile.Content)
	if err != nil {
		return "", errors.Wrap(err, "error decoding config file")
	}
	proj := &model.Project{}
	opts := &model.GetProjectOpts{
		Ref:          j.project,
		Revision:     j.project.Branch,
		Token:        token,
		ReadFileFrom: model.ReadfromGithub,
	}
	intermediateProject, config, err := model.LoadProjectInto(ctx, configBytes, opts, j.project.Id, proj)
	if err != nil {
		return "", errors.Wrap(err, "error parsing config file")
	}
	metadata := model.VersionMetadata{
		IsAdHoc:         true,
		Message:         definition.Message,
		PeriodicBuildID: definition.ID,
		Alias:           definition.Alias,
	}
	projectInfo := &model.ProjectInfo{
		Ref:                 j.project,
		Project:             proj,
		IntermediateProject: intermediateProject,
		Config:              config,
	}
	v, err := repotracker.CreateVersionFromConfig(ctx, projectInfo, metadata, false, nil)
	if err != nil {
		return "", errors.Wrap(err, "error creating version from config")
	}
	if v == nil {
		return "", errors.New("no version created")
	}
	return v.Id, nil
}
