package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"strings"

	httplogger "github.com/gleicon/go-httplogger"
)

func redirector(w http.ResponseWriter, r *http.Request) {
	log.Printf("req: %+v \n", r)
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

	// log.Printf("location: %s\n", u.String())

	log.Printf("redirect: %+v \n", u)

	http.Redirect(w, r, u.String(), *redirectStatus)
}

var httpAddr = flag.String("listen", ":8080", "TCP host:port to listen on for http requests")
var targetProto = flag.String("target_proto", "https", "protocol to redirect to, so far the only other supported option is http")
var targetHost = flag.String("target_host", "", "hardcode this domainname in redirect instead of passing on request")
var targetPort = flag.String("target_port", "", "port to use in redirect, default is to not have an explicit port")
var targetPath = flag.String("target_path", "", "hardcode this path in redirect, default means use request path")
var passQuery = flag.Bool("passquery", false, "set true if you want to pass request query parameters in redirect")
var redirectStatus = flag.Int("status", http.StatusMovedPermanently, "http status 3xx code to return")

func main() {

	flag.Parse()

	if !isValidRedirectStatus(*redirectStatus) {
		log.Fatal("redirect status must be one of 301,302,307,308")
	}

	log.Printf("ysnp: listening on host %s\n", *httpAddr)
	log.Printf("ysnp: target pattern %s://%s:%s/%s\n", *targetProto, *targetHost, *targetPort, *targetPath)

	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/", redirector)

	httpSrv := http.Server{
		Addr:    *httpAddr,
		Handler: httplogger.HTTPLogger(serveMux),
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
