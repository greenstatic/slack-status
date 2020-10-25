package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

const (
	versionMajor  = 2
	versionMinor  = 0
	versionBugfix = 0
)

var rootCmd = &cobra.Command{
	Use:   "slack-status",
	Short: "Set your status in Slack.",
	Long: `Set your status in Slack.

To setup the required authentication token(s) read the "init" subcommand help.

To set your status see the "set" subcommand.

Source code: https://github.com/greenstatic/slack-status`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			log.Panic(err)
		}
	},
}
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("slack-status v%d.%d.%d\n", versionMajor, versionMinor, versionBugfix)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", configPathDefault, "Config file")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(slackSetCmd)
	rootCmd.AddCommand(versionCmd)
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

func isFile(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.Mode().IsRegular(), err
}
