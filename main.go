package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
)

const (
	namespace = "Container"
)

var (
	processCPUUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "process_cpu_usage",
		Help:      "process_cpu_usage[%]",
	},
		[]string{"image", "container", "user", "pid", "tty", "stat", "start", "time", "command"},
	)
	processMemUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "process_memory_usage",
		Help:      "process_memory_usage[%]",
	},
		[]string{"image", "container", "user", "pid", "tty", "stat", "start", "time", "command"},
	)
	processGPUUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "process_gpu_usage",
		Help:      "process_gpu_usage[MiB]",
	},
		[]string{"image", "container", "user", "pid", "tty", "stat", "start", "time", "command"},
	)
	processVsz = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "process_vsz",
		Help:      "process_vsz[MB]",
	},
		[]string{"image", "container", "user", "pid", "tty", "stat", "start", "time", "command"},
	)
	processRss = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "process_rss",
		Help:      "process_rss[MB]",
	},
		[]string{"image", "container", "user", "pid", "tty", "stat", "start", "time", "command"},
	)
)

func init() {
	prometheus.MustRegister(processCPUUsage)
	prometheus.MustRegister(processMemUsage)
	prometheus.MustRegister(processGPUUsage)
	prometheus.MustRegister(processVsz)
	prometheus.MustRegister(processRss)
}

type myExporter struct {
	cli            *client.Client
	containers     []types.Container
	adminContainer []string
	gpuInfos       []string
	ctx            context.Context
}

func (e myExporter) setMetrics() {
	for _, container := range e.containers {
		if !isIgnorableContains(e.adminContainer, container.Names[0]) {
			processes, _ := e.cli.ContainerTop(e.ctx, container.ID, []string{"au"})
			for _, process := range processes.Processes {
				cpuUsage, _ := strconv.ParseFloat(process[2], 64)
				memUsage, _ := strconv.ParseFloat(process[3], 64)
				vszUsage, _ := strconv.ParseFloat(process[4], 64)
				rssUsage, _ := strconv.ParseFloat(process[5], 64)
				gpuUsage := e.getGpuUsage(process[1])

				processCPUUsage.WithLabelValues(getLabels(container, process)).Set(cpuUsage)
				processGPUUsage.WithLabelValues(getLabels(container, process)).Set(gpuUsage)
				processMemUsage.WithLabelValues(getLabels(container, process)).Set(memUsage)
				processVsz.WithLabelValues(getLabels(container, process)).Set(vszUsage)
				processRss.WithLabelValues(getLabels(container, process)).Set(rssUsage)

			}
		}
	}
}

func (e myExporter) getGpuUsage(targetPid string) float64 {
	gpuUsage := 0.0
	for _, info := range e.gpuInfos {
		if strings.Split(info, ", ")[0] == targetPid {
			gpuUsage, _ = strconv.ParseFloat(strings.Split(strings.Split(info, ", ")[1], " ")[0], 64)
		}
	}
	return gpuUsage
}

func getLabels(container types.Container, process []string) (string, string, string, string, string, string, string, string, string) {
	return container.Image, container.Names[0], process[0], process[1], process[6], process[7], process[8], process[9], process[10]
}

func readIgnoreContainer(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		return []string{}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	ignoreList := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		ignoreList = append(ignoreList, line)
	}

	return ignoreList
}

func isIgnorableContains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}

	return false
}

func main() {

	gpuInfos := strings.Split(GetGpuInfo(), "\n")

	flag.Parse()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	adminContainer := readIgnoreContainer("/go/conf/containerIgnore")

	go func() {
		for {
			containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
			if err != nil {
				panic(err)
			}
			exporter := myExporter{cli, containers, adminContainer, gpuInfos, ctx}
			exporter.setMetrics()
			time.Sleep(60 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
