package report

import "encoding/json"

// GetProjects returns the project summary report
func GetProjects(projects []ProjectCommits) ([]byte, error) {
	notes := retrieveNotes(projects, false, false, false, "Mon Jan 02")
	projectTotals := map[string]int{}
	for _, n := range notes {
		projectTotals[n.Project] += n.Note.Total()
	}

	return json.Marshal(projectTotals)
}
