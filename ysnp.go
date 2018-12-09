package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	httplogger "github.com/gleicon/go-httplogger"
)

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

	httpAddr := flag.String("listen", ":8080", "TCP host:port to listen on for http requests")
	flag.Parse()

	// dropUri := flag.Bool("dropuri", false, "set true if you want to discard requested uri in redirect")
	log.Printf("ysnp: listening on host %s\n", *httpAddr)
	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/", redirector)

	httpSrv := http.Server{
		Addr:    *httpAddr,
		Handler: httplogger.HTTPLogger(serveMux),
	}
	log.Fatal(httpSrv.ListenAndServe())

}
