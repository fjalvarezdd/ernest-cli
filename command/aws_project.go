/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package command

// CmdProject subcommand
import (
	"fmt"

	h "github.com/ernestio/ernest-cli/helper"
	"github.com/ernestio/ernest-cli/model"
	"github.com/fatih/color"
	"github.com/urfave/cli"
)

// CreateAWSProject : Creates an AWS project
var CreateAWSProject = cli.Command{
	Name:        "aws",
	Usage:       h.T("aws.project.create.usage"),
	Description: h.T("aws.project.create.description"),
	ArgsUsage:   h.T("aws.project.create.args"),
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "region, r",
			Value: "",
			Usage: "Project region",
		},
		cli.StringFlag{
			Name:  "access_key_id, k",
			Value: "",
			Usage: "AWS access key id",
		},
		cli.StringFlag{
			Name:  "secret_access_key, s",
			Value: "",
			Usage: "AWS Secret access key",
		},
		cli.StringFlag{
			Name:  "template, t",
			Value: "",
			Usage: "Project template",
		},
		cli.BoolFlag{
			Name:  "fake, f",
			Usage: "Fake project",
		},
	},
	Action: func(c *cli.Context) error {
		var errs []string
		var accessKeyID, secretAccessKey, region string
		var fake bool
		m, cfg := setup(c)

		if len(c.Args()) < 1 {
			msg := "You should specify the project name"
			color.Red(msg)
			return nil
		}

		if cfg.Token == "" {
			color.Red("You're not allowed to perform this action, please log in")
			return nil
		}
		name := c.Args()[0]

		template := c.String("template")
		if template != "" {
			var t model.ProjectTemplate
			if err := getProjectTemplate(template, &t); err != nil {
				color.Red(err.Error())
				return nil
			}
			accessKeyID = t.Token
			secretAccessKey = t.Secret
			region = t.Region
			fake = t.Fake
		}
		if c.String("secret_access_key") != "" {
			secretAccessKey = c.String("secret_access_key")
		}
		if c.String("access_key_id") != "" {
			accessKeyID = c.String("access_key_id")
		}
		if c.String("region") != "" {
			region = c.String("region")
		}
		if !fake {
			fake = c.Bool("fake")
		}

		if secretAccessKey == "" {
			errs = append(errs, "Specify a valid secret access key with --secret_access_key flag")
		}

		if accessKeyID == "" {
			errs = append(errs, "Specify a valid access key id with --access_key_id flag")
		}

		if region == "" {
			errs = append(errs, "Specify a valid region with --region flag")
		}

		if len(errs) > 0 {
			color.Red("Please, fix the error shown below to continue")
			for _, e := range errs {
				fmt.Println("  - " + e)
			}
			return nil
		}

		rtype := "aws"

		if fake {
			rtype = "aws-fake"
		}
		body, err := m.CreateAWSProject(cfg.Token, name, rtype, region, accessKeyID, secretAccessKey)
		if err != nil {
			color.Red(body)
		} else {
			color.Green("Project '" + name + "' successfully created ")
		}
		return nil
	},
}

// UpdateAWSProject : Updates the specified VCloud project
var UpdateAWSProject = cli.Command{
	Name:        "aws",
	Usage:       h.T("aws.project.create.usage"),
	ArgsUsage:   h.T("aws.project.create.usage"),
	Description: h.T("aws.project.create.usage"),
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "access_key_id",
			Value: "",
			Usage: "Your AWS access key id",
		},
		cli.StringFlag{
			Name:  "secret_access_key",
			Value: "",
			Usage: "Your AWS secret access key",
		},
	},
	Action: func(c *cli.Context) error {
		var accessKeyID, secretAccessKey string
		m, cfg := setup(c)
		if cfg.Token == "" {
			color.Red("You're not allowed to perform this action, please log in")
			return nil
		}

		if len(c.Args()) == 0 {
			color.Red("You should specify the project name")
			return nil
		}
		name := c.Args()[0]
		accessKeyID = c.String("access_key_id")
		secretAccessKey = c.String("secret_access_key")

		if accessKeyID == "" {
			color.Red("You should specify your aws access key id with '--access_key_id' flag")
			return nil
		}
		if secretAccessKey == "" {
			color.Red("You should specify your aws secret access key with '--secret_access_key' flag")
			return nil
		}

		err := m.UpdateAWSProject(cfg.Token, name, accessKeyID, secretAccessKey)
		if err != nil {
			color.Red(err.Error())
			return nil
		}
		color.Green("Project " + name + " successfully updated")

		return nil
	},
}
