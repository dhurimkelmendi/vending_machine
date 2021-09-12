package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/dhurimkelmendi/vending_machine/config"
	"github.com/dhurimkelmendi/vending_machine/internal/trace"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
)

func logRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		st := trace.NewResponseStaller(w)

		h.ServeHTTP(st, r)

		entry := logrus.WithFields(logrus.Fields{
			// When you're operating a webservice that is accessed by clients, it might be difficult to correlate requests
			// (that a client can see) with server logs (that the server can see).
			// The idea of the X-Request-ID is that a client can create some random ID and pass it to the server.
			// The server then include that ID in every log statement that it creates.
			// If a client receives an error it can include the ID in a bug report, allowing the server operator to look up
			// the corresponding log statements (without having to rely on timestamps, IPs, etc).
			// As this ID is generated (randomly) by the client it does not contain any sensitive information,
			// and should thus not violate the user's privacy.
			// As a unique ID is created per request it does also not help with tracking users.
			"X-Request-Id": r.Header.Get("X-Request-Id"),
			"status":       st.Status,
			"method":       r.Method,
			"path":         r.RequestURI,
			"duration":     time.Since(start),
			"RemoteAddr":   r.RemoteAddr,
			// Sometimes the user access the web server via a proxy or load balancer.
			// The above IP address will be the IP address of the proxy or load balancer and not the user's machine.
			// let's get the request HTTP header "X-Forwarded-For (XFF)" if the value returned is not null,
			// then this is the real IP address of the user.
			"X-Forwarded-For": r.Header.Get("X-Forwarded-For"),
		})

		if st.Status >= 400 {
			entry.Warn()
			return
		}
		entry.Info()
	})
}

func requestIDAdder(prefix string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-Id")
			if id == "" {
				id = prefix + "." + ksuid.New().String()
				r.Header.Set("X-Request-Id", id)
			}
			w.Header().Set("X-Request-Id", id)
			h.ServeHTTP(w, r)
		})
	}
}

func getCORSHandler() func(http.Handler) http.Handler {
	cfg := config.GetDefaultInstance()

	allowAllOrigins := cfg.AllowAllCORSOrigins
	allowedOrigins := strings.Split(cfg.CORSOrigins, ",")

	return cors.New(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					return true
				}
			}
			return allowAllOrigins
		},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
	}).Handler
}

// Routes returns the registered HTTP endpoints for the web application.
func Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(getCORSHandler())
	r.Use(logRequest)

	// ctrl := controllers.GetControllersDefaultInstance()
	// Public routes
	r.Route("/api/v1", func(r chi.Router) {

	})
	return r
}
