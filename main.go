package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

var drupal_legacy_url *url.URL

func main() {
	// Configure and check for required environment variables.
	env_drupal_legacy_url, ok_legacy := os.LookupEnv("DRUPAL_LEGACY_URL")
	env_listen, ok_listen := os.LookupEnv("LISTEN")
	env_port, ok_port := os.LookupEnv("PORT")

	if !ok_legacy {
		log.Fatal("Error loading DRUPAL_LEGACY_URL env var.")
	}

	if !ok_listen {
		env_listen = "0.0.0.0"
	}

	if !ok_port {
		log.Fatal("Error loading PORT env var.")
	}

	drupal_legacy_url, _ = url.Parse(env_drupal_legacy_url)

	director := func(req *http.Request) {
		// Check if there is an incoming X-Forwarded-Proto header value. This could
		// happen if this application is running behind reverse proxy. We can use
		// this info to get the protocol (e.g. https) from the original request,
		// since req will not have it.
		forwarded_proto := req.Header.Get("X-Forwarded-Proto")

		if forwarded_proto == "" {
			forwarded_proto = "http"
		}

		// Check if there is an incoming X-Forwarded-Host header value.
		forwarded_host := req.Header.Get("X-Forwarded-Host")

		if forwarded_host == "" {
			req.Header.Add("X-Forwarded-Host", req.Host)
		}

		// req.Header.Add("X-Forwarded-Proto", forwarded_proto)
		req.URL.Scheme = drupal_legacy_url.Scheme
		req.URL.Host = drupal_legacy_url.Host
		req.Host = drupal_legacy_url.Host
	}

	// Modify the response to include source/target hostname.
	modifyResponse := func(response *http.Response) error {
		response.Header.Set("X-ILR-Proxy-Source", response.Request.Host)
		return nil
	}

	errorHandler := func(rw http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error: %+v for path %s.", err, r.URL.Path)
	}

	proxy := &httputil.ReverseProxy{Director: director, ModifyResponse: modifyResponse, ErrorHandler: errorHandler}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		proxy.ServeHTTP(w, req)
	})

	log.Printf("Drupal Legacy served from %s.", drupal_legacy_url)
	log.Printf("Listening on %s:%s", env_listen, env_port)
	log.Fatal(http.ListenAndServe(env_listen+":"+env_port, nil))
}
