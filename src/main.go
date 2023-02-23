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
	Logger                 = logrus.New()
	EndpointCounterMetrics = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "gomicroapi",
			Name:      "http_request_count",
			Help:      "The total number of requests made to some endpoint",
		},
		[]string{"status", "method", "path"},
	)
	EndpointDurationMetrics = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "gomicroapi",
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
	Logger.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	Logger.SetOutput(os.Stdout)
	Logger.SetLevel(logrus.DebugLevel)
	// prometheus config
	prometheus.MustRegister(EndpointCounterMetrics)
	prometheus.MustRegister(EndpointDurationMetrics)
	// message variable from environment
	Message = GetFromEnvOrDefaultAsString("MESSAGE", "Hello World")
	Logger.Infof("Setting MESSAGE from ENV :: %s", Message)
}

func PutRequestMetrics(path, method, status string) {
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

func GetFromEnvOrDefaultAsString(envParam, defaultValue string) string {
	value := os.Getenv(envParam)
	if value == "" {
		value = defaultValue
	}
	return value
}

func ExtractIpFromRemoteAddr(remoteAddr string) string {
	addressParts := strings.Split(remoteAddr, ":")
	if len(addressParts) > 0 {
		return addressParts[0]
	}
	return ""
}

func LogRequestMetrics(statusCode int, path, host, method, ip string, message interface{}) {
	Logger.WithFields(logrus.Fields{
		"status": statusCode,
		"method": method,
		"path":   path,
		"host":   host,
		"ip":     ip,
	}).Infof(fmt.Sprint(message))
	PutRequestMetrics(path, method, fmt.Sprint(statusCode))
}

func BuildJSONResponse(statusCode int, message interface{}) ([]byte, error) {
	var responseHTTP = make(map[string]interface{})
	responseHTTP["statusCode"] = statusCode
	responseHTTP["data"] = message
	response, _ := json.Marshal(responseHTTP)
	return []byte(string(response)), nil
}

func HandleMessageRequestGet() http.HandlerFunc {
	return func(writter http.ResponseWriter, req *http.Request) {
		writter.Header().Set("Content-Type", "application/json")
		statusCode := http.StatusOK
		defer func() {
			LogRequestMetrics(statusCode, req.URL.Path, req.Host, req.Method, ExtractIpFromRemoteAddr(req.RemoteAddr), Message)
		}()
		if req.Method != "GET" {
			statusCode = http.StatusMethodNotAllowed
			writter.WriteHeader(statusCode)
			return
		}
		responseJSONBytes, _ := BuildJSONResponse(statusCode, Message)
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func HandleDefaultRequestGet(response interface{}) http.HandlerFunc {
	return func(writter http.ResponseWriter, req *http.Request) {
		writter.Header().Set("Content-Type", "application/json")
		statusCode := http.StatusOK
		defer func() {
			LogRequestMetrics(statusCode, req.URL.Path, req.Host, req.Method, ExtractIpFromRemoteAddr(req.RemoteAddr), Message)
		}()
		if req.Method != "GET" {
			statusCode = http.StatusMethodNotAllowed
			writter.WriteHeader(statusCode)
			return
		}
		responseJSONBytes, _ := BuildJSONResponse(statusCode, response)
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func HandleConfigurationRequestPut() http.HandlerFunc {
	return func(writter http.ResponseWriter, req *http.Request) {
		writter.Header().Set("Content-Type", "application/json")
		statusCode := http.StatusAccepted
		defer func() {
			LogRequestMetrics(statusCode, req.URL.Path, req.Host, req.Method, ExtractIpFromRemoteAddr(req.RemoteAddr), Message)
		}()
		if req.Method != "PUT" {
			statusCode = http.StatusMethodNotAllowed
			writter.WriteHeader(statusCode)
			return
		}
		// max body size accepted is 1024 bytes (1 KB), otherwise it will return bad request
		limitedReader := &io.LimitedReader{R: req.Body, N: 1024}
		bodyCopy := new(bytes.Buffer)
		_, err := io.Copy(bodyCopy, limitedReader)
		if err != nil {
			Logger.Errorf("Error :: %v", err)
			statusCode = http.StatusInternalServerError
			writter.WriteHeader(statusCode)
			return
		}
		payload := ConfigurationDto{}
		bodyData := bodyCopy.Bytes()
		req.Body = ioutil.NopCloser(bytes.NewReader(bodyData))
		json.Unmarshal(bodyData, &payload)
		if payload.Message == "" {
			statusCode = http.StatusBadRequest
			writter.WriteHeader(statusCode)
			return
		}
		Message = strings.Trim(payload.Message, " ")
		Logger.Infof("Setting MESSAGE from CONFIGURATION :: %s", Message)
		responseJSONBytes, _ := BuildJSONResponse(statusCode, nil)
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func main() {
	// add routes
	http.HandleFunc("/api/v1/message", HandleMessageRequestGet())
	http.HandleFunc("/api/v1/configuration", HandleConfigurationRequestPut())
	http.HandleFunc("/api/v1/ping", HandleDefaultRequestGet("Pong"))
	http.HandleFunc("/api/v1/health/live", HandleDefaultRequestGet("Alive"))
	http.HandleFunc("/api/v1/health/ready", HandleDefaultRequestGet("Ready"))
	http.Handle("/metrics", promhttp.Handler())
	// start listening inside other goroutine
	go func() {
		http.ListenAndServe(":9000", nil)
	}()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	<-ctx.Done()
}
