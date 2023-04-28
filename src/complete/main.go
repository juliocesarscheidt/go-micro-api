package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	// "os/signal"
	"runtime"
	"strconv"
	"strings"
	// "syscall"
	"runtime/debug"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	logger                 = logrus.New()
	endpointCounterMetrics = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "gomicroapi",
			Name:      "http_request_count",
			Help:      "The total number of requests made to some endpoint",
		},
		[]string{"status", "method", "path"},
	)
	endpointDurationMetrics = promauto.NewHistogramVec(
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
	// Message is the global message received through environment variable
	Message string
	// Environment is the global environment received through environment variable
	Environment string
)

func init() {
	// logging config
	logger.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)
	// message and env variables from environment
	Message = getFromEnvOrDefaultAsString("MESSAGE", "Hello World")
	logger.Infof("Setting MESSAGE from ENV :: %s", Message)
	Environment = getFromEnvOrDefaultAsString("ENVIRONMENT", "development")
	logger.Infof("Setting ENVIRONMENT from ENV :: %s", Environment)
}

func initTracer() (*sdktrace.TracerProvider, error) {
	exporter, err := stdouttrace.New()
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("go-micro-api"),
			semconv.ServiceVersion("v1.0.0"),
			attribute.String("environment", Environment),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, err
}

func goroutineID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func getFromEnvOrDefaultAsString(envParam, defaultValue string) string {
	value := os.Getenv(envParam)
	if value == "" {
		value = defaultValue
	}
	return value
}

func extractIPFromRemoteAddr(remoteAddr string) string {
	addressParts := strings.Split(remoteAddr, ":")
	if len(addressParts) > 0 {
		return addressParts[0]
	}
	return ""
}

func buildJSONResponse(statusCode int, data interface{}) ([]byte, error) {
	var responseJSON = make(map[string]interface{})
	responseJSON["statusCode"] = statusCode
	responseJSON["data"] = data
	response, _ := json.Marshal(responseJSON)
	return []byte(string(response)), nil
}

func logRequest(statusCode int, path, host, method, ip, message string) {
	logger.WithFields(logrus.Fields{
		"status": statusCode,
		"method": method,
		"path":   path,
		"host":   host,
		"ip":     ip,
	}).Infof(message)

	debug.PrintStack()

	fmt.Printf("Goroutine ID :: %v\n", goroutineID())
	fmt.Printf("Num Goroutines :: %v\n", runtime.NumGoroutine())

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	oneMillion := math.Pow(10, 6)
	fmt.Printf("Memory Allocated: %.2f MBs | %.2f bytes\n", float64(memStats.Alloc)/oneMillion, float64(memStats.Alloc))
	fmt.Printf("Memory Obtained From Sys: %.2f MBs | %.2f bytes\n", float64(memStats.Sys)/oneMillion, float64(memStats.Sys))
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, req *http.Request) {
		statusCode := http.StatusOK
		// prometheus timer
		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(s float64) {
			endpointDurationMetrics.WithLabelValues(fmt.Sprint(statusCode), req.Method, req.URL.Path).Observe(s)
			// ms := s * 1_000     // milliseconds - 10^(-3)
			// us := s * 1_000_000 // microseconds - 10^(-6)
			// fmt.Printf("time seconds :: %v\n", s)
			// fmt.Printf("time milliseconds :: %v\n", ms)
			// fmt.Printf("time microseconds :: %v\n", us)
		}))
		// increment counter
		endpointCounterMetrics.WithLabelValues(fmt.Sprint(statusCode), req.Method, req.URL.Path).Inc()
		// call next route
		next.ServeHTTP(writter, req)
		// increment timer
		timer.ObserveDuration()
	})
}

func handleMessageRequestGet() http.HandlerFunc {
	return func(writter http.ResponseWriter, req *http.Request) {
		writter.Header().Set("Content-Type", "application/json")

		statusCode := http.StatusOK
		ctx := req.Context()
		// log request in other goroutine
		go logRequest(statusCode, req.URL.Path, req.Host, req.Method, extractIPFromRemoteAddr(req.RemoteAddr), Message)
		// otel tracing
		span := trace.SpanFromContext(ctx)
		span.AddEvent("trace", trace.WithAttributes(
			attribute.String("message", Message),
		))
		span.SetStatus(codes.Ok, "Ok")
		defer span.End()

		// inject some sleep time to simulate a job being done
		time.Sleep(time.Duration(100 * time.Millisecond))

		responseJSONBytes, _ := buildJSONResponse(statusCode, Message)
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func handleDefaultRequestGet(response string) http.HandlerFunc {
	return func(writter http.ResponseWriter, req *http.Request) {
		writter.Header().Set("Content-Type", "application/json")

		statusCode := http.StatusOK
		ctx := req.Context()
		// log request in other goroutine
		go logRequest(statusCode, req.URL.Path, req.Host, req.Method, extractIPFromRemoteAddr(req.RemoteAddr), response)
		// otel tracing
		span := trace.SpanFromContext(ctx)
		span.AddEvent("trace", trace.WithAttributes(
			attribute.String("message", Message),
		))
		span.SetStatus(codes.Ok, "Ok")
		defer span.End()

		responseJSONBytes, _ := buildJSONResponse(statusCode, response)
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func main() {
	logger.Infof("Goroutine ID :: %d", goroutineID())
	logger.Infof("Num Goroutines :: %d", runtime.NumGoroutine())

	ctx := context.Background()
	// create otel tracer
	tp, err := initTracer()
	if err != nil {
		logger.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// create router
	router := mux.NewRouter()
	// add metrics route
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	// add routes with otel tracing
	router.Handle("/api/v1/otel/message", otelhttp.NewHandler(http.HandlerFunc(
		handleMessageRequestGet()), "/api/v1/otel/message")).Methods("GET")
	router.Handle("/api/v1/otel/ping", otelhttp.NewHandler(http.HandlerFunc(
		handleDefaultRequestGet("Pong")), "/api/v1/otel/ping")).Methods("GET")
	// add routes with prometheus metrics
	subRouterProm := router.PathPrefix("/api/v1").Subrouter()
	subRouterProm.Use(prometheusMiddleware)
	subRouterProm.HandleFunc("/message", handleMessageRequestGet()).Methods("GET")
	subRouterProm.HandleFunc("/ping", handleDefaultRequestGet("Pong")).Methods("GET")
	subRouterProm.HandleFunc("/health/live", handleDefaultRequestGet("Alive")).Methods("GET")
	subRouterProm.HandleFunc("/health/ready", handleDefaultRequestGet("Ready")).Methods("GET")

	// start listening
	http.ListenAndServe(":9000", router)
}
