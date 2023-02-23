package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	Message string
)

// ConfigurationDto is the content of request for configuration update
type ConfigurationDto struct {
	Message string `json:"message"`
}

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
	// message variable from environment
	Message = getFromEnvOrDefaultAsString("MESSAGE", "Hello World")
	Log.Infof("Setting MESSAGE from ENV :: %s", Message)
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

func getFromEnvOrDefaultAsString(envParam, defaultValue string) string {
	value := os.Getenv(envParam)
	if value == "" {
		value = defaultValue
	}
	return value
}

func buildJSONResponse(statusCode int, message interface{}) ([]byte, error) {
	var responseHTTP = make(map[string]interface{})
	responseHTTP["statusCode"] = statusCode
	responseHTTP["data"] = message
	response, _ := json.Marshal(responseHTTP)
	return []byte(string(response)), nil
}

func buildHTTPResponse(statusCode int, message interface{}, path, host, method, remoteAddr string) []byte {
	responseJSONBytes, _ := buildJSONResponse(statusCode, message)
	putEndpointMetrics(path, method, fmt.Sprint(statusCode))
	ip := strings.Split(remoteAddr, ":")[0]
	Log.WithFields(logrus.Fields{
		"host":   host,
		"ip":     ip,
		"path":   path,
		"method": method,
	}).Infof(fmt.Sprint(message))
	return responseJSONBytes
}

func handleMessageRequestGet(statusCode int) http.HandlerFunc {
	return func(writter http.ResponseWriter, req *http.Request) {
		writter.Header().Set("Content-Type", "application/json")
		if req.Method != "GET" {
			writter.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		responseJSONBytes := buildHTTPResponse(statusCode, Message, req.URL.Path, req.Host, req.Method, req.RemoteAddr)
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func handleDefaultRequestGet(statusCode int, message interface{}) http.HandlerFunc {
	return func(writter http.ResponseWriter, req *http.Request) {
		writter.Header().Set("Content-Type", "application/json")
		if req.Method != "GET" {
			writter.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		responseJSONBytes := buildHTTPResponse(statusCode, message, req.URL.Path, req.Host, req.Method, req.RemoteAddr)
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func handleConfigurationRequestPut(statusCode int, message interface{}) http.HandlerFunc {
	return func(writter http.ResponseWriter, req *http.Request) {
		writter.Header().Set("Content-Type", "application/json")
		if req.Method != "PUT" {
			writter.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// max body size accepted is 1024 bytes (1 KB), otherwise it will return bad request
		limitedReader := &io.LimitedReader{R: req.Body, N: 1024}
		bodyCopy := new(bytes.Buffer)
		_, err := io.Copy(bodyCopy, limitedReader)
		if err != nil {
			Log.Errorf("Error :: %v", err)
			writter.WriteHeader(http.StatusInternalServerError)
			return
		}
		payload := ConfigurationDto{}
		bodyData := bodyCopy.Bytes()
		req.Body = ioutil.NopCloser(bytes.NewReader(bodyData))
		json.Unmarshal(bodyData, &payload)
		if payload.Message == "" {
			writter.WriteHeader(http.StatusBadRequest)
			return
		}
		Message = payload.Message
		Log.Infof("Setting MESSAGE from CONFIGURATION :: %s", Message)
		responseJSONBytes := buildHTTPResponse(statusCode, message, req.URL.Path, req.Host, req.Method, req.RemoteAddr)
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func main() {
	// add routes
	http.HandleFunc("/api/v1/message", handleMessageRequestGet(http.StatusOK))
	http.HandleFunc("/api/v1/configuration", handleConfigurationRequestPut(http.StatusAccepted, nil))
	http.HandleFunc("/api/v1/ping", handleDefaultRequestGet(http.StatusOK, "Pong"))
	http.HandleFunc("/api/v1/health/live", handleDefaultRequestGet(http.StatusOK, "Alive"))
	http.HandleFunc("/api/v1/health/ready", handleDefaultRequestGet(http.StatusOK, "Ready"))
	http.Handle("/metrics", promhttp.Handler())
	// start listening inside other goroutine
	go func() {
		http.ListenAndServe(":9000", nil)
	}()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	<-ctx.Done()
}
