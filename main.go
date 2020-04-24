package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os/user"
	"path/filepath"
	"time"
)

const (
	versionMajor = 1
	versionMinor = 0
	versionBugfix = 1
	configPathDefault = "~/slack-status"
)

var rootCmd = &cobra.Command{
	Use:   "slack-status",
	Short: "Set your status in Slack.",
	Long: `Set your status in Slack.

Running with no arguments will cause your status to be cleared.
When enabling do not disturb (dnd) you must specify a duration.

Source code: https://github.com/greenstatic/slack-status`,
	Run: func(cmd *cobra.Command, args []string) {
		if _boolOrPanic(cmd.PersistentFlags().GetBool("version")) {
			fmt.Printf("slack-status v%d.%d.%d\n", versionMajor, versionMinor, versionBugfix)
			return
		}

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

		configPath := _stringOrPanic(cmd.PersistentFlags().GetString("config"))
		if configPath == configPathDefault {
			usr, err := user.Current()
			if err != nil {
				log.Fatal(err)
			}
			configPath = filepath.Join(usr.HomeDir, "slack-status")
		}

		s := Status{
			Message:      _stringOrPanic(cmd.PersistentFlags().GetString("message")),
			Emoji:        _stringOrPanic(cmd.PersistentFlags().GetString("emoji")),
			Duration:     duration,
			Away:         _boolOrPanic(cmd.PersistentFlags().GetBool("away")),
			DoNotDisturb: dnd,
			Group:        _stringOrPanic(cmd.PersistentFlags().GetString("group")),
			Workspace:    _stringOrPanic(cmd.PersistentFlags().GetString("workspace")),
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

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "~/slack-status", "Config file")
	rootCmd.PersistentFlags().Bool("away", false, "Set your status as away")
	rootCmd.PersistentFlags().StringP("group", "g", "", "Limit setting of status to a group")
	rootCmd.PersistentFlags().StringP("workspace", "w", "", "Limit setting of status to a workspace")
	rootCmd.PersistentFlags().StringP("duration", "d", "", "Set status duration, units can be: [m,h]. Leave blank for for no expiration")
	rootCmd.PersistentFlags().Bool("dnd", false, "Set status as do not disturb")
	rootCmd.PersistentFlags().StringP("emoji", "e", ":male-technologist:", "Emoji to set when setting your status")
	rootCmd.PersistentFlags().StringP("message", "m", "", "Status message")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Show version number")
}

func _boolOrPanic(b bool, err error) bool {
	if err != nil {
		panic(err)
	}
	return b
}

func _intOrPanic(i int, err error) int {
	if err != nil {
		panic(err)
	}
	return i
}

func _stringOrPanic(s string, err error) string {
	if err != nil {
		panic(err)
	}
	return s
}
