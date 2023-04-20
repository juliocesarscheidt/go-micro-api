package main

import (
	// "context"
	"encoding/json"
	"fmt"
	// "math"
	"net/http"
	"os"
	// "os/signal"
	"runtime"
	"strconv"
	"strings"
	// "syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	Logger                 = logrus.New()
	EndpointCounterMetrics = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "gomicroapi",
			Name:      "http_request_count",
			Help:      "The total number of requests made to some endpoint",
		},
		[]string{"status", "method", "path"},
	)
	EndpointDurationMetrics = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "gomicroapi",
			Name:      "http_request_duration_seconds",
			Help:      "Latency of some endpoint requests in seconds",
			Buckets:   prometheus.DefBuckets,
			// Buckets: prometheus.LinearBuckets(0.01, 0.05, 10),
			// Buckets: prometheus.LinearBuckets(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10),
		},
		[]string{"status", "method", "path"},
	)
	Message     string
	Environment string
)

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
	// message and env variables from environment
	Message = GetFromEnvOrDefaultAsString("MESSAGE", "Hello World")
	Logger.Infof("Setting MESSAGE from ENV :: %s", Message)
	Environment = GetFromEnvOrDefaultAsString("ENVIRONMENT", "development")
	Logger.Infof("Setting ENVIRONMENT from ENV :: %s", Environment)
}

type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func GoroutineId() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
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

func BuildJSONResponse(statusCode int, data interface{}) ([]byte, error) {
	var responseJson = make(map[string]interface{})
	responseJson["statusCode"] = statusCode
	responseJson["data"] = data
	response, _ := json.Marshal(responseJson)
	return []byte(string(response)), nil
}

func LogRequest(statusCode int, path, host, method, ip, message string) {
	Logger.WithFields(logrus.Fields{
		"status": statusCode,
		"method": method,
		"path":   path,
		"host":   host,
		"ip":     ip,
	}).Infof(message)

	// var memStats runtime.MemStats
	// runtime.ReadMemStats(&memStats)
	// oneMillion := math.Pow(10, 6)
	// fmt.Printf("Memory Allocated: %.2f MBs\n", float64(memStats.Alloc)/oneMillion)
	// fmt.Printf("Memory Obtained From Sys: %.2f MBs\n", float64(memStats.Sys)/oneMillion)

	// fmt.Printf("Goroutine ID :: %v\n", GoroutineId())
	// fmt.Printf("Num Goroutines :: %v\n", runtime.NumGoroutine())
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, req *http.Request) {
		// create a wrapper for writter
		recorder := &StatusRecorder{
			ResponseWriter: writter,
			Status:         http.StatusOK,
		}
		// prometheus timer
		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(s float64) {
			EndpointDurationMetrics.WithLabelValues(fmt.Sprint(recorder.Status), req.Method, req.URL.Path).Observe(s)
			// ms := s * 1_000     // milliseconds - 10^(-3)
			// us := s * 1_000_000 // microseconds - 10^(-6)
			// fmt.Printf("time seconds :: %v\n", s)
			// fmt.Printf("time milliseconds :: %v\n", ms)
			// fmt.Printf("time microseconds :: %v\n", us)
		}))
		// call next route
		next.ServeHTTP(recorder, req)
		// retrieve status code
		statusCode := recorder.Status
		// increment counter
		EndpointCounterMetrics.WithLabelValues(fmt.Sprint(statusCode), req.Method, req.URL.Path).Inc()
		// increment timer
		timer.ObserveDuration()
	})
}

func HandleMessageRequestGet() http.HandlerFunc {
	return func(writter http.ResponseWriter, req *http.Request) {
		writter.Header().Set("Content-Type", "application/json")

		statusCode := http.StatusOK
		// log request in other goroutine
		go LogRequest(statusCode, req.URL.Path, req.Host, req.Method, ExtractIpFromRemoteAddr(req.RemoteAddr), Message)

		// inject some sleep time to simulate a job being done
		time.Sleep(time.Duration(100 * time.Millisecond))

		responseJSONBytes, _ := BuildJSONResponse(statusCode, Message)
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func HandleDefaultRequestGet(response string) http.HandlerFunc {
	return func(writter http.ResponseWriter, req *http.Request) {
		writter.Header().Set("Content-Type", "application/json")

		statusCode := http.StatusOK
		// log request in other goroutine
		go LogRequest(statusCode, req.URL.Path, req.Host, req.Method, ExtractIpFromRemoteAddr(req.RemoteAddr), response)

		responseJSONBytes, _ := BuildJSONResponse(statusCode, response)
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func main() {
	fmt.Printf("Goroutine ID :: %v\n", GoroutineId())
	fmt.Printf("Num Goroutines :: %v\n", runtime.NumGoroutine())

	// create router
	router := mux.NewRouter()
	// add metrics route
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// add routes with prometheus metrics
	subRouterProm := router.PathPrefix("/api/v1").Subrouter()
	subRouterProm.Use(prometheusMiddleware)
	subRouterProm.HandleFunc("/message", HandleMessageRequestGet()).Methods("GET")
	subRouterProm.HandleFunc("/ping", HandleDefaultRequestGet("Pong")).Methods("GET")
	subRouterProm.HandleFunc("/health/live", HandleDefaultRequestGet("Alive")).Methods("GET")
	subRouterProm.HandleFunc("/health/ready", HandleDefaultRequestGet("Ready")).Methods("GET")

	// start listening
	http.ListenAndServe(":9000", router)
}
