package main

import (
	"flag"
	"net"
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
var passQuery = flag.Bool("passquery", false, "set true if you want to pass request query parameters in redirect")
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
	u := r.URL
	if *targetHost == "" {
		host, _, err := net.SplitHostPort(r.Host)
		if err == nil {
			u.Host = host
		} else {
			u.Host = r.Host
		}
	} else {
		u.Host = *targetHost
	}
	if *targetPort != "" {
		host := strings.Split(u.Host, ":")[0]
		u.Host = host + ":" + *targetPort
	}
	if *targetProto == "http" {
		u.Scheme = "http"
	} else {
		u.Scheme = "https"
	}
	if *targetPath != "" {
		u.Path = *targetPath
	}
	if !*passQuery {
		u.RawQuery = ""
	}

	log.Debugf("redirect: %+v \n", u)

	http.Redirect(w, r, u.String(), *redirectStatus)
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
			"remoteAddr": r.RemoteAddr,
			// "requestTime": t.Format("02/Jan/2006:15:04:05 -0000"),
			"requestTime": t.Format(time.RFC3339),
			"method":      r.Method,
			"requestURL":  r.URL.Path,
			"proto":       r.Proto,
			"userAgent":   r.UserAgent(),
		}).Info()
	})
}
