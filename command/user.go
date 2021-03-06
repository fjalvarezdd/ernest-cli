/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package command

// CmdUser subcommand
import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"text/tabwriter"
	"unicode"

	h "github.com/ernestio/ernest-cli/helper"
	"github.com/ernestio/ernest-cli/model"
	"github.com/ernestio/ernest-cli/view"
	"github.com/fatih/color"
	"github.com/howeyc/gopass"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

// ListUsers : Gets a list of accessible users
var ListUsers = cli.Command{
	Name:        "list",
	Usage:       h.T("user.list.usage"),
	ArgsUsage:   h.T("user.list.args"),
	Description: h.T("user.list.description"),
	Action: func(c *cli.Context) error {
		m, cfg := setup(c)
		users, err := m.ListUsers(cfg.Token)
		if err != nil {
			h.PrintError(err.Error())
		}

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 0, '\t', 0)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Group", "Admin"})
		for _, u := range users {
			id := strconv.Itoa(u.ID)
			admin := "no"
			if u.IsAdmin {
				admin = "yes"
			}
			table.Append([]string{id, u.Username, u.GroupName, admin})
		}
		table.Render()

		return nil
	},
}

// CreateUser : Creates a new user
var CreateUser = cli.Command{
	Name:        "create",
	Usage:       h.T("user.create.usage"),
	ArgsUsage:   h.T("user.create.args"),
	Description: h.T("user.create.description"),
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "email",
			Value: "",
			Usage: "Email for the user",
		},
	},
	Action: func(c *cli.Context) error {
		if len(c.Args()) < 1 {
			h.PrintError("You should specify an user username and a password")
		}
		if len(c.Args()) < 2 {
			h.PrintError("You should specify the user password")
		}

		usr := c.Args()[0]
		email := c.String("email")
		pwd := c.Args()[1]
		m, cfg := setup(c)
		err := m.CreateUser(cfg.Token, usr, email, usr, pwd)
		if err != nil {
			h.PrintError(err.Error())
		}
		color.Green("User " + usr + " successfully created")
		return nil
	},
}

// PasswordUser : Allows users or admins to change its passwords
var PasswordUser = cli.Command{
	Name:        "change-password",
	Usage:       h.T("user.change_password.usage"),
	Description: h.T("user.change_password.description"),
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "user",
			Value: "",
			Usage: "The username of the user to change password",
		},
		cli.StringFlag{
			Name:  "password",
			Value: "",
			Usage: "The new user password",
		},
		cli.StringFlag{
			Name:  "current-password",
			Value: "",
			Usage: "The current user password",
		},
	},
	Action: func(c *cli.Context) error {
		m, cfg := setup(c)

		username := c.String("user")
		password := c.String("password")
		currentPassword := c.String("current-password")

		session, err := m.GetSession(cfg.Token)
		if err != nil {
			h.PrintError("You don’t have permissions to perform this action")
		}

		if !session.IsAdmin && username != "" {
			h.PrintError("You don’t have permissions to perform this action")
		}

		if session.IsAdmin && username != "" {
			if password == "" {
				h.PrintError("Please provide a valid password for the user with `--password`")
			}

			// Just change the password with the given values for the given user
			usr, err := m.GetUserByUsername(cfg.Token, username)
			if err = m.ChangePasswordByAdmin(cfg.Token, usr.ID, usr.Username, usr.GroupID, password); err != nil {
				h.PrintError(err.Error())
			}
			color.Green("`" + usr.Username + "` password has been changed")
		} else {
			// Ask the user for credentials
			var users []model.User
			if users, err = m.ListUsers(cfg.Token); err != nil {
				h.PrintError("You don’t have permissions to perform this action")
			}
			if len(users) == 0 {
				h.PrintError("You don’t have permissions to perform this action")
			}

			var user model.User
			for _, u := range users {
				if u.Username == cfg.User {
					user = u
					break
				}
			}

			oldpassword := currentPassword
			newpassword := password
			rnewpassword := password

			if oldpassword == "" || newpassword == "" {
				fmt.Printf("You're about to change your password, please respond the questions below: \n")
				fmt.Printf("Current password: ")
				opass, _ := gopass.GetPasswdMasked()
				oldpassword = string(opass)

				fmt.Printf("New password: ")
				npass, _ := gopass.GetPasswdMasked()
				newpassword = string(npass)

				fmt.Printf("Confirm new password: ")
				rnpass, _ := gopass.GetPasswdMasked()
				rnewpassword = string(rnpass)
			}

			if newpassword != rnewpassword {
				h.PrintError("Aborting... New password and confirmation doesn't match.")
			}

			err = m.ChangePassword(cfg.Token, user.ID, user.Username, user.GroupID, oldpassword, newpassword)
			if err != nil {
				h.PrintError(err.Error())
			}
			color.Green("Your password has been changed")
		}

		return nil
	},
}

// DisableUser : Will disable a user (change its password)
var DisableUser = cli.Command{
	Name:        "disable",
	Usage:       h.T("user.disable.usage"),
	ArgsUsage:   h.T("user.disable.args"),
	Description: h.T("user.disable.description"),
	Action: func(c *cli.Context) error {
		if len(c.Args()) < 1 {
			h.PrintError("You should specify an username")
		}

		m, cfg := setup(c)
		username := c.Args()[0]

		session, err := m.GetSession(cfg.Token)
		if err != nil {
			h.PrintError("You don’t have permissions to perform this action")
		}

		if !session.IsAdmin {
			h.PrintError("You don’t have permissions to perform this action")
		}

		user, err := m.GetUserByUsername(cfg.Token, username)
		if err != nil {
			h.PrintError(err.Error())
		}

		if err = m.ChangePasswordByAdmin(cfg.Token, user.ID, user.Username, user.GroupID, randString(16)); err != nil {
			h.PrintError(err.Error())
		}

		color.Green("Account `" + username + "` has been disabled")
		return nil
	},
}

// InfoUser :
var InfoUser = cli.Command{
	Name:        "info",
	Usage:       h.T("user.info.usage"),
	Description: h.T("user.info.description"),
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "user",
			Value: "",
			Usage: "Username",
		},
	},
	Action: func(c *cli.Context) error {
		m, cfg := setup(c)
		session, err := m.GetSession(cfg.Token)
		if err != nil {
			h.PrintError("You don’t have permissions to perform this action")
		}

		username := c.String("user")
		if username != "" && session.IsAdmin == false {
			h.PrintError("You don’t have permissions to access '" + username + "' information")
		}
		if username == "" {
			username = cfg.User
		}

		user, err := m.GetUser(cfg.Token, username)
		if err != nil {
			h.PrintError(err.Error())
		}

		view.PrintUserInfo(user)
		return nil
	},
}

// generate random string
func randString(n int) string {
	g := big.NewInt(0)
	max := big.NewInt(130)
	bs := make([]byte, n)

	for i := range bs {
		g, _ = rand.Int(rand.Reader, max)
		r := rune(g.Int64())
		for !unicode.IsNumber(r) && !unicode.IsLetter(r) {
			g, _ = rand.Int(rand.Reader, max)
			r = rune(g.Int64())
		}
		bs[i] = byte(g.Int64())
	}
	return string(bs)
}

// CmdUser ...
var CmdUser = cli.Command{
	Name:  "user",
	Usage: h.T("user.usage"),
	Subcommands: []cli.Command{
		ListUsers,
		CreateUser,
		PasswordUser,
		DisableUser,
		InfoUser,
	},
}
