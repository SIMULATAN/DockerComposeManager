package cmd

import (
	"context"
	"fmt"
	cl "github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/compose/v2/cmd/formatter"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

type composeOptions struct {
	*cl.ProjectOptions
}

type upOptions struct {
	*composeOptions
	Detach             bool
	noStart            bool
	noDeps             bool
	cascadeStop        bool
	exitCodeFrom       string
	scale              []string
	noColor            bool
	noPrefix           bool
	attachDependencies bool
	attach             []string
	wait               bool
}

func (opts upOptions) apply(project *types.Project, services []string) error {
	if opts.noDeps {
		enabled, err := project.GetServices(services...)
		if err != nil {
			return err
		}
		for _, s := range project.Services {
			if !utils.StringContains(services, s.Name) {
				project.DisabledServices = append(project.DisabledServices, s)
			}
		}
		project.Services = enabled
	}

	if opts.exitCodeFrom != "" {
		_, err := project.GetService(opts.exitCodeFrom)
		if err != nil {
			return err
		}
	}

	for _, scale := range opts.scale {
		split := strings.Split(scale, "=")
		if len(split) != 2 {
			return fmt.Errorf("invalid --scale option %q. Should be SERVICE=NUM", scale)
		}
		name := split[0]
		replicas, err := strconv.Atoi(split[1])
		if err != nil {
			return err
		}
		err = setServiceScale(project, name, uint64(replicas))
		if err != nil {
			return err
		}
	}

	return nil
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
		project, err := cl.ProjectFromOptions(&cl.ProjectOptions{Name: config.Name, ConfigPaths: []string{filepath.Join(path, config.File)}, Environment: map[string]string{}})
		CheckErr(err, "Could not create compose project '"+config.Name+"' stored in", "'"+filepath.Join(path, config.File)+"'")
		Debug("Project:", project)

		Warn(project.Services)

		runUp(context.TODO(), compose.NewComposeService(dockerCli), upOptions, project, nil)

		// TODO: start the project
		return nil
	}))
}

func runUp(ctx context.Context, backend api.Service, upOptions upOptions, project *types.Project, services []string) error {
	if len(project.Services) == 0 {
		return fmt.Errorf("no service selected")
	}

	var consumer api.LogConsumer
	if !upOptions.Detach {
		consumer = formatter.NewLogConsumer(ctx, os.Stdout, !upOptions.noColor, !upOptions.noPrefix)
	}

	attachTo := services
	if len(upOptions.attach) > 0 {
		attachTo = upOptions.attach
	}
	if upOptions.attachDependencies {
		attachTo = project.ServiceNames()
	}

	create := api.CreateOptions{
		Services:             services,
		RemoveOrphans:        false,
		IgnoreOrphans:        false,
		Recreate:             api.RecreateDiverged,
		RecreateDependencies: api.RecreateDiverged,
		Inherit:              true,
		QuietPull:            false,
	}

	if upOptions.noStart {
		return backend.Create(ctx, project, create)
	}

	return backend.Up(ctx, project, api.UpOptions{
		Create: create,
		Start: api.StartOptions{
			Project:      project,
			Attach:       consumer,
			AttachTo:     attachTo,
			ExitCodeFrom: upOptions.exitCodeFrom,
			CascadeStop:  upOptions.cascadeStop,
			Wait:         upOptions.wait,
		},
	})
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
