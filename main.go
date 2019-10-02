package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
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

type myCollector struct{}

var (
	processCPUUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "process_cpu_usage",
		Help:      "process_cpu_usage[%]",
	},
		[]string{"container", "user", "pid", "tty", "stat", "start", "time", "command"},
	)
	processMemUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "process_memory_usage",
		Help:      "process_memory_usage[%]",
	},
		[]string{"container", "user", "pid", "tty", "stat", "start", "time", "command"},
	)
	processVsz = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "process_vsz",
		Help:      "process_vsz[MB]",
	},
		[]string{"container", "user", "pid", "tty", "stat", "start", "time", "command"},
	)
	processRss = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "process_rss",
		Help:      "process_rss[MB]",
	},
		[]string{"container", "user", "pid", "tty", "stat", "start", "time", "command"},
	)
)

func init() {
	prometheus.MustRegister(processCPUUsage)
	prometheus.MustRegister(processMemUsage)
	prometheus.MustRegister(processVsz)
	prometheus.MustRegister(processRss)
}

func isArrayContains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}

	return false
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

func main() {
	flag.Parse()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))
	if err != nil {
		panic(err)
	}
	adminContainer := readIgnoreContainer("/go/conf/containerIgnore")
	go func() {
		for {
			containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
			if err != nil {
				panic(err)
			}

			for _, container := range containers {
				if !isArrayContains(adminContainer, container.Names[0]) {
					processes, _ := cli.ContainerTop(context.Background(), container.ID, []string{"au"})
					for _, process := range processes.Processes {
						cpuUsage, _ := strconv.ParseFloat(process[2], 64)
						memUsage, _ := strconv.ParseFloat(process[3], 64)
						vszUsage, _ := strconv.ParseFloat(process[4], 64)
						rssUsage, _ := strconv.ParseFloat(process[5], 64)

						processCPUUsage.WithLabelValues(container.Names[0], process[0], process[1], process[6], process[7], process[8], process[9], process[10]).Set(cpuUsage)
						processMemUsage.WithLabelValues(container.Names[0], process[0], process[1], process[6], process[7], process[8], process[9], process[10]).Set(memUsage)
						processVsz.WithLabelValues(container.Names[0], process[0], process[1], process[6], process[7], process[8], process[9], process[10]).Set(vszUsage)
						processRss.WithLabelValues(container.Names[0], process[0], process[1], process[6], process[7], process[8], process[9], process[10]).Set(rssUsage)

					}
				}
			}
			time.Sleep(60 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
