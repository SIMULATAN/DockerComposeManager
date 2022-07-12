package cmd

import (
	"fmt"
	"github.com/docker/cli/cli/command"
	"github.com/jwalton/gchalk"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

var dockerCli command.Cli

func RootCommand(cli command.Cli) *cobra.Command {
	dockerCli = cli
	rootCmd := &cobra.Command{
		Use:   "DockerComposeManager",
		Short: "Manage your Docker Compose Files with ease",
		Long:  "Manage your Docker Compose Files with ease, with features like environment variable substitution, groups, orders, autostart utilities, and more.",
		Run: func(cmd *cobra.Command, args []string) {
			// root command - only runs if no subcommand is specified
			fmt.Println("Welcome to Docker Compose Manager!\nAdd the '-h' flag to see all available commands.")
		},
	}

	// add global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ./"+defaultConfigName+")")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&dockerAddr, "docker-addr", "d", "unix:///var/run/docker.sock", "docker address (default is unix:///var/run/docker.sock)")

	rootCmd.AddCommand(
		configsCmd,
		initCmd,
	)

	return rootCmd
}

var cfgFile string
var verbose bool
var dockerAddr string

const defaultConfigName = ".dcm.yml"

func init() {
	// load config file
	cobra.OnInitialize(initConfig)
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
	viper.SetEnvPrefix("DCM")
	viper.EnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func loadConfig() {
	if err := viper.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			Warn("No project file found.")
		} else {
			fmt.Println("Error occurred while reading config:", err) // failed to read configuration file whilst it exists
			os.Exit(-1)
		}
	}

	Debug("Using config file", viper.ConfigFileUsed())
}

func saveConfig(v *viper.Viper) error {
	// make parents dirs
	cobra.CheckErr(os.MkdirAll(filepath.Dir(v.ConfigFileUsed()), os.ModePerm))
	return v.WriteConfig()
}

func CheckErr(err error, prefix ...string) {
	if err != nil {
		// prevent space before ":"
		Error(strings.Join(prefix, " ")+":", err)
		os.Exit(-1)
	}
}

func Debug(msg ...any) {
	if verbose {
		fmt.Println(gchalk.Yellow("[DEBUG] " + arrToStr(msg)))
	}
}

func Warn(msg ...any) {
	// space to make the messages in a column
	fmt.Println(gchalk.RGB(255, 136, 0)("[WARN ] " + arrToStr(msg)))
}

func Error(msg ...any) {
	fmt.Println(gchalk.Red("[ERROR] " + arrToStr(msg)))
}

// arrToStr converts an array of any to a string, joined by a space
// if the array is empty, it returns an empty string
func arrToStr(msg []any) string {
	var result = ""
	for i, v := range msg {
		var parsed = fmt.Sprintf("%+v", v)
		if i < len(msg)-1 {
			parsed += " "
		}
		result += parsed
	}
	return result
}
