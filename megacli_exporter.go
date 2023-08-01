package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

const (
	exporter     = "megacli_exporter"
	default_port = 9422
)

func runMegaCliSasStatus() []byte {
	out, err := exec.Command("megaclisas-status").Output()

	if err != nil {
		level.Error(logger).Log("err", err, "msg", out)
	}

	return out
}

func getSection(raw []byte, section string) []map[string]string {
	// split by newline, look for "-- Controller information --", next line is keys in hashmap, then values on lines bewlow that until empty line
	// create array of raid sets
	var controllersInformation []map[string]string

	// split by newline
	lines := strings.Split(string(raw), "\n")

	headerLineFound := false
	keys := []string{}

	// find line with "-- section --"
	for i, line := range lines {
		if !headerLineFound && line == "-- "+section+" --" {
			headerLineFound = true

			// add next line to keys, split by "|"
			for _, column := range strings.Split(lines[i+1], "|") {
				key := strings.ReplaceAll(column, "-", "")
				key = strings.ReplaceAll(key, " ", "")
				key = strings.ReplaceAll(key, "/", "")
				key = strings.ToLower(key)

				keys = append(keys, key)
			}
		}
		if headerLineFound && line == "" {
			break
		}
		if headerLineFound && !strings.HasPrefix(line, "--") {
			// line is values
			values := strings.Split(line, "|")

			// create hashmap
			controllerInformation := make(map[string]string)

			// add keys and values to hashmap
			for i, key := range keys {
				// trim and remove spaces
				key = strings.TrimSpace(key)
				values[i] = strings.TrimSpace(values[i])

				controllerInformation[key] = values[i]
			}

			// add hashmap to array
			controllersInformation = append(controllersInformation, controllerInformation)
		}

	}

	return controllersInformation
}

func getControllerInformation(raw []byte) []map[string]string {
	return getSection(raw, "Controller information")
}

func getArrayInformation(raw []byte) []map[string]string {
	return getSection(raw, "Array information")
}

func getDiskInformation(raw []byte) []map[string]string {
	return getSection(raw, "Disk information")
}

func recordMetrics() {
	// set all controller information as constant metrics
	for _, controllerInformation := range megaRaidControllersInformation {
		controllerInformation.Set(1)
	}
	megaRaidExporterCollectUp.Set(0)

	// create new gauge for each array, and each disk
	// var arrayGauges []prometheus.Gauge
	// var diskGauges []prometheus.Gauge

	// create new gauge for each raid set
	go func() {
		for {
			// run cli
			raw := runMegaCliSasStatus()

			// parse it
			array_info := getArrayInformation(raw)
			disk_info := getDiskInformation(raw)

			// if same amount of arrays, then just update the labels if changed
			fmt.Println(array_info)
			fmt.Println(disk_info)

			time.Sleep(5 * time.Second)
		}
	}()

}

func registerControllerInformationMetrics() []prometheus.Gauge {
	controllersInformation := getControllerInformation(runMegaCliSasStatus())
	controllersInformationGauges := []prometheus.Gauge{}

	for _, controllerInformation := range controllersInformation {
		megaRaidControllerInformation := promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "megacli_controller_information",
			Help:        "Constant metric with value 1 labeled with info about MegaRAID controller.",
			ConstLabels: controllerInformation,
		})

		controllersInformationGauges = append(controllersInformationGauges, megaRaidControllerInformation)
	}

	return controllersInformationGauges
}

var (
	logger = promlog.New(&promlog.Config{})

	megaRaidControllersInformation = registerControllerInformationMetrics()
	megaRaidExporterCollectUp      = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "megacli_exporter_up",
		Help: "'0' if a scrape of the MegaRAID CLI was successful, '1' otherwise.",
	})
)

func main() {
	toolkitFlags := webflag.AddFlags(kingpin.CommandLine, ":"+fmt.Sprint(default_port))

	kingpin.Version(version.Print(exporter))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	prometheus.Register(version.NewCollector(exporter))

	recordMetrics()

	level.Info(logger).Log("msg", "Starting megacli_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	http.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{}
	if err := web.ListenAndServe(srv, toolkitFlags, logger); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}
