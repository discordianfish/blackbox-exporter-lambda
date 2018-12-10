package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/blackbox_exporter/config"
	"github.com/prometheus/blackbox_exporter/prober"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	yaml "gopkg.in/yaml.v2"
)

var (
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	authToken = os.Getenv("AUTH_TOKEN")

	errInvalidToken  = errors.New("Invalid auth token")
	errInvalidPath   = errors.New("Invalid path")
	errInvalidProber = errors.New("Invalid prober")

	responseConfigInvalid = events.APIGatewayProxyResponse{
		StatusCode: http.StatusBadRequest,
		Body:       "Invalid config",
	}
)

func requireAuth(header string) (events.APIGatewayProxyResponse, error) {
	parts := strings.Split(header, " ")
	if (len(parts) != 2) || parts[0] != "Bearer" || subtle.ConstantTimeCompare([]byte(parts[1]), []byte(authToken)) == 0 {
		level.Debug(logger).Log("msg", errInvalidToken)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusForbidden, Body: errInvalidToken.Error()}, errInvalidToken
	}
	return events.APIGatewayProxyResponse{}, nil
}

func handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var (
		proberp    = request.Path
		authHeader = request.Headers["Authorization"]
		target     = request.QueryStringParameters["target"]
		confp      = request.QueryStringParameters["config"]
	)
	if len(proberp) < 1 {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, errInvalidPath
	}
	proberp = proberp[1:]
	logger := log.With(logger, "target", target, "prober", proberp)
	level.Debug(logger).Log("msg", "Got request")

	res, err := requireAuth(authHeader)
	if err != nil {
		return res, err
	}

	registry := prometheus.NewRegistry()

	module := config.Module{}
	switch proberp {
	case "http":
		if err := yaml.Unmarshal([]byte(confp), &module.HTTP); err != nil {
			return responseConfigInvalid, err
		}
		prober.ProbeHTTP(ctx, target, module, registry, logger)
	case "tcp":
		if err := yaml.Unmarshal([]byte(confp), &module.TCP); err != nil {
			return responseConfigInvalid, err
		}
		prober.ProbeTCP(ctx, target, module, registry, logger)
	case "dns":
		if err := yaml.Unmarshal([]byte(confp), &module.DNS); err != nil {
			return responseConfigInvalid, err
		}
		prober.ProbeDNS(ctx, target, module, registry, logger)
	case "icmp":
		if err := yaml.Unmarshal([]byte(confp), &module.ICMP); err != nil {
			return responseConfigInvalid, err
		}
		prober.ProbeICMP(ctx, target, module, registry, logger)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Prober " + proberp + " does not exist",
		}, errInvalidProber
	}

	mfs, err := registry.Gather()
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	headers := http.Header{}
	for k, v := range request.Headers {
		headers.Add(k, v)
	}
	contentType := expfmt.Negotiate(http.Header(headers))
	buf := &bytes.Buffer{}

	enc := expfmt.NewEncoder(buf, contentType)
	for _, mf := range mfs {
		if err := enc.Encode(mf); err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":   string(contentType),
			"Content-Length": fmt.Sprint(buf.Len()),
		},
		Body: buf.String(),
	}, nil
}

func main() {
	lambda.Start(handle)
}
