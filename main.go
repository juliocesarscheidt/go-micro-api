package main

import (
	"fmt"
	"os"
	"time"
	"net/http"
	"encoding/json"
	"os/signal"
	"context"

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
	response, err := json.Marshal(responseHTTP)
	if err != nil {
		return nil, err
	}
	return []byte(string(response)), nil
}

func returnHTTPResponse(statusCode int, message string) http.HandlerFunc {
	return func(writter http.ResponseWriter, req *http.Request) {
		responseJSONBytes, _ := buildJSONResponse(statusCode, message)

		log.WithFields(logrus.Fields{
			"host": req.Host,
			"path": req.URL.Path,
			"method": req.Method,
		}).Infof("")

		writter.Header().Set("Content-Type", "application/json")
		writter.WriteHeader(statusCode)
		writter.Write(responseJSONBytes)
	}
}

func ternary(statement bool, a, b interface{}) interface{} {
	if statement {
		return a
	}
	return b
}

func serve(address string, message string) {
	http.HandleFunc("/api/v1/", returnHTTPResponse(http.StatusOK, message))
	http.HandleFunc("/api/v1/healthcheck", returnHTTPResponse(http.StatusOK, "Healthy"))

	http.ListenAndServe(address, nil)
}

func main() {
	port := ternary(os.Getenv("API_PORT") != "", os.Getenv("API_PORT"), "9000").(string) // default 9000
	address := fmt.Sprintf(":%s", port)
	log.Infof("Using API_PORT :: %s", port)
	message := ternary(os.Getenv("MESSAGE") != "", os.Getenv("MESSAGE"), "Hello World").(string) // default Hello World
	log.Infof("Using MESSAGE :: %s", message)

	go func() {
		serve(address, message)
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	<-ctx.Done()
}
