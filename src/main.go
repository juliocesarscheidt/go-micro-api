package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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
	// prometheus config
	prometheus.MustRegister(EndpointCounterMetrics)
	prometheus.MustRegister(EndpointDurationMetrics)
	// message and env variables from environment
	Message = GetFromEnvOrDefaultAsString("MESSAGE", "Hello World")
	Logger.Infof("Setting MESSAGE from ENV :: %s", Message)
	Environment = GetFromEnvOrDefaultAsString("ENVIRONMENT", "development")
	Logger.Infof("Setting ENVIRONMENT from ENV :: %s", Environment)
}

func initTracer() (*sdktrace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
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

// ConfigurationDto is the content of request for configuration update
type ConfigurationDto struct {
	Message string `json:"message"`
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
		ctx := req.Context()
		span := trace.SpanFromContext(ctx)
		span.AddEvent("trace", trace.WithAttributes(
			attribute.String("message", Message),
		))
		defer span.End()
		defer func() {
			LogRequestMetrics(statusCode, req.URL.Path, req.Host, req.Method, ExtractIpFromRemoteAddr(req.RemoteAddr), Message)
		}()
		if req.Method != "GET" {
			statusCode = http.StatusMethodNotAllowed
			span.SetStatus(codes.Error, "Method Not Allowed")
			writter.WriteHeader(statusCode)
			return
		}
		responseJSONBytes, _ := BuildJSONResponse(statusCode, Message)
		span.SetStatus(codes.Ok, "Ok")
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func HandleDefaultRequestGet(response interface{}) http.HandlerFunc {
	return func(writter http.ResponseWriter, req *http.Request) {
		writter.Header().Set("Content-Type", "application/json")
		statusCode := http.StatusOK
		ctx := req.Context()
		span := trace.SpanFromContext(ctx)
		span.AddEvent("trace", trace.WithAttributes(
			attribute.String("message", response.(string)),
		))
		defer span.End()
		defer func() {
			LogRequestMetrics(statusCode, req.URL.Path, req.Host, req.Method, ExtractIpFromRemoteAddr(req.RemoteAddr), response)
		}()
		if req.Method != "GET" {
			statusCode = http.StatusMethodNotAllowed
			span.SetStatus(codes.Error, "Method Not Allowed")
			writter.WriteHeader(statusCode)
			return
		}
		responseJSONBytes, _ := BuildJSONResponse(statusCode, response)
		span.SetStatus(codes.Ok, "Ok")
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func main() {
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
	// add routes
	http.Handle("/api/v1/message", otelhttp.NewHandler(http.HandlerFunc(HandleMessageRequestGet()), "/api/v1/message"))
	http.Handle("/api/v1/ping", otelhttp.NewHandler(http.HandlerFunc(HandleDefaultRequestGet("Pong")), "/api/v1/ping"))
	http.Handle("/api/v1/health/live", otelhttp.NewHandler(http.HandlerFunc(HandleDefaultRequestGet("Alive")), "/api/v1/health/live"))
	http.Handle("/api/v1/health/ready", otelhttp.NewHandler(http.HandlerFunc(HandleDefaultRequestGet("Ready")), "/api/v1/health/ready"))
	// prometheus metrics
	http.Handle("/metrics", promhttp.Handler())
	// start listening inside other goroutine
	go func() {
		http.ListenAndServe(":9000", nil)
	}()
	// cancel on SIGTERM or SIGINT
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	<-ctx.Done()
}
