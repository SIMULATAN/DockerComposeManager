package cmd

import (
	"fmt"
	"github.com/compose-spec/compose-go/cli"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

var argsValidator = func(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("invalid argument count")
	}
	// TODO: add more commands
	if args[0] != "up" {
		return fmt.Errorf("invalid command")
	}
	return nil
}

// Metadata is the information stored in .dcmeta.yml files and provide DCM-Related information
type Metadata struct {
	File string `yaml:"file"`
	Name string `yaml:"name"`
}

var configsCmd = &cobra.Command{
	Use:   "configs <up/down>",
	Short: "Spin up or down your compose files",
	Long:  `The main command that starts or stops your compose files.`,
	Args:  argsValidator,
	Run: func(cmd *cobra.Command, args []string) {
		loadConfig()
		if args[0] == "up" {
			up()
		}
	},
}

func up() {
	home, err := os.Getwd()
	cobra.CheckErr(err)

	cobra.CheckErr(filepath.WalkDir(home, func(path string, d os.DirEntry, err error) error {
		// check if we got a file
		if !d.IsDir() {
			return nil
		}

		// check if a docker compose file with the default name exists
		exists, fileFound := checkForComposeFiles(path)

		// read in metadata - dcmetaData could be null and that's ok
		dcmetaFile := filepath.Join(path, ".dcmeta.yml")
		Debug("Trying to read " + dcmetaFile + "...")
		dcmetaData, err := os.ReadFile(dcmetaFile)

		if !exists && os.IsNotExist(err) {
			// no metadata file, no compose file, noop
			Debug("No compose or metadata file found in", path, "skipping...")
			return nil
		}

		// if the file exists but errors out, we count it as a fail
		if !os.IsNotExist(err) {
			CheckErr(err, "Error reading '"+dcmetaFile+"'")
		}
		Debug("Read metadata from", dcmetaFile)

		// initial properties, values that present in the metadata file will get overwritten here
		var config = Metadata{
			File: fileFound,
			Name: filepath.Base(path),
		}

		// config, possibly nil, contains the metadata stored in .dcmeta.yml as a "Metadata" object
		if err == nil {
			Debug("Parsing metadata from", dcmetaFile)
			CheckErr(yaml.Unmarshal(dcmetaData, &config), "Parsing .dcmeta.yml failed")
			Debug("Parsed metadata from", dcmetaFile)
		}

		Debug("Loaded Config:", config)

		// make a project from the settings - the empty map is to avoid a nil pointer
		project, err := cli.ProjectFromOptions(&cli.ProjectOptions{Name: config.Name, ConfigPaths: []string{filepath.Join(path, config.File)}, Environment: map[string]string{}})
		CheckErr(err, "Could not create compose project '"+config.Name+"' stored in", "'"+filepath.Join(path, config.File)+"'")
		Debug("Project:", project)

		_, err = client.NewClientWithOpts(client.WithHost(dockerAddr))
		CheckErr(err, "Could not create docker client")
		fmt.Println("Created docker client")

		// TODO: start the project
		return nil
	}))
}

func checkForComposeFiles(home string) (bool, string) {
	// check for default compose files
	return exists(home, "docker-compose.yml", "docker-compose.yaml")
}

func exists(home string, files ...string) (bool, string) {
	for _, file := range files {
		if _, err := os.Open(filepath.Join(home, file)); err == nil {
			return true, file
		}
	}
	return false, ""
}
