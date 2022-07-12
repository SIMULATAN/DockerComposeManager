package main

import (
	"DockerComposeManager/cmd"
	"github.com/docker/cli/cli-plugins/manager"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"
)

func main() {
	plugin.Run(func(dockerCli command.Cli) *cobra.Command {
		root := cmd.RootCommand(dockerCli)
		return root
	}, manager.Metadata{})
}
