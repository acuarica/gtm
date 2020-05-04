// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/git-time-metric/gtm/report"
	"github.com/git-time-metric/gtm/util"
	"github.com/mitchellh/cli"
)

// ExportCmd ...
type ExportCmd struct {
	UI cli.Ui
}

// NewExport returns a new WebCmd struct
func NewExport() (cli.Command, error) {
	return ExportCmd{}, nil
}

// Help returns help for the export command
func (c ExportCmd) Help() string {
	helpText := `
Usage: gtm export [options]

  Export time data in json format to be further processed by other applications.

Options:

  -data=commits              Specify time data to be exported [commits|status|projects] (default status)

  Commit Limiting:

  -from-date=yyyy-mm-dd      Show commits starting from this date
  -to-date=yyyy-mm-dd        Show commits thru the end of this date
  -search=""                 Show commits which contain either author or message substring
`
	return strings.TrimSpace(helpText)
}

// Run executes clean command with args
func (c ExportCmd) Run(args []string) int {
	var data, fromDate, toDate, search string
	cmdFlags := flag.NewFlagSet("export", flag.ContinueOnError)
	cmdFlags.StringVar(&data, "data", "status", "")
	cmdFlags.StringVar(&fromDate, "from-date", "", "")
	cmdFlags.StringVar(&toDate, "to-date", "", "")
	cmdFlags.StringVar(&search, "search", "", "")
	cmdFlags.Usage = func() { c.UI.Output(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if !util.StringInSlice([]string{"commits", "status", "projects"}, data) {
		c.UI.Error(fmt.Sprintf("export -data=%s is not a valid option\n", data))
		return 1
	}

	var (
		out []byte
		err error
	)

	switch data {
	case "commits":
		out, err = report.GetCommitNotes(fromDate, toDate)
	case "status":
		out, err = report.GetStatusNotes()
	case "projects":
		out, err = report.GetProjectList()
	}

	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	fmt.Println(string(out))

	return 0
}

// Synopsis return help for web command
func (c ExportCmd) Synopsis() string {

	return "Export time data to be further processed"
}
