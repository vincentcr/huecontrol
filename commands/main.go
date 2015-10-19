package commands

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd := &cobra.Command{Use: "htflood"}
	setupCommands(rootCmd)
	rootCmd.Execute()
}

func setupCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(lightsCommand)
	rootCmd.AddCommand(groupsCommand)
	rootCmd.AddCommand(schedulesCommand)
}

func checkedRun(run func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {

	return func(cmd *cobra.Command, args []string) {
		if err := run(cmd, args); err != nil {
			log.Fatalf("%v\n", err)
			os.Exit(1)
		}
	}
}
