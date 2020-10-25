package main

import (
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/spf13/cobra"
	"log"
	"path/filepath"
	"strings"
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

	ProfilePicturePath string
}

var slackSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a custom slack status",
	Long: `Running with no arguments will cause your status to be cleared.
When enabling do not disturb (dnd) you must specify a duration.
`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath := GetConfigFilepath(_stringOrPanic(cmd.Root().PersistentFlags().GetString("config")))

		duration := 0
		durationStr := _stringOrPanic(cmd.PersistentFlags().GetString("duration"))
		if durationStr != "" {
			d, err := time.ParseDuration(durationStr)
			if err != nil {
				log.Fatalf("failed to parse duration string: %s \n", durationStr)
			}

			if d.Minutes() < 1 {
				log.Fatalf("duration needs to be at least 1 minute not: %f minute(s)\n", d.Minutes())
			}

			duration = int(d.Minutes())
		}

		dnd := _boolOrPanic(cmd.PersistentFlags().GetBool("dnd"))
		if duration == 0 && dnd {
			log.Fatal("do not disturb requires a duration")
		}

		// Check if profile picture flag (if present) points to valid image file
		picPath := _stringOrPanic(cmd.PersistentFlags().GetString("profilePic"))
		if picPath != "" {
			isFile_, err := isFile(picPath)
			if err != nil {
				log.Fatal("Failed to check if profile picture path is a file: " + picPath)
			}

			if !isFile_ {
				log.Fatal("Profile picture file path is not valid: " + picPath)
			}

			isValid, err := isValidImage(picPath)
			if err != nil {
				log.Fatal("Failed to verify if profile picture file is valid")
			}

			if !isValid {
				log.Fatal("Invalid profile picture file, valid foramts are: jpeg, jpg, png & gif")
			}

		}

		s := Status{
			Message:            _stringOrPanic(cmd.PersistentFlags().GetString("message")),
			Emoji:              _stringOrPanic(cmd.PersistentFlags().GetString("emoji")),
			Duration:           duration,
			Away:               _boolOrPanic(cmd.PersistentFlags().GetBool("away")),
			DoNotDisturb:       dnd,
			Group:              _stringOrPanic(cmd.PersistentFlags().GetString("group")),
			Workspace:          _stringOrPanic(cmd.PersistentFlags().GetString("workspace")),
			ProfilePicturePath: picPath,
		}

		c := Config{}
		if err := c.Parse(configPath); err != nil {
			log.Fatal(err)
		}

		workspacesApplied, err := s.Apply(c.Workspaces)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Successfully applied to %d wokrspaces\n", workspacesApplied)
	},
}

func init() {
	slackSetCmd.PersistentFlags().Bool("away", false, "Set your status as away")
	slackSetCmd.PersistentFlags().StringP("group", "g", "", "Limit setting of status to a group")
	slackSetCmd.PersistentFlags().StringP("workspace", "w", "", "Limit setting of status to a workspace")
	slackSetCmd.PersistentFlags().StringP("duration", "d", "", "Set status duration, units can be: [m,h]. Leave blank for for no expiration")
	slackSetCmd.PersistentFlags().Bool("dnd", false, "Set status as do not disturb")
	slackSetCmd.PersistentFlags().StringP("emoji", "e", ":male-technologist:", "Emoji to set when setting your status")
	slackSetCmd.PersistentFlags().StringP("message", "m", "", "Status message")
	slackSetCmd.PersistentFlags().StringP("profilePic", "p", "", "Profile picture path (valid formats: jpeg, jpg, png & gif)")

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

	if err := api.SetUserCustomStatusWithUser(workspace.User, s.Message, emoji, duration); err != nil {
		return errors.Wrap(err, "failed to set custom status in: "+workspace.Name)
	}

	if s.DoNotDisturb {
		log.Printf("[%s] setting do not disturb for %d minute(s)\n", workspace.Name, s.Duration)
		if _, err := api.SetSnooze(s.Duration); err != nil {
			return errors.Wrap(err, "failed to set snooze in: "+workspace.Name)
		}
	} else {
		log.Printf("[%s] resetting do not disturb\n", workspace.Name)
		if _, err := api.SetSnooze(0); err != nil {
			return errors.Wrap(err, "failed to reset snooze in: "+workspace.Name)
		}
	}

	if s.ProfilePicturePath != "" {
		// Set profile picture
		log.Printf("[%s] setting profile picture: %s\n", workspace.Name, s.ProfilePicturePath)
		if err := api.SetUserPhoto(s.ProfilePicturePath, slack.UserSetPhotoParams{}); err != nil {
			return errors.Wrap(err, "failed to set profile picture in: "+workspace.Name)
		}
	}

	return nil
}

func isValidImage(path string) (bool, error) {
	fileName := filepath.Base(path)
	fileNameSep := strings.Split(fileName, ".")
	if len(fileNameSep) < 2 {
		return false, errors.New("file doesn't have an extension")
	}

	fileFormat := fileNameSep[len(fileNameSep)-1]

	validFileFormats := []string{"jpeg", "jpg", "gif", "png"}
	for _, format := range validFileFormats {
		if strings.ToLower(fileFormat) == format {
			return true, nil
		}
	}

	return false, nil
}
