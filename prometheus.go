package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	prometheusEnabled = false

	prometheusRequestsTotal      prometheus.Counter
	prometheusErrorsTotal        *prometheus.CounterVec
	prometheusRequestDuration    prometheus.Histogram
	prometheusDownloadDuration   prometheus.Histogram
	prometheusProcessingDuration prometheus.Histogram
	prometheusBufferSize         *prometheus.HistogramVec
	prometheusBufferDefaultSize  *prometheus.GaugeVec
	prometheusBufferMaxSize      *prometheus.GaugeVec
)

func initPrometheus() {
	if len(conf.PrometheusBind) == 0 {
		return
	}

	prometheusRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "requests_total",
		Help: "A counter of the total number of HTTP requests imgproxy processed.",
	})

	prometheusErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "errors_total",
		Help: "A counter of the occured errors separated by type.",
	}, []string{"type"})

	prometheusRequestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "request_duration_seconds",
		Help: "A histogram of the response latency.",
	})

	prometheusDownloadDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "download_duration_seconds",
		Help: "A histogram of the source image downloading latency.",
	})

	prometheusProcessingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "processing_duration_seconds",
		Help: "A histogram of the image processing latency.",
	})

	prometheusBufferSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "buffer_size_megabytes",
		Help: "A histogram of the buffer size in megabytes.",
	}, []string{"type"})

	prometheusBufferDefaultSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "buffer_default_size_megabytes",
		Help: "A gauge of the buffer default size in megabytes.",
	}, []string{"type"})

	prometheusBufferMaxSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "buffer_max_size_megabytes",
		Help: "A gauge of the buffer max size in megabytes.",
	}, []string{"type"})

	prometheus.MustRegister(
		prometheusRequestsTotal,
		prometheusErrorsTotal,
		prometheusRequestDuration,
		prometheusDownloadDuration,
		prometheusProcessingDuration,
		prometheusBufferSize,
		prometheusBufferDefaultSize,
		prometheusBufferMaxSize,
	)

	prometheusEnabled = true

	s := http.Server{
		Addr:    conf.PrometheusBind,
		Handler: promhttp.Handler(),
	}

	go func() {
		logNotice("Starting Prometheus server at %s\n", s.Addr)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logFatal(err.Error())
		}
	}()
}

func startPrometheusDuration(m prometheus.Histogram) func() {
	t := time.Now()
	return func() {
		m.Observe(time.Since(t).Seconds())
	}
}

func incrementPrometheusErrorsTotal(t string) {
	prometheusErrorsTotal.With(prometheus.Labels{"type": t}).Inc()
}

func observePrometheusBufferSize(t string, cap int) {
	size := float64(cap) / 1024.0 / 1024.0
	prometheusBufferSize.With(prometheus.Labels{"type": t}).Observe(size)
}

func setPrometheusBufferDefaultSize(t string, cap int) {
	size := float64(cap) / 1024.0 / 1024.0
	prometheusBufferDefaultSize.With(prometheus.Labels{"type": t}).Set(size)
}

func setPrometheusBufferMaxSize(t string, cap int) {
	size := float64(cap) / 1024.0 / 1024.0
	prometheusBufferMaxSize.With(prometheus.Labels{"type": t}).Set(size)
}
