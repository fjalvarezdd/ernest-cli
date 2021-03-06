/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package command

import (
	"github.com/fatih/color"
	"github.com/urfave/cli"

	h "github.com/ernestio/ernest-cli/helper"
)

// NullWriter to disable logging
type NullWriter int

// Write sends to nowhere the log messages
func (NullWriter) Write([]byte) (int, error) { return 0, nil }

// MonitorEnv command
// Monitorizes an environment and shows the actions being performed on it
var MonitorEnv = cli.Command{
	Name:        "monitor",
	Aliases:     []string{"m"},
	Usage:       h.T("monitor.usage"),
	ArgsUsage:   h.T("monitor.args"),
	Description: h.T("monitor.description"),
	Action: func(c *cli.Context) error {
		m, cfg := setup(c)
		if cfg.Token == "" {
			color.Red("You're not allowed to perform this action, please log in")
			return nil
		}

		if len(c.Args()) == 0 {
			h.PrintError("You should specify an existing project name")
		}
		if len(c.Args()) == 1 {
			h.PrintError("You should specify an existing env name")
		}

		project := c.Args()[0]
		env := c.Args()[1]

		id, err := m.LatestBuildID(cfg.Token, project, env)
		if err != nil {
			h.PrintError(err.Error())
		}

		build, err := m.BuildStatus(cfg.Token, project, env, id)
		if err != nil {
			h.PrintError(err.Error())
		}

		if build.Status == "done" {
			color.Yellow("Environment has been successfully built")
			color.Yellow("You can check its information running `ernest-cli env info " + project + " / " + env + "`")
			return nil
		}

		return h.Monitorize(cfg.URL, "/events", cfg.Token, build.ID)
	},
}
