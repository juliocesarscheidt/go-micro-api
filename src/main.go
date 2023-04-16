package main

import (
	"context"
	"encoding/json"
	"fmt"
	// "math"
	"net/http"
	"os"
	"os/signal"
	// "runtime"
	"strings"
	"syscall"
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

type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
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
	// memAllocated := memStats.Alloc
	// memObtainedFromSys := memStats.Sys
	// fmt.Printf("Memory Allocated: %.2f MBs | %.2f bytes\n", float64(memAllocated)/math.Pow(10, 6), float64(memAllocated))
	// fmt.Printf("Memory Obtained From Sys: %.2f MBs | %.2f bytes\n", float64(memObtainedFromSys)/math.Pow(10, 6), float64(memObtainedFromSys))
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
		ctx := req.Context()
		// log request
		LogRequest(statusCode, req.URL.Path, req.Host, req.Method, ExtractIpFromRemoteAddr(req.RemoteAddr), Message)
		// otel tracing
		span := trace.SpanFromContext(ctx)
		span.AddEvent("trace", trace.WithAttributes(
			attribute.String("message", Message),
		))
		span.SetStatus(codes.Ok, "Ok")
		defer span.End()
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
		ctx := req.Context()
		// log request
		LogRequest(statusCode, req.URL.Path, req.Host, req.Method, ExtractIpFromRemoteAddr(req.RemoteAddr), response)
		// otel tracing
		span := trace.SpanFromContext(ctx)
		span.AddEvent("trace", trace.WithAttributes(
			attribute.String("message", Message),
		))
		span.SetStatus(codes.Ok, "Ok")
		defer span.End()

		responseJSONBytes, _ := BuildJSONResponse(statusCode, response)
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func main() {
	// runtime.GOMAXPROCS(1)

	ctx := context.Background()
	// create otel tracer
	tp, err := initTracer()
	if err != nil {
		Logger.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			Logger.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// create router
	router := mux.NewRouter()
	// add metrics route
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	// add routes with otel tracing
	router.Handle("/api/v1/otel/message", otelhttp.NewHandler(http.HandlerFunc(
		HandleMessageRequestGet()), "/api/v1/otel/message")).Methods("GET")
	router.Handle("/api/v1/otel/ping", otelhttp.NewHandler(http.HandlerFunc(
		HandleDefaultRequestGet("Pong")), "/api/v1/otel/ping")).Methods("GET")
	// add routes with prometheus metrics
	subRouterProm := router.PathPrefix("/api/v1").Subrouter()
	subRouterProm.Use(prometheusMiddleware)
	subRouterProm.HandleFunc("/message", HandleMessageRequestGet()).Methods("GET")
	subRouterProm.HandleFunc("/ping", HandleDefaultRequestGet("Pong")).Methods("GET")
	subRouterProm.HandleFunc("/health/live", HandleDefaultRequestGet("Alive")).Methods("GET")
	subRouterProm.HandleFunc("/health/ready", HandleDefaultRequestGet("Ready")).Methods("GET")
	// start listening inside other goroutine
	go func() {
		http.ListenAndServe(":9000", router)
	}()
	// cancel on SIGTERM or SIGINT
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	<-ctx.Done()
}
