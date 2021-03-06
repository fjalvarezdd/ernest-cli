/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package command

import (
	"net/url"

	h "github.com/ernestio/ernest-cli/helper"
	"github.com/ernestio/ernest-cli/model"
	"github.com/fatih/color"
	"github.com/urfave/cli"
)

// Target command
// Configures the ernest target instance
var Target = cli.Command{
	Name:        "target",
	Aliases:     []string{"t"},
	Usage:       h.T("target.usage"),
	ArgsUsage:   h.T("target.args"),
	Description: h.T("target.description"),
	Action: func(c *cli.Context) error {
		if len(c.Args()) < 1 {
			h.PrintError("You should specify the target url")
		}
		_, cfg := setup(c)
		cfg.URL = c.Args()[0]
		if err := persistTarget(cfg); err != nil {
			h.PrintError("Couldn't write config file ~/.ernest check permissions")
		}
		color.Green("Target set")
		return nil
	},
}

func persistTarget(cfg *model.Config) error {
	u, _ := url.Parse(cfg.URL)
	if u.Scheme == "http" {
		color.Yellow("Warning! Your are using an insecure target for Ernest")
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		color.Red("You should specify a valid url for the target")
		return nil
	}
	err := model.SaveConfig(cfg)
	if err != nil {
		color.Red(err.Error())
		return err
	}
	return nil
}
