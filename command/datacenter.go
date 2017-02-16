/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package command

// CmdDatacenter subcommand
import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/ernestio/ernest-cli/model"
	"github.com/ernestio/ernest-cli/view"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

// UpdateDatacenters : Will update the datacenter specific fields
var UpdateDatacenters = cli.Command{
	Name:        "update",
	Usage:       "Updates an existing datacenter.",
	Description: "Update an existing datacenter on the targeted instance of Ernest.",
	Subcommands: []cli.Command{
		UpdateVCloudDatacenter,
		UpdateAWSDatacenter,
	},
}

// CreateDatacenters ...
var CreateDatacenters = cli.Command{
	Name:        "create",
	Usage:       "Create a new datacenter.",
	Description: "Create a new datacenter on the targeted instance of Ernest.",
	Subcommands: []cli.Command{
		CreateVcloudDatacenter,
		CreateAWSDatacenter,
		CreateAzureDatacenter,
	},
}

// CmdDatacenter ...
var CmdDatacenter = cli.Command{
	Name:  "datacenter",
	Usage: "Datacenter related subcommands",
	Subcommands: []cli.Command{
		ListDatacenters,
		CreateDatacenters,
		UpdateDatacenters,
		DeleteDatacenter,
	},
}

// ListDatacenters ...
var ListDatacenters = cli.Command{
	Name:      "list",
	Usage:     "List available datacenters.",
	ArgsUsage: " ",
	Description: `List available datacenters.

   Example:
    $ ernest datacenter list
	`,
	Action: func(c *cli.Context) error {
		m, cfg := setup(c)
		if cfg.Token == "" {
			color.Red("You're not allowed to perform this action, please log in")
			return nil
		}
		datacenters, err := m.ListDatacenters(cfg.Token)
		if err != nil {
			color.Red(err.Error())
			return nil
		}

		view.PrintDatacenterList(datacenters)

		return nil
	},
}

// CreateAWSDatacenter : Creates an AWS datacenter
var CreateAWSDatacenter = cli.Command{
	Name:  "aws",
	Usage: "Create a new aws datacenter.",
	Description: `Create a new AWS datacenter on the targeted instance of Ernest.

	Example:
	 $ ernest datacenter create aws --region us-west-2 --access_key_id AKIAIOSFODNN7EXAMPLE --secret_access_key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY my_datacenter

   Template example:
    $ ernest datacenter create aws --template mydatacenter.yml mydatacenter
    Where mydatacenter.yaml will look like:
      ---
      fake: true
      access_key_id : AKIAIOSFODNN7EXAMPLE
      secret_access_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
      region: us-west-2
	 `,
	ArgsUsage: "<datacenter-name>",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "region, r",
			Value: "",
			Usage: "Datacenter region",
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
			Usage: "Datacenter template",
		},
		cli.BoolFlag{
			Name:  "fake, f",
			Usage: "Fake datacenter",
		},
	},
	Action: func(c *cli.Context) error {
		var errs []string
		var accessKeyID, secretAccessKey, region string
		var fake bool
		m, cfg := setup(c)

		if len(c.Args()) < 1 {
			msg := "You should specify the datacenter name"
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
			var t model.DatacenterTemplate
			if err := getDatacenterTemplate(template, &t); err != nil {
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
		if fake == false {
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
		body, err := m.CreateAWSDatacenter(cfg.Token, name, rtype, region, accessKeyID, secretAccessKey)
		if err != nil {
			color.Red(body)
		} else {
			color.Green("Datacenter '" + name + "' successfully created ")
		}
		return nil
	},
}

// CreateVcloudDatacenter : Creates a VCloud Datacenter
var CreateVcloudDatacenter = cli.Command{
	Name:  "vcloud",
	Usage: "Create a new vcloud datacenter.",
	Description: `Create a new vcloud datacenter on the targeted instance of Ernest.

   Example:
    $ ernest datacenter create vcloud --user username --password xxxx --org MY-ORG-NAME --vse-url http://vse.url --vcloud-url https://myernest.com --public-network MY-PUBLIC-NETWORK mydatacenter

   Template example:
    $ ernest datacenter create vcloud --template mydatacenter.yml mydatacenter
    Where mydatacenter.yaml will look like:
      ---
      fake: true
      org: org
      password: pwd
      public-network: MY-NETWORK
      user: bla
      vcloud-url: "http://ss.com"
      vse-url: "http://ss.com"

	`,
	ArgsUsage: "<datacenter-name>",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "user",
			Value: "",
			Usage: "Your VCloud valid user name",
		},
		cli.StringFlag{
			Name:  "password",
			Value: "",
			Usage: "Your VCloud valid password",
		},
		cli.StringFlag{
			Name:  "org",
			Value: "",
			Usage: "Your vCloud Organization",
		},
		cli.StringFlag{
			Name:  "vse-url",
			Value: "",
			Usage: "VSE URL",
		},
		cli.StringFlag{
			Name:  "vcloud-url",
			Value: "",
			Usage: "VCloud URL",
		},
		cli.StringFlag{
			Name:  "public-network",
			Value: "",
			Usage: "Public Network",
		},
		cli.StringFlag{
			Name:  "template",
			Value: "",
			Usage: "Datacenter template",
		},
		cli.BoolFlag{
			Name:  "fake",
			Usage: "Fake datacenter",
		},
	},
	Action: func(c *cli.Context) error {
		var errs []string

		if len(c.Args()) == 0 {
			color.Red("You should specify the datacenter name")
			return nil
		}
		m, cfg := setup(c)
		if cfg.Token == "" {
			color.Red("You're not allowed to perform this action, please log in")
			return nil
		}

		name := c.Args()[0]
		var url, network, user, org, password, username string
		var fake bool

		template := c.String("template")
		if template != "" {
			var t model.DatacenterTemplate
			if err := getDatacenterTemplate(template, &t); err != nil {
				color.Red(err.Error())
				return nil
			}
			url = t.URL
			network = t.Network
			user = t.User
			org = t.Org
			password = t.Password
			fake = t.Fake
		}
		if c.String("vcloud-url") != "" {
			url = c.String("vcloud-url")
		}
		if c.String("public-network") != "" {
			network = c.String("public-network")
		}
		if c.String("user") != "" {
			user = c.String("user")
		}
		if c.String("org") != "" {
			org = c.String("org")
		}
		if c.String("password") != "" {
			password = c.String("password")
		}
		if fake == false {
			fake = c.Bool("fake")
		}
		username = user + "@" + org

		if url == "" {
			errs = append(errs, "Specify a valid VCloud URL with --vcloud-url flag")
		}
		if network == "" {
			errs = append(errs, "Specify a valid public network with --public-network flag")
		}
		if user == "" {
			errs = append(errs, "Specify a valid user name with --user")
		}
		if org == "" {
			errs = append(errs, "Specify a valid organization with --org")
		}
		if password == "" {
			errs = append(errs, "Specify a valid password with --password")
		}
		rtype := "vcloud"
		if fake {
			rtype = "vcloud-fake"
		}
		if len(errs) > 0 {
			color.Red("Please, fix the error shown below to continue")
			for _, e := range errs {
				fmt.Println("  - " + e)
			}
			return nil
		}

		body, err := m.CreateVcloudDatacenter(cfg.Token, name, rtype, username, password, url, network, c.String("vse-url"))
		if err != nil {
			color.Red(body)
		} else {
			color.Green("Datacenter '" + name + "' successfully created ")
		}
		return nil
	},
}

// CreateAzureDatacenter : Creates an AWS datacenter
var CreateAzureDatacenter = cli.Command{
	Name:  "azure",
	Usage: "Create a new azure datacenter.",
	Description: `Create a new Azure datacenter on the targeted instance of Ernest.

	Example:
	 $ ernest-cli datacenter create azure --subscription_id XXXXXXX-XXXX-XXXX-XXXXXXXXXXXXXXXX --client_id XXXXXXX-XXXX-XXXX-XXXXXXXXXXXXXXXX --client_secret XXxxxXxXxXXXxxXXXXXxxXxxXxxxXxXxXxXxXxxxX= --tenant_id XXXXXXX-XXXX-XXXX-XXXXXXXXXXXXXXXX --environment public my_azure_datacenter_name""

   Template example:
    $ ernest datacenter create azure --template mydatacenter.yml mydatacenter
    Where mydatacenter.yaml will look like:
      ---
			subscription_id: XXXXXXX-XXXX-XXXX-XXXXXXXXXXXXXXXX
			client_id: XXXXXXX-XXXX-XXXX-XXXXXXXXXXXXXXXX
			client_secret: XXxxxXxXxXXXxxXXXXXxxXxxXxxxXxXxXxXxXxxxX=
			tenant_id: XXXXXXX-XXXX-XXXX-XXXXXXXXXXXXXXXX
			environment public 
	 `,
	ArgsUsage: "<datacenter-name>",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "subscription_id, s",
			Value: "",
			Usage: "The azure subscription ID to use.",
		},
		cli.StringFlag{
			Name:  "client_id, ci",
			Value: "",
			Usage: "The client ID to use.",
		},
		cli.StringFlag{
			Name:  "client_secret, cs",
			Value: "",
			Usage: "The client secret to use.",
		},
		cli.StringFlag{
			Name:  "tenant_id, t",
			Value: "",
			Usage: "The tenant ID to use.",
		},
		cli.StringFlag{
			Name:  "template",
			Value: "",
			Usage: "Datacenter template",
		},
		cli.StringFlag{
			Name:  "environment, e",
			Value: "public",
			Usage: "Fake datacenter",
		},
		cli.BoolFlag{
			Name:  "fake, f",
			Usage: "Fake datacenter",
		},
	},
	Action: func(c *cli.Context) error {
		var errs []string
		var subscription, client, secret, tenant, environment string
		var fake bool
		m, cfg := setup(c)

		if len(c.Args()) < 1 {
			msg := "You should specify the datacenter name"
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
			var t model.DatacenterTemplate
			if err := getDatacenterTemplate(template, &t); err != nil {
				color.Red(err.Error())
				return nil
			}
			subscription = t.SubscriptionID
			secret = t.ClientSecret
			client = t.ClientID
			tenant = t.TenantID
			environment = t.Environment
			fake = t.Fake
		}
		if c.String("subscription_id") != "" {
			subscription = c.String("subscription_id")
		}
		if c.String("client_id") != "" {
			client = c.String("client_id")
		}
		if c.String("client_secret") != "" {
			secret = c.String("client_secret")
		}
		if c.String("tenant_id") != "" {
			tenant = c.String("tenant_id")
		}
		if c.String("environment") != "" {
			environment = c.String("environment")
		}
		if fake == false {
			fake = c.Bool("fake")
		}

		if subscription == "" {
			errs = append(errs, "Specify a valid subscription id with --subscription_id flag")
		}

		if client == "" {
			errs = append(errs, "Specify a valid client id id with --client_id flag")
		}

		if secret == "" {
			errs = append(errs, "Specify a valid client secret with --client_secret flag")
		}

		if tenant == "" {
			errs = append(errs, "Specify a valid tenant id with --tenant_id flag")
		}

		if environment != "public" && environment != "usgovernment" && environment != "german" && environment != "china" {
			errs = append(errs, "Specify a valid environment with --environment flag. Valid values are public, usgovernment, german and china")
		}

		if len(errs) > 0 {
			color.Red("Please, fix the error shown below to continue")
			for _, e := range errs {
				fmt.Println("  - " + e)
			}
			return nil
		}

		rtype := "azure"

		if fake {
			rtype = "azure-fake"
		}
		body, err := m.CreateAzureDatacenter(cfg.Token, name, rtype, subscription, client, secret, tenant, environment)
		if err != nil {
			color.Red(body)
		} else {
			color.Green("Datacenter '" + name + "' successfully created ")
		}
		return nil
	},
}

// DeleteDatacenter : Datacenter deletion command definition
var DeleteDatacenter = cli.Command{
	Name:      "delete",
	Usage:     "Deletes the specified datacenter.",
	ArgsUsage: "<datacenter-name>",
	Description: `Deletes the name specified datacenter.

   Example:
    $ ernest datacenter delete my_datacenter
	`,
	Action: func(c *cli.Context) error {
		m, cfg := setup(c)
		if cfg.Token == "" {
			color.Red("You're not allowed to perform this action, please log in")
			return nil
		}

		if len(c.Args()) == 0 {
			color.Red("You should specify the datacenter name")
			return nil
		}
		name := c.Args()[0]

		err := m.DeleteDatacenter(cfg.Token, name)
		if err != nil {
			color.Red(err.Error())
			return nil
		}
		color.Green("Datacenter " + name + " successfully removed")

		return nil
	},
}

// UpdateVCloudDatacenter : Updates the specified VCloud datacenter
var UpdateVCloudDatacenter = cli.Command{
	Name:      "vcloud",
	Usage:     "Updates the specified VCloud datacenter.",
	ArgsUsage: "<datacenter-name>",
	Description: `Updates the specified VCloud datacenter.

   Example:
    $ ernest datacenter update vcloud --user <me> --org <org> --password <secret> my_datacenter
	`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "user",
			Value: "",
			Usage: "Your VCloud valid user name",
		},
		cli.StringFlag{
			Name:  "password",
			Value: "",
			Usage: "Your VCloud valid password",
		},
		cli.StringFlag{
			Name:  "org",
			Value: "",
			Usage: "Your vCloud Organization",
		},
	},
	Action: func(c *cli.Context) error {
		var user, password, org string
		m, cfg := setup(c)
		if cfg.Token == "" {
			color.Red("You're not allowed to perform this action, please log in")
			return nil
		}
		if len(c.Args()) == 0 {
			color.Red("You should specify the datacenter name")
			return nil
		}
		name := c.Args()[0]
		user = c.String("user")
		password = c.String("password")
		org = c.String("org")

		if user == "" {
			color.Red("You should specify user name with '--user' flag")
			return nil
		}
		if password == "" {
			color.Red("You should specify user password with '--password' flag")
			return nil
		}
		if org == "" {
			color.Red("You should specify user org with '--org' flag")
			return nil
		}

		err := m.UpdateVCloudDatacenter(cfg.Token, name, user+"@"+org, password)
		if err != nil {
			color.Red(err.Error())
			return nil
		}
		color.Green("Datacenter " + name + " successfully updated")

		return nil
	},
}

// UpdateAWSDatacenter : Updates the specified VCloud datacenter
var UpdateAWSDatacenter = cli.Command{
	Name:      "aws",
	Usage:     "Updates the specified AWS datacenter.",
	ArgsUsage: "<datacenter-name>",
	Description: `Updates the specified AWS datacenter.

   Example:
		$ ernest datacenter update aws --access_key_id AKIAIOSFODNN7EXAMPLE --secret_access_key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY my_datacenter
	`,
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
			color.Red("You should specify the datacenter name")
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

		err := m.UpdateAWSDatacenter(cfg.Token, name, accessKeyID, secretAccessKey)
		if err != nil {
			color.Red(err.Error())
			return nil
		}
		color.Green("Datacenter " + name + " successfully updated")

		return nil
	},
}

func getDatacenterTemplate(template string, t *model.DatacenterTemplate) (err error) {
	payload, err := ioutil.ReadFile(template)
	if err != nil {
		return errors.New("Template file '" + template + "' not found")
	}
	if yaml.Unmarshal(payload, &t) != nil {
		return errors.New("Template file '" + template + "' is not valid yaml file")
	}
	return err
}
