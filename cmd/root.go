package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

var rootCmd = &cobra.Command{
	Use:   "DockerComposeManager",
	Short: "Manage your Docker Compose Files with ease",
	Long:  "Manage your Docker Compose Files with ease, with features like environment variable substitution, groups, orders, autostart utilities, and more.",
	Run:   func(cmd *cobra.Command, args []string) {},
}
var cfgFile string
var verbose bool
var projectName string

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		return err
	}
	cobra.CheckErr(saveConfig(viper.GetViper()))
	return nil
}

const defaultConfigName = ".dcm.yml"

func init() {
	// load config file
	cobra.OnInitialize(initConfig)

	// add global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ./.dcm.yml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func initConfig() {
	// load in config file
	if cfgFile != "" {
		// .ini files are a special case
		if filepath.Ext(cfgFile) == ".ini" {
			// ini has 'default.' as the prefix for keys
			fmt.Println("Ini is currently not supported due to compatiblity issues, please use one of the following formats:", viper.SupportedExts)
			os.Exit(1)
		}
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// get current working directory - the project configuration will lay there
		home, err := os.Getwd()
		cobra.CheckErr(err)

		// default path: current working directory + default config name
		viper.AddConfigPath(home)
		viper.SetConfigName(defaultConfigName)
		viper.SetConfigFile(filepath.Join(home, defaultConfigName))
		viper.SetConfigType("yml")
	}
	// load settings from env too
	viper.AutomaticEnv()
}

func loadConfig() {
	if err := viper.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No project file found. Please run the 'init' command to create one.")
		} else {
			fmt.Println("Error occurred while reading config:", err) // failed to read configuration file whilst it exists
		}
		os.Exit(-1)
	}

	if verbose {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func saveConfig(v *viper.Viper) error {
	cobra.CheckErr(os.MkdirAll(filepath.Dir(v.ConfigFileUsed()), os.ModePerm))
	return v.WriteConfig()
}
