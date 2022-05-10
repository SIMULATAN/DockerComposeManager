package main

import (
	"DockerComposeManager/cmd"
	"github.com/spf13/cobra"
)

func main() {
	cobra.CheckErr(cmd.RootCommand().Execute())
}
