// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/git-time-metric/gtm/report"
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

	http.HandleFunc("/", c.viewHandler)
	http.HandleFunc("/data/commits", c.createDataHandler(report.GetCommitNotes))
	http.HandleFunc("/data/projects/totals", c.createDataHandler(report.GetProjectTotals))
	http.HandleFunc("/data/timeline", c.createDataHandler(report.GetTimeline))
	http.HandleFunc("/data/status/totals", c.createDataHandler(report.GetStatusTotals))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	return 0
}

func (c WebCmd) viewHandler(w http.ResponseWriter, r *http.Request) {
	// FIXME: Working dir needs to be in project root
	t, _ := template.ParseFiles("web/index.html")
	t.Execute(w, nil)
}

func (c WebCmd) createDataHandler(f func(r *http.Request) ([]byte, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		js, err := f(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

// Synopsis return help for web command
func (c WebCmd) Synopsis() string {
	return "Starts local web server for reporting and status"
}
