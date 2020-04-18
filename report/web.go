package report

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/git-time-metric/gtm/metric"
	"github.com/git-time-metric/gtm/note"
	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/scm"
	"github.com/git-time-metric/gtm/util"
)

// GetTimeline gets
func GetCommitNotes(r *http.Request) ([]byte, error) {
	notes, err := getCommitNotes(r)
	if err != nil {
		return nil, err
	}

	return json.Marshal(notes)
}

// GetProjectTotals returns the project summary report
func GetProjectTotals(r *http.Request) ([]byte, error) {
	notes, err := getCommitNotes(r)
	if err != nil {
		return nil, err
	}

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

// GetTimeline gets
func GetTimeline(r *http.Request) ([]byte, error) {
	notes, err := getCommitNotes(r)
	if err != nil {
		return nil, err
	}

	timeline, err := notes.timeline()
	if err != nil {
		return nil, err
	}

	timelinef := []struct {
		Day      string
		Seconds  int
		Hours    [24]int
		Duration string
	}{}
	for _, entry := range timeline {
		entryf := struct {
			Day      string
			Seconds  int
			Hours    [24]int
			Duration string
		}{entry.Day, entry.Seconds, entry.Hours, entry.Duration()}
		timelinef = append(timelinef, entryf)
	}

	return json.Marshal(timelinef)
}

func getCommitNotes(r *http.Request) (commitNoteDetails, error) {
	hasArg := func(arg string) bool {
		_, ok := r.URL.Query()[arg]
		return ok
	}
	getInt := func(arg string, def int) int {
		if hasArg(arg) {
			n, err := strconv.Atoi(r.URL.Query()[arg][0])
			if err != nil {
				return def
			}
			return n
		}
		return def
	}

	all := hasArg("all")
	today := hasArg("today")
	yesterday := hasArg("yesterday")
	thisWeek := hasArg("thisweek")
	lastWeek := hasArg("lastweek")
	thisMonth := hasArg("thismonth")
	lastMonth := hasArg("lastmonth")
	thisYear := hasArg("thisyear")
	lastYear := hasArg("lastyear")
	n := getInt("n", math.MaxUint32)

	index, err := project.NewIndex()
	if err != nil {
		return nil, err
	}

	tagList := []string{}
	projectList, err := index.Get(tagList, all)
	if err != nil {
		return nil, err
	}

	limiter, err := scm.NewCommitLimiter(n,
		"", "", "", "",
		today, yesterday, thisWeek, lastWeek,
		thisMonth, lastMonth, thisYear, lastYear)
	if err != nil {
		return nil, err
	}
	fmt.Println("%s", limiter)

	projCommits := []ProjectCommits{}
	for _, p := range projectList {
		commits, err := scm.CommitIDs(limiter, p)
		if err != nil {
			return nil, err
		}
		// fmt.Println("%s", p)
		projCommits = append(projCommits, ProjectCommits{Path: p, Commits: commits})
	}

	// for _, pc := range projCommits {
	// 	fmt.Println("%d", len(pc.Commits))
	// }

	notes := retrieveNotes(projCommits, false, false, false, "Mon Jan 02")
	return notes, nil
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
