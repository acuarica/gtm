// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"flag"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strings"

	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/report"
	"github.com/git-time-metric/gtm/scm"
	"github.com/mitchellh/cli"
)

// WebCmd ...
type WebCmd struct {
	UI cli.Ui
}

// NewWeb returns a new WebCmd struct
func NewWeb() (cli.Command, error) {
	return WebCmd{}, nil
}

// Help returns help for the web command
func (c WebCmd) Help() string {
	helpText := `
Usage: gtm web [options]

  Starts local web server for reporting and status.

Options:

  -port=8090                 Default port
`
	return strings.TrimSpace(helpText)
}

// Run executes clean command with args
func (c WebCmd) Run(args []string) int {
	var port int
	cmdFlags := flag.NewFlagSet("web", flag.ContinueOnError)
	cmdFlags.IntVar(&port, "port", 8080, "")
	cmdFlags.Usage = func() { c.UI.Output(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	c.UI.Output(fmt.Sprintf("Starting local web server at port %d ... ", port))

	http.HandleFunc("/", c.ViewHandler)
	http.HandleFunc("/data", c.DataHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	return 0
}

// ViewHandler creates the index handler.
func (c WebCmd) ViewHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("web/index.html")
	t.Execute(w, nil)
}

// DataHandler returns the commit data for all projects.
func (c WebCmd) DataHandler(w http.ResponseWriter, r *http.Request) {
	index, err := project.NewIndex()
	if err != nil {
		c.UI.Error(err.Error())
	}

	tagList := []string{}
	projects, err := index.Get(tagList, true)
	if err != nil {
		c.UI.Error(err.Error())
	}

	limiter, err := scm.NewCommitLimiter(math.MaxUint32,
		"", "", "", "",
		false, false, false, false,
		false, false, false, false)
	if err != nil {
		c.UI.Error(err.Error())
	}

	projCommits := []report.ProjectCommits{}
	for _, p := range projects {
		commits, err := scm.CommitIDs(limiter, p)
		if err != nil {
			c.UI.Error(err.Error())
		}
		fmt.Println("%s", p)
		projCommits = append(projCommits, report.ProjectCommits{Path: p, Commits: commits})
	}

	for _, pc := range projCommits {
		fmt.Println("%d", len(pc.Commits))
	}

	js, err := report.GetProjects(projCommits)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// Synopsis return help for web command
func (c WebCmd) Synopsis() string {
	return "Starts local web server for reporting and status"
}
