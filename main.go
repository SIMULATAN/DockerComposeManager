package main

import "DockerComposeManager/cmd"

func main() {
	err := cmd.Execute()
	if err != nil {
		return
	}
}
