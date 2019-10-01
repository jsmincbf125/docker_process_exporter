package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func isArrayContains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}

	return false
}

func main() {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	adminContainer := []string{"/rancher-agent", "/node_exporter", "/dcgm-epxporter"}

	for _, container := range containers {
		if !isArrayContains(adminContainer, container.Names[0]) {
			processes, _ := cli.ContainerTop(context.Background(), container.ID, []string{"au"})
			fmt.Println(container.Names)
			fmt.Println(processes.Titles)
			for _, process := range processes.Processes {
				fmt.Println(process)
			}
		}
		fmt.Println("")

	}
}
