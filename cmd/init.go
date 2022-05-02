package cmd

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// check if file exists
		if _, openErr := os.Open(viper.ConfigFileUsed()); openErr == nil {
			// y/N prompt to confirm overwriting existing config
			prompt := promptui.Prompt{
				Label:     "You already have a config file. Do you want to overwrite it",
				IsConfirm: true,
			}

			_, promptErr := prompt.Run()

			// error is only nil if the user confirms with "y"
			if promptErr != nil {
				fmt.Println("Aborting, no changes were made.")
				os.Exit(0)
			}
		}

		fmt.Println("Initializing a new project in '" + viper.ConfigFileUsed() + "' called '" + args[0] + "'")
		initProject(args[0])
	},
}

func initProject(name string) {
	projectName = name
	viper.Set("projectname", projectName)
	cobra.CheckErr(saveConfig(viper.GetViper()))
}

func init() {
	rootCmd.AddCommand(initCmd)
}
