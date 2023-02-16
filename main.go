package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
	"html"
	"net/http"
	"time"
)

var (
	COUNTER = promauto.NewCounter(prometheus.CounterOpts{
		Name: "hello_world_total",
		Help: "Hello World requested",
	})

	GAUGE = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hello_world_connection",
		Help: "Number of /gauge in progress",
	})

	SUMMARY = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "hello_world_latency_seconds",
		Help: "Latency Time for a request /summary",
	})

	HISTOGRAM = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "hello_world_latency_histogram",
		Help:    "A histogram of Latency Time for a request /histogram",
		Buckets: prometheus.LinearBuckets(0.1, 0.1, 10),
	})
)

func index(w http.ResponseWriter, r *http.Request) {
	COUNTER.Inc()
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func gauge(w http.ResponseWriter, r *http.Request) {
	GAUGE.Inc()
	defer GAUGE.Dec()
	time.Sleep(10 * time.Second)
	fmt.Fprintf(w, "Gauge, %q", html.EscapeString(r.URL.Path))
}

func summary(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	time.Sleep(1 * time.Second)
	defer SUMMARY.Observe(float64(time.Now().Sub(start)))
	fmt.Fprintf(w, "Summary, %q", html.EscapeString(r.URL.Path))
}

func histogram(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	time.Sleep(1 * time.Second)
	defer HISTOGRAM.Observe(float64(time.Now().Sub(start)))
	fmt.Fprintf(w, "Histogram, %q", html.EscapeString(r.URL.Path))
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/gauge", gauge)
	http.HandleFunc("/summary", summary)
	http.HandleFunc("/histogram", histogram)

	// CPU 사용량 메트릭 정의
	cpuUsage := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_usage",
		Help: "CPU usage",
	})

	// 메트릭 등록
	prometheus.MustRegister(cpuUsage)

	// 1초마다 CPU 사용량을 갱신하고, 디스크용량 메트릭도 갱신
	go func() {
		for {
			// 시스템 CPU 사용량 수집
			usage := float64(0)
			if percent, err := getCPUPercent(); err != nil {
				fmt.Printf("Error getting CPU usage: %v\n", err)
			} else {
				usage = percent
				fmt.Printf("CPU usage: %.2f%%\n", usage)

				// CPU 사용량 메트릭 갱신
				cpuUsage.Set(usage)
			}

			time.Sleep(1 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

func getCPUPercent() (float64, error) {
	cpuTimes, err := cpu.Times(false)
	if err != nil {
		return 0, err
	}

	idleTime := cpuTimes[0].Idle
	totalTime := cpuTimes[0].Total()

	time.Sleep(1 * time.Second)

	cpuTimes, err = cpu.Times(false)
	if err != nil {
		return 0, err
	}

	idleTime = cpuTimes[0].Idle - idleTime
	totalTime = cpuTimes[0].Total() - totalTime

	usage := 100 * (1 - (idleTime / totalTime))
	return usage, nil
}
