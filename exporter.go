package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "fahcollector"
	labelPod  = "pod"
	labelNode = "node"
)

type Exporter struct {
	SkipTLSValidation bool
	FAHCollectorURL   string

	progressDesc        *prometheus.Desc
	stateDownloadDesc   *prometheus.Desc
	stateRunningDesc    *prometheus.Desc
	stateReadyDesc      *prometheus.Desc
	stateUnexpectedDesc *prometheus.Desc
}

var _ prometheus.Collector = (*Exporter)(nil)

func NewExporter(url string, insecure bool) *Exporter {
	var e Exporter
	e.FAHCollectorURL = url
	e.SkipTLSValidation = insecure

	e.progressDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "progress", "percent"),
		"FAH current progress percentage",
		[]string{
			labelPod, labelNode,
		},
		nil,
	)
	descState := func(name, desc string) *prometheus.Desc {
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "state", name),
			desc,
			[]string{
				labelPod, labelNode,
			},
			nil,
		)
	}
	e.stateDownloadDesc = descState("download", "FAH current state is DOWNLOAD")
	e.stateRunningDesc = descState("running", "FAH current state is RUNNING")
	e.stateReadyDesc = descState("ready", "FAH current state is READY")
	e.stateUnexpectedDesc = descState("unexpected", "FAH returned an unexpected state")

	return &e
}

func (e Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.progressDesc
	ch <- e.stateDownloadDesc
	ch <- e.stateRunningDesc
	ch <- e.stateReadyDesc
	ch <- e.stateUnexpectedDesc
}

type FAHCResp struct {
	Hostname    string `json:"sc_hostname"`
	Nodename    string `json:"sc_nodename"`
	PercentDone string `json:"percentdone"`
	State       string `json:"state"`
}

func (e Exporter) Collect(ch chan<- prometheus.Metric) {
	// get fah-collector JSON
	if e.SkipTLSValidation {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	resp, err := http.Get(e.FAHCollectorURL)
	if err != nil {
		log.Printf("Collect: http.Get err=%v", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Collect: status not ok code=%d", resp.StatusCode)
		return
	}
	j := map[string]FAHCResp{}
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		log.Printf("Collect: parse json err=%v", err)
		return
	}
	resp.Body.Close()

	for k, v := range j {
		pd, err := strconv.ParseFloat(strings.TrimSuffix(v.PercentDone, "%"), 64)
		if err != nil {
			log.Printf("Collect: iterating key=%v err=%v", k, err)
			continue
		}
		// log.Printf("Collect: do v=%+v", v)
		e.collectInstance(ch, v.Hostname, v.Nodename, pd, v.State)
	}
}

func (e Exporter) collectInstance(ch chan<- prometheus.Metric, labelPod, labelNode string, progress float64, state string) {
	// collect progress
	ch <- prometheus.MustNewConstMetric(
		e.progressDesc,
		prometheus.GaugeValue,
		progress,
		labelPod, labelNode,
	)
	// collect state
	var (
		isDownload   float64
		isRunning    float64
		isReady      float64
		isUnexpected float64
	)
	switch state {
	case "DOWNLOAD":
		isDownload = 1
	case "RUNNING":
		isRunning = 1
	case "READY":
		isReady = 1
	default:
		isUnexpected = 1
	}
	ch <- prometheus.MustNewConstMetric(
		e.stateDownloadDesc,
		prometheus.CounterValue,
		isDownload,
		labelPod, labelNode,
	)
	ch <- prometheus.MustNewConstMetric(
		e.stateRunningDesc,
		prometheus.CounterValue,
		isRunning,
		labelPod, labelNode,
	)
	ch <- prometheus.MustNewConstMetric(
		e.stateReadyDesc,
		prometheus.CounterValue,
		isReady,
		labelPod, labelNode,
	)
	ch <- prometheus.MustNewConstMetric(
		e.stateUnexpectedDesc,
		prometheus.CounterValue,
		isUnexpected,
		labelPod, labelNode,
	)
}
