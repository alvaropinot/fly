package commands

import (
	"fmt"
	"os"

	"github.com/concourse/atc"
	"github.com/concourse/fly/commands/internal/displayhelpers"
	"github.com/concourse/fly/rc"
	"github.com/concourse/fly/ui"
	"github.com/vito/go-interact/interact"
)

type ProviderConfig interface {
	Name() string
	DisplayName() string
	IsConfigured() bool
	Validate() error
}

type SetTeamCommand struct {
	TeamName        string    `short:"n" long:"team-name" required:"true" description:"The team to create or modify"`
	SkipInteractive bool      `long:"non-interactive" description:"Force apply configuration"`
	TeamFlags       TeamFlags `group:"Authentication"`
}

type TeamFlags struct {
	Users  []string `json:"users" long:"user" description:"List of auth users"`
	Groups []string `json:"groups" long:"group" description:"List of auth groups"`
	NoAuth bool     `long:"no-really-i-dont-want-any-auth" description:"Flag to disable any authorization method for your team"`
}

func (config TeamFlags) toMap() map[string][]string {
	auth := map[string][]string{}

	if len(config.Users) > 0 {
		auth["users"] = config.Users
	}

	if len(config.Groups) > 0 {
		auth["groups"] = config.Groups
	}

	return auth
}

func (command *SetTeamCommand) Execute([]string) error {
	target, err := rc.LoadTarget(Fly.Target, Fly.Verbose)
	if err != nil {
		return err
	}

	err = target.Validate()
	if err != nil {
		return err
	}

	err = command.ValidateFlags()
	if err != nil {
		return err
	}

	fmt.Println("Team Name:", command.TeamName)

	confirm := true
	if !command.SkipInteractive {
		confirm = false
		err = interact.NewInteraction("apply configuration?").Resolve(&confirm)
		if err != nil {
			return err
		}
	}

	if !confirm {
		displayhelpers.Failf("bailing out")
	}

	team := atc.Team{
		Auth: command.TeamFlags.toMap(),
	}

	_, created, updated, err := target.Client().Team(command.TeamName).CreateOrUpdate(team)
	if err != nil {
		return err
	}

	if created {
		fmt.Println("team created")
	} else if updated {
		fmt.Println("team updated")
	}

	return nil
}

func (command *SetTeamCommand) ValidateFlags() error {

	if command.TeamFlags.NoAuth {
		displayhelpers.PrintWarningHeader()
		fmt.Fprintln(ui.Stderr, ui.WarningColor("no auth methods configured. you asked for it!"))
		fmt.Fprintln(ui.Stderr, "")

	} else if len(command.TeamFlags.Groups) == 0 && len(command.TeamFlags.Users) == 0 {
		fmt.Fprintln(ui.Stderr, "no auth methods configured! to continue, run:")
		fmt.Fprintln(ui.Stderr, "")
		fmt.Fprintln(ui.Stderr, "    "+ui.Embolden("fly -t %s set-team -n %s --no-really-i-dont-want-any-auth", Fly.Target, command.TeamName))
		fmt.Fprintln(ui.Stderr, "")
		fmt.Fprintln(ui.Stderr, "this will leave the team open to anyone to mess with!")
		os.Exit(1)
	}

	return nil
}
