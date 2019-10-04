package main

import (
	"bytes"
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// GetGpuInfo return result of nvidia-smi
func GetGpuInfo() string {
	gpuinfo := ""
	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	for _, container := range containers {
		if strings.Split(container.Image, ":")[0] == "nvidia/dcgm-exporter" {
			execConfig := types.ExecConfig{
				User:         "root",
				Privileged:   true,
				Tty:          true,
				AttachStdin:  true,
				AttachStderr: true,
				AttachStdout: true,
				Detach:       true,
				DetachKeys:   "",
				Env:          []string{""},
				WorkingDir:   "",
				Cmd:          strings.Split("nvidia-smi --query-compute-apps=pid,used_memory --format=csv,noheader", " "),
			}
			execStartCheck := types.ExecStartCheck{
				Detach: false,
				Tty:    false,
			}

			idRes, err := cli.ContainerExecCreate(context.Background(), container.ID, execConfig)
			if err != nil {
				panic(err)
			}

			res, err := cli.ContainerExecAttach(context.Background(), idRes.ID, execStartCheck)

			if err != nil {
				panic(err)
			}

			defer res.Close()

			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			_, err = stdcopy.StdCopy(stdout, stderr, res.Reader)
			if err != nil {
				panic(err)
			}
			gpuinfo = stdout.String()

		}

	}

	return gpuinfo

}
