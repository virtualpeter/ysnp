package main

import (
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	logflag "github.com/reenjii/logflag"
)

var httpAddr = flag.String("listen", ":8080", "TCP host:port to listen on for http requests")
var targetProto = flag.String("target_proto", "https", "protocol to redirect to, so far the only other supported option is http")
var targetHost = flag.String("target_host", "", "hardcode this domainname in redirect instead of passing on request")
var targetPort = flag.String("target_port", "", "port to use in redirect, default is to not have an explicit port")
var targetPath = flag.String("target_path", "", "hardcode this path in redirect, default means use request path")
var blockQuery = flag.Bool("blockquery", false, "set if you want to block passing of request query parameters in redirect")
var redirectStatus = flag.Int("status", http.StatusMovedPermanently, "http status 3xx code to return")

func init() {
	log.SetOutput(os.Stdout)
}

func main() {

	flag.Parse()
	logflag.Parse()

	if !isValidRedirectStatus(*redirectStatus) {
		log.Fatal("redirect status must be one of 301,302,307,308")
	}

	log.WithFields(log.Fields{
		"listenAddr":  *httpAddr,
		"targetProto": *targetProto,
		"targetHost":  *targetHost,
		"targetPort":  *targetPort,
		"targetPath":  *targetPath,
	}).Info("Configuration")

	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/", redirector)

	httpSrv := http.Server{
		Addr:    *httpAddr,
		Handler: httpLogger(serveMux),
	}
	log.Fatal(httpSrv.ListenAndServe())

}

func isValidRedirectStatus(f int) bool {
	answer := false
	for _, x := range []int{301, 302, 307, 308} {
		if x == f {
			answer = true
		}
	}
	return answer
}

func redirector(w http.ResponseWriter, r *http.Request) {
	log.Debugf("req: %+v \n", r)
	u := r
	if *targetHost == "" {
		u.URL.Host = r.URL.Hostname()

	} else {
		u.URL.Host = *targetHost
	}
	if *targetPort != "" {
		host := strings.Split(u.URL.Host, ":")[0]
		u.Host = host + ":" + *targetPort
	}
	if *targetProto == "http" {
		u.URL.Scheme = "http"
	} else {
		u.URL.Scheme = "https"
	}
	if *targetPath != "" {
		u.URL.Path = *targetPath
	}
	if *blockQuery {
		path := strings.Split(u.URL.Path, "&")[0]
		u.URL.Path = path
	}

	log.Debugf("Location: %s\n", u.URL.String())

	http.Redirect(w, r, u.URL.String(), *redirectStatus)
}

type stResponseWriter struct {
	http.ResponseWriter
	HTTPStatus   int
	ResponseSize int
}

func httpLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		interceptWriter := stResponseWriter{w, 0, 0}

		handler.ServeHTTP(&interceptWriter, r)
		log.WithFields(log.Fields{
			"remoteAddr":  r.RemoteAddr,
			"requestTime": t.Format(time.RFC3339),
			"method":      r.Method,
			"requestURL":  r.URL.Path,
			"proto":       r.Proto,
			"userAgent":   r.UserAgent(),
		}).Info()
	})
}
