package main

import (
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"log"
	"time"
)

type Status struct {
	Message  string
	Emoji    string
	Duration int
	Away     bool
	// Do not disturb duration
	DoNotDisturb bool
	// Limit to a Group
	Group string
	// Limit to a Workspace
	Workspace string
}

func (s *Status) Apply(workspaces []Workspace) (workspacesApplied int, err error) {
	for _, workspace := range workspaces {
		if s.Group != "" && !workspace.isInGroup(s.Group) {
			continue
		}
		if s.Workspace != "" && workspace.Name != s.Workspace {
			continue
		}

		// Its a match, lets set status
		if err := s.Set(workspace); err != nil {
			return workspacesApplied, err
		}

		workspacesApplied++
	}
	return
}

func (s *Status) Set(workspace Workspace) error {
	api := slack.New(workspace.AccessToken, slack.OptionDebug(true))

	presence := "auto"
	if s.Away {
		presence = "away"
	}

	if err := api.SetUserPresence(presence); err != nil {
		return errors.Wrap(err, "failed to set user preference in: "+workspace.Name)
	}

	// This is used to clear the status, if message is empty then the emoji should be too
	emoji := s.Emoji
	if s.Message == "" {
		emoji = ""
	}

	if s.Message == "" {
		log.Printf("[%s] clearing status message\n", workspace.Name)
	} else {
		if s.Duration == 0 {
			log.Printf("[%s] setting status message: '%s'\n", workspace.Name, s.Message)
		} else {
			log.Printf("[%s] setting status message: '%s' for %d minute(s)\n", workspace.Name, s.Message, s.Duration)
		}
	}

	duration := int64(s.Duration)
	if duration != 0 {
		duration = time.Now().Add(time.Duration(s.Duration) * time.Minute).Unix()
	}

	if err := api.SetUserCustomStatus(s.Message, emoji, duration); err != nil {
		return errors.Wrap(err, "failed to set custom status in: "+workspace.Name)
	}

	if s.DoNotDisturb {
		log.Printf("[%s] setting do not disturb for %d minute(s)\n", workspace.Name, s.Duration)
		if _, err := api.SetSnooze(s.Duration); err != nil {
			return errors.Wrap(err, "failed to set snooze in: " + workspace.Name)
		}
	}

	return nil
}
