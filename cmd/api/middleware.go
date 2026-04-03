package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"expvar"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ReneeB2022/test1/internal/data"
	"github.com/ReneeB2022/test1/internal/validator"
	"golang.org/x/time/rate"
)

type metricsResponseWriter struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		wrapped:    w,
		statusCode: http.StatusOK,
	}
}

func (mw *metricsResponseWriter) Header() http.Header {
	return mw.wrapped.Header()
}
func (mw *metricsResponseWriter) WriteHeader(statusCode int) {
	mw.wrapped.WriteHeader(statusCode)
	if !mw.headerWritten {
		mw.statusCode = statusCode
		mw.headerWritten = true
	}
}
func (mw *metricsResponseWriter) Write(b []byte) (int, error) {
	mw.headerWritten = true
	return mw.wrapped.Write(b)
}

// We need a function to get the original http.ResponseWriter
func (mw *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return mw.wrapped
}

type gzipcompress struct {
	http.ResponseWriter
	size          int
	buffer        *bytes.Buffer
	gz            *gzip.Writer
	headerWritten bool
	code          int
}

func (a *applicationDependencies) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := a.contextGetUser(r)

		if user.IsAnonymous() {
			a.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *applicationDependencies) requirePermission(permissionCode string, next http.HandlerFunc) http.HandlerFunc {

	fn := func(w http.ResponseWriter, r *http.Request) {
		user := a.contextGetUser(r)
		// get all the permissions associated with the user
		permissions, err := a.permissionModel.GetAllForUser(user.ID)
		if err != nil {
			a.serverErrorResponse(w, r, err)
			return
		}
		if !permissions.Include(permissionCode) {
			a.notPermittedResponse(w, r)
			return
		}
		// they are good. Let's keep going
		next.ServeHTTP(w, r)
	}

	return a.requireActivatedUser(fn)

}

// This middleware checks if the user is activated
// It call the authentication middleware to help it do its job
func (a *applicationDependencies) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := a.contextGetUser(r)

		if !user.Activated {
			a.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	//We pass the activation check middleware to the authentication
	// middleware to call (next) if the authentication check succeeds
	// In other words, only check if the user is activated if they are
	// actually authenticated.
	return a.requireAuthenticatedUser(fn)
}

func (a *applicationDependencies) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Vary", "Authorization")

		// Get the Authorization header from the request. It should have the
		// Bearer token
		authorizationHeader := r.Header.Get("Authorization")

		// If there is no Authorization header then we have an Anonymous user
		if authorizationHeader == "" {
			r = a.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}
		// Bearer token present so parse it. The Bearer token is in the form
		// Authorization: Bearer IEYZQUBEMPPAKPOAWTPV6YJ6RM
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			a.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Get the actual token
		token := headerParts[1]
		// Validate
		v := validator.New()

		data.ValidateTokenPlaintext(v, token)
		if !v.IsEmpty() {
			a.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Get the user info associated with this authentication token
		user, err := a.userModel.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				a.invalidAuthenticationTokenResponse(w, r)
			default:
				a.serverErrorResponse(w, r, err)
			}
			return
		}
		// Add the retrieved user info to the context
		r = a.contextSetUser(r, user)

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

func (w *gzipcompress) Write(b []byte) (int, error) {
	// If already compressing, write straight to gzip
	if w.gz != nil {
		return w.gz.Write(b)
	}

	// Still buffering: check if this write puts us over the limit
	if w.buffer.Len()+len(b) > w.size {
		// size exceeded: Initialize gzip
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length") // Length will change due to compression

		if w.code == 0 {
			w.code = http.StatusOK
		}
		w.ResponseWriter.WriteHeader(w.code)

		w.gz = gzip.NewWriter(w.ResponseWriter)

		// Flush the existing buffer to gzip first
		if _, err := w.gz.Write(w.buffer.Bytes()); err != nil {
			return 0, err
		}
		w.buffer.Reset()

		return w.gz.Write(b)
	}

	// Under size: just buffer for now
	return w.buffer.Write(b)
}

func (w *gzipcompress) WriteHeader(code int) {
	w.code = code
}

func GzipMiddleware(size int, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Vary", "Accept-Encoding")

		gzipw := &gzipcompress{
			ResponseWriter: w,
			size:           size,
			buffer:         &bytes.Buffer{},
		}

		next.ServeHTTP(gzipw, r)

		// Finalize: if we never started gzipping, write the buffered content normally
		if gzipw.gz == nil {
			if gzipw.code == 0 {
				gzipw.code = http.StatusOK
			}
			gzipw.ResponseWriter.WriteHeader(gzipw.code)
			gzipw.ResponseWriter.Write(gzipw.buffer.Bytes())
		} else {
			gzipw.gz.Close() // Flush and close gzip stream
		}
	})
}

func (a *applicationDependencies) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Vary", "Origin")
		// Let's check the request origin to see if it's in the trusted list
		w.Header().Add("Vary", "Access-Control-Request-Method")
		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range a.config.cors.trustedOrigins {
				if origin == a.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					if r.Method == http.MethodOptions &&
						r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods",
							"OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers",
							"Authorization, Content-Type")

						w.WriteHeader(http.StatusOK)
						return
					}

					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// this middleware will run for every request received
func (a *applicationDependencies) metrics(next http.Handler) http.Handler {
	// Setup our variable to track the metrics
	var (
		totalRequestsReceived           = expvar.NewInt("total_requests_received")
		totalResponsesSent              = expvar.NewInt("total_responses_sent")
		totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_μs")
		totalResponsesSentByStatus      = expvar.NewMap("total_responses_sent_by_status")
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// start is when we receive the request and start processing it
		start := time.Now()
		// update our request received counter
		totalRequestsReceived.Add(1)
		// call the next handler in our chain
		mw := newMetricsResponseWriter(w)
		// we send our custom responseWriter down the middleware chain
		next.ServeHTTP(mw, r)

		// remember the middleware chain goes in both directions, so we can
		// do things when we return back to our middleware.We will increment
		// the responses sent counter
		totalResponsesSent.Add(1)

		// calculate the processing time for this request. Remember we set start
		// at the beginning, so now since we are back in the middleware we can
		// compute the time taken
		totalResponsesSentByStatus.Add(strconv.Itoa(mw.statusCode), 1)
		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(duration)
	})
}

func LoggingMiddleware(logger http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Method: %v, Pathing: %v,\nTimestamp: %v\n", r.Method, r.URL.Path, time.Now().Format(time.RFC1123))
		logger.ServeHTTP(w, r)
	})
}
func (a *applicationDependencies) rateLimit(next http.Handler) http.Handler {
	// Define a rate limiter struct
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time // remove map entries that are stale
	}
	var mu sync.Mutex                      // use to synchronize the map
	var clients = make(map[string]*client) // the actual map
	// A goroutine to remove stale entries from the map
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock() // begin cleanup
			// delete any entry not seen in three minutes
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock() // finish clean up
		}
	}()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the IP address
		if a.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				a.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock() // exclusive access to the map
			// check if ip address already in map, if not add it
			_, found := clients[ip]
			if !found {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(a.config.limiter.rps), a.config.limiter.burst)}
			}
			// Update the last seem for the client
			clients[ip].lastSeen = time.Now()

			// Check the rate limit status
			if !clients[ip].limiter.Allow() {
				mu.Unlock() // no longer need exclusive access to the map
				a.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock() // others are free to get exclusive access to the map
		}
		next.ServeHTTP(w, r)
	})
}

func (a *applicationDependencies) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer will be called when the stack unwinds
		defer func() {
			// recover() checks for panics
			err := recover()
			if err != nil {
				w.Header().Set("Connection", "close")
				a.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
