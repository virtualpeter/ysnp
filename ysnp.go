package main

import (
	"log"
	"net"
	"net/http"

	httplogger "github.com/gleicon/go-httplogger"
)

var httpAddr = ":8080"

func redirector(w http.ResponseWriter, r *http.Request) {
	// log.Printf("req: %+v \n", r)
	u := r.URL
	host, _, err := net.SplitHostPort(r.Host)
	if err == nil {
		u.Host = host
	} else {
		u.Host = r.Host
	}
	u.Scheme = "https"
	// log.Printf("location: %s\n", u.String())
	http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
}

func main() {
	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/", redirector)

	httpSrv := http.Server{
		Addr:    httpAddr,
		Handler: httplogger.HTTPLogger(serveMux),
	}
	log.Fatal(httpSrv.ListenAndServe())

}
