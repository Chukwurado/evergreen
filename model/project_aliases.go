package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/db"
	mgobson "github.com/evergreen-ci/evergreen/db/mgo/bson"
	"github.com/evergreen-ci/utility"
	"github.com/mongodb/anser/bsonutil"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	idKey          = bsonutil.MustHaveTag(ProjectAlias{}, "ID")
	projectIDKey   = bsonutil.MustHaveTag(ProjectAlias{}, "ProjectID")
	aliasKey       = bsonutil.MustHaveTag(ProjectAlias{}, "Alias")
	gitTagKey      = bsonutil.MustHaveTag(ProjectAlias{}, "GitTag")
	remotePathKey  = bsonutil.MustHaveTag(ProjectAlias{}, "RemotePath")
	variantKey     = bsonutil.MustHaveTag(ProjectAlias{}, "Variant")
	taskKey        = bsonutil.MustHaveTag(ProjectAlias{}, "Task")
	variantTagsKey = bsonutil.MustHaveTag(ProjectAlias{}, "VariantTags")
	taskTagsKey    = bsonutil.MustHaveTag(ProjectAlias{}, "TaskTags")
)

const (
	ProjectAliasCollection = "project_aliases"
)

// ProjectAlias defines a single alias mapping an alias name to two regexes which
// define the variants and tasks for the alias. Users can use these aliases for
// operations within the system.
//
// For example, a user can specify that alias with the CLI tool so that a project
// admin can define a set of default builders for patch builds. Pull request
// testing uses a special alias, "__github" to determine the default
// variants and tasks to run in a patch build.
//
// An alias can be specified multiple times. The resulting variant/task
// combinations are the union of the aliases. For example, a user might set the
// following:
//
// ALIAS                  VARIANTS          TASKS
// __github               .*linux.*         .*test.*
// __github               ^ubuntu1604.*$    ^compile.*$
//
// This will cause a GitHub pull request to create and finalize a patch which runs
// all tasks containing the string “test” on all variants containing the string
// “linux”; and to run all tasks beginning with the string “compile” to run on all
// variants beginning with the string “ubuntu1604”.

// Git tags use a special alias "__git_tag" and create a new version for the matching
// variants/tasks, assuming the tag matches the defined git_tag regex.
// In this way, users can define different behavior for different kind of tags.
type ProjectAlias struct {
	ID          mgobson.ObjectId `bson:"_id,omitempty" json:"_id" yaml:"id"`
	ProjectID   string           `bson:"project_id" json:"project_id" yaml:"project_id"`
	Alias       string           `bson:"alias" json:"alias" yaml:"alias"`
	Variant     string           `bson:"variant,omitempty" json:"variant" yaml:"variant"`
	GitTag      string           `bson:"git_tag" json:"git_tag" yaml:"git_tag"`
	RemotePath  string           `bson:"remote_path" json:"remote_path" yaml:"remote_path"`
	VariantTags []string         `bson:"variant_tags,omitempty" json:"variant_tags" yaml:"variant_tags"`
	Task        string           `bson:"task,omitempty" json:"task" yaml:"task"`
	TaskTags    []string         `bson:"tags,omitempty" json:"tags" yaml:"task_tags"`
}

type ProjectAliases []ProjectAlias

// FindAliasesForProjectFromDb fetches all aliases for a given project without merging with aliases from the parser project
func FindAliasesForProjectFromDb(projectID string) ([]ProjectAlias, error) {
	var out []ProjectAlias
	q := db.Query(bson.M{
		projectIDKey: projectID,
	})
	err := db.FindAllQ(ProjectAliasCollection, q, &out)
	if err != nil {
		return nil, errors.Wrap(err, "error finding project aliases")
	}
	return out, nil
}

// FindAliasesForRepo fetches all aliases for a given project
func FindAliasesForRepo(repoId string) ([]ProjectAlias, error) {
	out := []ProjectAlias{}
	q := db.Query(bson.M{
		projectIDKey: repoId,
	})
	err := db.FindAllQ(ProjectAliasCollection, q, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// findMatchingAliasForRepo finds all aliases with a given name for a repo.
// Typically FindAliasInProjectOrRepo should be used.
func findMatchingAliasForRepo(repoID, alias string) ([]ProjectAlias, error) {
	var out []ProjectAlias
	q := db.Query(bson.M{
		projectIDKey: repoID,
		aliasKey:     alias,
	})
	err := db.FindAllQ(ProjectAliasCollection, q, &out)
	if err != nil {
		return nil, errors.Wrap(err, "error finding project aliases")
	}
	return out, nil
}

// findMatchingAliasForProjectRef finds all aliases with a given name for a project.
// Typically FindAliasInProjectOrRepo should be used.
// Returns true if we have an alias match or the alias doesn't match but
// other aliases in the category are defined, in which case we shouldn't check other sources.
func findMatchingAliasForProjectRef(projectID, alias string) ([]ProjectAlias, bool, error) {
	var out []ProjectAlias
	q := db.Query(bson.M{
		projectIDKey: projectID,
		aliasKey:     alias,
	})
	err := db.FindAllQ(ProjectAliasCollection, q, &out)
	if err != nil {
		return nil, false, errors.Wrap(err, "error finding project aliases")
	}

	if len(out) == 0 && IsPatchAlias(alias) {
		// return true if any patch aliases are defined
		numPatchAliases, err := countPatchAliases(projectID)
		if err != nil {
			return nil, false, errors.Wrap(err, "error counting patch aliases")
		}
		return nil, numPatchAliases > 0, nil
	}
	return out, len(out) > 0, nil
}

// findMatchingAliasForProjectConfig finds any aliases matching the alias input in the project config.
func findMatchingAliasForProjectConfig(projectID, alias string) ([]ProjectAlias, error) {
	projectConfig, err := FindProjectConfigToMerge(projectID, "")
	if err != nil {
		return nil, errors.Wrap(err, "error finding project config")
	}
	if projectConfig == nil {
		return nil, nil
	}

	projectConfigAliases := aliasesToMap(getFullProjectConfigAliases(projectConfig))
	return projectConfigAliases[alias], nil
}

func countPatchAliases(projectID string) (int, error) {
	return db.Count(ProjectAliasCollection, bson.M{
		projectIDKey: projectID,
		aliasKey:     bson.M{"$nin": evergreen.InternalAliases},
	})
}

func aliasesToMap(aliases []ProjectAlias) map[string][]ProjectAlias {
	output := make(map[string][]ProjectAlias)
	for _, alias := range aliases {
		output[alias.Alias] = append(output[alias.Alias], alias)
	}
	return output
}

func getFullProjectConfigAliases(projectConfig *ProjectConfig) []ProjectAlias {
	var projectConfigAliases []ProjectAlias
	if projectConfig != nil {
		for _, commitQueueAlias := range projectConfig.CommitQueueAliases {
			commitQueueAlias.Alias = evergreen.CommitQueueAlias
			projectConfigAliases = append(projectConfigAliases, commitQueueAlias)
		}
		for _, gitTagAlias := range projectConfig.GitTagAliases {
			gitTagAlias.Alias = evergreen.GitTagAlias
			projectConfigAliases = append(projectConfigAliases, gitTagAlias)
		}
		for _, githubPrAlias := range projectConfig.GitHubPRAliases {
			githubPrAlias.Alias = evergreen.GithubPRAlias
			projectConfigAliases = append(projectConfigAliases, githubPrAlias)
		}
		for _, gitHubCheckAlias := range projectConfig.GitHubChecksAliases {
			gitHubCheckAlias.Alias = evergreen.GithubChecksAlias
			projectConfigAliases = append(projectConfigAliases, gitHubCheckAlias)
		}
		for _, patchAlias := range projectConfig.PatchAliases {
			projectConfigAliases = append(projectConfigAliases, patchAlias)
		}
	}
	return projectConfigAliases
}

// FindAliasInProjectOrRepo finds all aliases with a given name for a project.
// If the project has no aliases, the repo is checked for aliases.
func FindAliasInProjectOrRepo(projectID, alias string) ([]ProjectAlias, error) {
	aliases, shouldExit, err := FindAliasInProjectOrRepoFromDb(projectID, alias)
	if err != nil {
		return nil, errors.Wrap(err, "error checking for existing aliases")
	}
	// If nothing is defined in the DB, check the project config,
	// unless the alias defined is a patch alias and there are patch aliases
	// defined on the project page.
	if len(aliases) > 0 || shouldExit {
		return aliases, nil
	}
	return findMatchingAliasForProjectConfig(projectID, alias)
}

// FindAliasInProjectOrRepoFromDb finds all aliases with a given name for a project without merging with parser project.
// If the project has no aliases, the repo is checked for aliases.
func FindAliasInProjectOrRepoFromDb(projectID, alias string) ([]ProjectAlias, bool, error) {
	aliases, shouldExit, err := findMatchingAliasForProjectRef(projectID, alias)
	if err != nil {
		return aliases, false, errors.Wrapf(err, "error finding aliases for project ref '%s'", projectID)
	}
	if shouldExit {
		return aliases, true, nil
	}
	return tryGetRepoAliases(projectID, alias, aliases)
}

func tryGetRepoAliases(projectID string, alias string, aliases []ProjectAlias) ([]ProjectAlias, bool, error) {
	project, err := FindBranchProjectRef(projectID)
	if err != nil {
		return aliases, false, errors.Wrapf(err, "error finding project '%s'", projectID)
	}
	if project == nil {
		return aliases, false, errors.Errorf("project '%s' does not exist", projectID)
	}
	if !project.UseRepoSettings() {
		return aliases, false, nil
	}

	aliases, err = findMatchingAliasForRepo(project.RepoRefId, alias)
	if err != nil {
		return aliases, false, errors.Wrapf(err, "error finding aliases for repo '%s'", project.RepoRefId)
	}
	shouldExit := false
	if IsPatchAlias(alias) {
		numRepoPatchAliases, err := countPatchAliases(project.RepoRefId)
		if err != nil {
			return nil, false, errors.Wrap(err, "error counting patch aliases")
		}
		shouldExit = numRepoPatchAliases > 0
	}
	return aliases, shouldExit, nil
}

func FindMatchingGitTagAliasesInProject(projectID, tag string) ([]ProjectAlias, error) {
	aliases, err := FindAliasInProjectOrRepo(projectID, evergreen.GitTagAlias)
	if err != nil {
		return nil, err
	}
	matchingAliases, err := aliasesMatchingGitTag(aliases, tag)
	if err != nil {
		return nil, err
	}
	for _, alias := range matchingAliases {
		if alias.RemotePath != "" && len(matchingAliases) > 1 {
			return matchingAliases, errors.Errorf("git tag '%s' matches multiple aliases but a remote path is defined", tag)
		}
	}
	return matchingAliases, nil
}

// IsValidId returns whether the supplied Id is a valid patch doc id (BSON ObjectId).
func IsValidId(id string) bool {
	return mgobson.IsObjectIdHex(id)
}

// NewId constructs a valid patch Id from the given hex string.
func NewId(id string) mgobson.ObjectId { return mgobson.ObjectIdHex(id) }

func (p *ProjectAlias) Upsert() error {
	if len(p.ProjectID) == 0 {
		return errors.New("empty project ID")
	}
	if p.ID.Hex() == "" {
		p.ID = mgobson.NewObjectId()
	}
	update := bson.M{
		aliasKey:       p.Alias,
		gitTagKey:      p.GitTag,
		remotePathKey:  p.RemotePath,
		projectIDKey:   p.ProjectID,
		variantKey:     p.Variant,
		variantTagsKey: p.VariantTags,
		taskTagsKey:    p.TaskTags,
		taskKey:        p.Task,
	}

	_, err := db.Upsert(ProjectAliasCollection, bson.M{
		idKey: p.ID,
	}, bson.M{"$set": update})
	if err != nil {
		return errors.Wrapf(err, "failed to insert project alias '%s'", p.ID)
	}
	return nil
}

func UpsertAliasesForProject(aliases []ProjectAlias, projectId string) error {
	catcher := grip.NewBasicCatcher()
	for i := range aliases {
		if aliases[i].ProjectID != projectId { // new project, so we need a new document (new ID)
			aliases[i].ProjectID = projectId
			aliases[i].ID = ""
		}
		catcher.Add(aliases[i].Upsert())
	}
	return catcher.Resolve()
}

// RemoveProjectAlias removes a project alias with the given document ID from the
// database.
func RemoveProjectAlias(id string) error {
	if id == "" {
		return errors.New("can't remove project alias with empty id")
	}
	err := db.Remove(ProjectAliasCollection, bson.M{idKey: mgobson.ObjectIdHex(id)})
	if err != nil {
		return errors.Wrapf(err, "failed to remove project alias %s", id)
	}
	return nil
}

func IsPatchAlias(alias string) bool {
	return !utility.StringSliceContains(evergreen.InternalAliases, alias)
}

func (a ProjectAliases) HasMatchingGitTag(tag string) (bool, error) {
	matchingAliases, err := aliasesMatchingGitTag(a, tag)
	if err != nil {
		return false, err
	}
	return len(matchingAliases) > 0, nil
}

func aliasesMatchingGitTag(a ProjectAliases, tag string) (ProjectAliases, error) {
	res := []ProjectAlias{}
	for _, alias := range a {
		gitTagRegex, err := regexp.Compile(alias.GitTag)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to compile regex %s", gitTagRegex)
		}
		if isValidRegexOrTag(tag, alias.GitTag, nil, nil, gitTagRegex) {
			res = append(res, alias)
		}
	}
	return res, nil
}

func (a ProjectAliases) AliasesMatchingVariant(variant string, variantTags []string) (ProjectAliases, error) {
	res := []ProjectAlias{}
	for _, alias := range a {
		variantRegex, err := regexp.Compile(alias.Variant)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to compile regex %s", variantRegex)
		}
		if isValidRegexOrTag(variant, alias.Variant, variantTags, alias.VariantTags, variantRegex) {
			res = append(res, alias)
		}
	}
	return res, nil
}

// HasMatchingTask assumes that the aliases given already match the preferred variant.
func (a ProjectAliases) HasMatchingTask(taskName string, taskTags []string) (bool, error) {
	for _, alias := range a {
		taskRegex, err := regexp.Compile(alias.Task)
		if err != nil {
			return false, errors.Wrapf(err, "unable to compile regex %s", taskRegex)
		}
		if isValidRegexOrTag(taskName, alias.Task, taskTags, alias.TaskTags, taskRegex) {
			return true, nil
		}
	}
	return false, nil
}

func isValidRegexOrTag(curItem, aliasRegex string, curTags, aliasTags []string, regexp *regexp.Regexp) bool {
	isValidRegex := aliasRegex != "" && regexp.MatchString(curItem)
	isValidTag := false
	for _, tag := range aliasTags {
		if utility.StringSliceContains(curTags, tag) {
			isValidTag = true
			break
		}
		// a negated tag
		if len(tag) > 0 && tag[0] == '!' && !utility.StringSliceContains(curTags, tag[1:]) {
			isValidTag = true
			break
		}
	}

	return isValidRegex || isValidTag
}

func ValidateProjectAliases(aliases []ProjectAlias, aliasType string) []string {
	errs := []string{}
	for i, pd := range aliases {
		if strings.TrimSpace(pd.Alias) == "" {
			errs = append(errs, fmt.Sprintf("%s: alias name #%d can't be empty string", aliasType, i+1))
		}
		if pd.Alias == evergreen.GitTagAlias {
			errs = append(errs, validateGitTagAlias(pd, aliasType, i+1)...)
			continue
		}
		if strings.TrimSpace(pd.GitTag) != "" || strings.TrimSpace(pd.RemotePath) != "" {
			errs = append(errs, fmt.Sprintf("%s: cannot define git tag or remote path on line #%d", aliasType, i+1))
		}
		errs = append(errs, validateAliasPatchDefinition(pd, aliasType, i+1)...)
	}

	return errs
}

func validateAliasPatchDefinition(pd ProjectAlias, aliasType string, lineNum int) []string {
	errs := []string{}
	if (strings.TrimSpace(pd.Variant) == "") == (len(pd.VariantTags) == 0) {
		errs = append(errs, fmt.Sprintf("%s: must specify exactly one of variant regex or variant tags on line #%d", aliasType, lineNum))
	}
	if (strings.TrimSpace(pd.Task) == "") == (len(pd.TaskTags) == 0) {
		errs = append(errs, fmt.Sprintf("%s: must specify exactly one of task regex or task tags on line #%d", aliasType, lineNum))
	}

	if _, err := regexp.Compile(pd.Variant); err != nil {
		errs = append(errs, fmt.Sprintf("%s: variant regex #%d is invalid", aliasType, lineNum))
	}
	if _, err := regexp.Compile(pd.Task); err != nil {
		errs = append(errs, fmt.Sprintf("%s: task regex #%d is invalid", aliasType, lineNum))
	}
	return errs
}

func validateGitTagAlias(pd ProjectAlias, aliasType string, lineNum int) []string {
	errs := []string{}
	if strings.TrimSpace(pd.GitTag) == "" {
		errs = append(errs, fmt.Sprintf("%s: must define valid git tag regex on line #%d", aliasType, lineNum))
	}
	if _, err := regexp.Compile(pd.GitTag); err != nil {
		errs = append(errs, fmt.Sprintf("%s: git tag regex #%d is invalid", aliasType, lineNum))
	}
	// if path is defined then no patch definition can be given
	if strings.TrimSpace(pd.RemotePath) != "" && populatedPatchDefinition(pd) {
		errs = append(errs, fmt.Sprintf("%s: cannot define remote path and task/variant constraints on line #%d", aliasType, lineNum))
	}
	if strings.TrimSpace(pd.RemotePath) == "" {
		errs = append(errs, validateAliasPatchDefinition(pd, aliasType, lineNum)...)
	}
	return errs
}

func populatedPatchDefinition(pd ProjectAlias) bool {
	return strings.TrimSpace(pd.Variant) != "" || strings.TrimSpace(pd.Task) != "" ||
		len(pd.VariantTags) != 0 || len(pd.Task) != 0
}
