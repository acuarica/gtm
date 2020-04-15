package report

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"path/filepath"

	"github.com/git-time-metric/gtm/metric"
	"github.com/git-time-metric/gtm/note"
	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/scm"
	"github.com/git-time-metric/gtm/util"
)

// GetProjectTotals returns the project summary report
func GetProjectTotals(r *http.Request) ([]byte, error) {
	index, err := project.NewIndex()
	if err != nil {
		return nil, err
	}

	tagList := []string{}
	projectList, err := index.Get(tagList, true)
	if err != nil {
		return nil, err
	}

	limiter, err := scm.NewCommitLimiter(math.MaxUint32,
		"", "", "", "",
		false, true, false, false,
		false, false, false, false)
	if err != nil {
		return nil, err
	}

	projCommits := []ProjectCommits{}
	for _, p := range projectList {
		commits, err := scm.CommitIDs(limiter, p)
		if err != nil {
			return nil, err
		}
		fmt.Println("%s", p)
		projCommits = append(projCommits, ProjectCommits{Path: p, Commits: commits})
	}

	for _, pc := range projCommits {
		fmt.Println("%d", len(pc.Commits))
	}

	notes := retrieveNotes(projCommits, false, false, false, "Mon Jan 02")
	projectTotals := map[string]int{}
	projects := map[string]struct {
		Total int
		Label string
	}{}

	for _, n := range notes {
		projectTotals[n.Project] += n.Note.Total()
	}

	for p, total := range projectTotals {
		projects[p] = struct {
			Total int
			Label string
		}{total, util.FormatDuration(total)}
	}

	return json.Marshal(projects)
}

// GetStatusTotals returns the status data.
func GetStatusTotals(r *http.Request) ([]byte, error) {
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
		Total int
		Label string
	}{}
	for _, projPath := range projects {
		if commitNote, err = metric.Process(true, projPath); err != nil {
			return nil, err
		}
		total := commitNote.Total()
		projName := filepath.Base(projPath)
		projectTotals[projName] = struct {
			Total int
			Label string
		}{total, util.FormatDuration(total)}
	}

	return json.Marshal(projectTotals)
}
