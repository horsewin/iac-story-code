package infrastructure

import (
	"net/http"
	"strings"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/uma-arai/iac-story-code/app/handler"
)

// Router ...
func Router() *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Xray middleware
	e.Use(XrWithSkipper())

	// Prometheus middleware
	p := prometheus.NewPrometheus("Echo", nil)
	p.Use(e)

	basePath := "cnis"

	AppHandler := handler.NewAppHandler(NewSQLHandler())
	healthCheckHandler := handler.NewHealthCheckHandler()
	helloWorldHandler := handler.NewHelloWorldHandler()

	e.GET("/healthcheck", healthCheckHandler.HealthCheck())
	e.GET(basePath+"/v1/helloworld", helloWorldHandler.SayHelloWorld())
	e.GET(basePath+"/v1/app", AppHandler.GetAppInfo())
	e.GET(basePath+"/v1/param", AppHandler.GetParamInfo())
	return e
}

// contains... check if a string is present in a slice
func contains(str string, arr []string) bool {
	for _, v := range arr {
		if strings.Contains(str, v) {
			return true
		}
	}

	return false
}

// XrWithSkipper ... Setup AWS X-ray middleware for skipping some URL request
func XrWithSkipper() echo.MiddlewareFunc {
	return echo.WrapMiddleware(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			targets := []string{"/healthcheck"}
			var next http.Handler
			if contains(r.URL.Path, targets) {
				next = h
			} else {
				next = xray.Handler(xray.NewFixedSegmentNamer("inbound-api"), h)
			}
			next.ServeHTTP(w, r)
		})
	})
}
