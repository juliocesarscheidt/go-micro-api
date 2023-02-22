package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	logrus "github.com/sirupsen/logrus"
)

var (
	Log                    = logrus.New()
	EndpointCounterMetrics = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "api",
			Name:      "http_request_count",
			Help:      "The total number of requests made to some endpoint",
		},
		[]string{"status", "method", "path"},
	)
	EndpointDurationMetrics = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "api",
			Name:      "http_request_duration_seconds",
			Help:      "Latency of some endpoint requests in seconds",
		},
		[]string{"status", "method", "path"},
	)
)

func init() {
	// logging config
	Log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	Log.SetOutput(os.Stdout)
	Log.SetLevel(logrus.DebugLevel)
	// prometheus config
	prometheus.MustRegister(EndpointCounterMetrics)
	prometheus.MustRegister(EndpointDurationMetrics)
}

func putEndpointMetrics(path, method, status string) {
	defer func() {
		EndpointCounterMetrics.WithLabelValues(status, method, path).Inc()
	}()
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(_time float64) {
		EndpointDurationMetrics.WithLabelValues(status, method, path).Observe(_time)
	}))
	defer func() {
		timer.ObserveDuration()
	}()
}

func buildJSONResponse(statusCode int, message interface{}) ([]byte, error) {
	var responseHTTP = make(map[string]interface{})
	responseHTTP["statusCode"] = statusCode
	responseHTTP["data"] = message
	response, _ := json.Marshal(responseHTTP)
	return []byte(string(response)), nil
}

func returnHTTPResponse(statusCode int, message interface{}) http.HandlerFunc {
	return func(writter http.ResponseWriter, req *http.Request) {
		writter.Header().Set("Content-Type", "application/json")
		responseJSONBytes, _ := buildJSONResponse(statusCode, message)
		putEndpointMetrics(req.URL.Path, req.Method, fmt.Sprint(statusCode))
		ip := strings.Split(req.RemoteAddr, ":")[0]
		Log.WithFields(logrus.Fields{
			"host":   req.Host,
			"ip":     ip,
			"path":   req.URL.Path,
			"method": req.Method,
		}).Infof("")
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func getFromEnvOrDefaultAsString(envParam, defaultValue string) string {
	value := os.Getenv(envParam)
	if value == "" {
		value = defaultValue
	}
	return value
}

func main() {
	message := getFromEnvOrDefaultAsString("MESSAGE", "Hello World")
	Log.Infof("Using var MESSAGE from env :: %s", message)
	// add routes
	http.HandleFunc("/api/v1/message", returnHTTPResponse(http.StatusOK, message))
	http.HandleFunc("/api/v1/ping", returnHTTPResponse(http.StatusOK, "Pong"))
	http.HandleFunc("/api/v1/health/live", returnHTTPResponse(http.StatusOK, "Alive"))
	http.HandleFunc("/api/v1/health/ready", returnHTTPResponse(http.StatusOK, "Ready"))
	http.Handle("/metrics", promhttp.Handler())
	// start listening inside other goroutine
	go func() {
		http.ListenAndServe(":9000", nil)
	}()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	<-ctx.Done()
}
