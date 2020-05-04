package report

import (
	"encoding/json"
	"math"
	"path/filepath"

	"github.com/git-time-metric/gtm/metric"
	"github.com/git-time-metric/gtm/note"
	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/scm"
	"github.com/git-time-metric/gtm/util"
)

// GetCommitNotes gets
func GetCommitNotes(fromDate string, toDate string) ([]byte, error) {
	index, err := project.NewIndex()
	if err != nil {
		return nil, err
	}

	tagList := []string{}
	projectList, err := index.Get(tagList, true)
	if err != nil {
		return nil, err
	}

	n := math.MaxUint32
	limiter, err := scm.NewCommitLimiter(n,
		fromDate, toDate, "", "",
		false, false, false, false, false, false, false, false)
	if err != nil {
		return nil, err
	}

	projCommits := []ProjectCommits{}
	for _, p := range projectList {
		commits, err := scm.CommitIDs(limiter, p)
		if err != nil {
			return nil, err
		}
		projCommits = append(projCommits, ProjectCommits{Path: p, Commits: commits})
	}

	notes := retrieveNotes(projCommits, false, false, false, "Mon Jan 02")

	return json.Marshal(notes)
}

// GetStatusNotes returns the status data.
func GetStatusNotes() ([]byte, error) {
	var (
		err        error
		commitNote note.CommitNote
	)

	index, err := project.NewIndex()
	if err != nil {
		return nil, err
	}

	tagList := []string{}
	projects, err := index.Get(tagList, true)
	if err != nil {
		return nil, err
	}

	projectTotals := map[string]struct {
		Total      int
		Label      string
		CommitNote note.CommitNote
	}{}
	for _, projPath := range projects {
		if commitNote, err = metric.Process(true, projPath); err != nil {
			return nil, err
		}
		total := commitNote.Total()
		projName := filepath.Base(projPath)
		projectTotals[projName] = struct {
			Total      int
			Label      string
			CommitNote note.CommitNote
		}{total, util.FormatDuration(total), commitNote}
	}

	return json.Marshal(projectTotals)
}

// GetProjectList returns the list of all available projects.
func GetProjectList() ([]byte, error) {
	index, err := project.NewIndex()
	if err != nil {
		return nil, err
	}

	tagList := []string{}
	projectList, err := index.Get(tagList, true)
	if err != nil {
		return nil, err
	}
	return json.Marshal(projectList)
}
