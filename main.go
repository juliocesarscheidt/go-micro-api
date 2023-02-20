package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	logrus "github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)
}

func buildJSONResponse(statusCode int, message string) ([]byte, error) {
	var responseHTTP = make(map[string]interface{})
	responseHTTP["statusCode"] = statusCode
	responseHTTP["data"] = message
	response, _ := json.Marshal(responseHTTP)
	return []byte(string(response)), nil
}

func returnHTTPResponse(statusCode int, message string) http.HandlerFunc {
	return func(writter http.ResponseWriter, req *http.Request) {
		writter.Header().Set("Content-Type", "application/json")
		responseJSONBytes, _ := buildJSONResponse(statusCode, message)
		remoteIp := strings.Split(req.RemoteAddr, ":")[0]
		log.WithFields(logrus.Fields{
			"host":   req.Host,
			"ip":     remoteIp,
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

func serve(address string, message string) {
	http.HandleFunc("/api/v1/message", returnHTTPResponse(http.StatusOK, message))
	http.HandleFunc("/api/v1/health/live", returnHTTPResponse(http.StatusOK, "Alive"))
	http.HandleFunc("/api/v1/health/ready", returnHTTPResponse(http.StatusOK, "Ready"))
	http.ListenAndServe(address, nil)
}

func main() {
	message := getFromEnvOrDefaultAsString("MESSAGE", "Hello World")
	log.Infof("Using var MESSAGE from env :: %s", message)
	go func() {
		serve(":9000", message)
	}()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	<-ctx.Done()
}
