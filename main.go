package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

var drupal_latest_url, drupal_legacy_url *url.URL

// These paths are always served from the legacy Drupal version (e.g. Drupal 7).
var drupal_legacy_paths = []string{
	"/buffalo/about",
	"/conference-center",
	"/eform",
	"/faculty-and-staff-resources",
	// This one needs authentication.
	"/faculty-reporting",
	"/ilr-in-buffalo",
	"/ilr-press",
	"/misc",
	"/modules/node",
	"/modules/system",
	"/modules/user",
	"/mobilizing-against-inequality",
	"/nyc-conference-center",
	"/privacy-policy",
	"/sitemap.xml",
	"/sites/all/libraries",
	"/sites/all/modules",
	"/sites/all/themes",
	// Note the trailing slash. Without it, Drupal-latest /sites/default/files-d8 will be included, too.
	"/sites/default/files/",
	"/student-forms",
	"/web-accessibility",
}

var shared_paths = []string{
	"/views/ajax",
}

// Return a target URL for a given path. `path` should be the path only, with no
// host or query string.
func getPathTarget(path string, referer string) *url.URL {
	// Use Drupal-legacy for any request path that starts with a value from
	// `drupal_legacy_paths`.
	for _, legacy_path_prefix := range drupal_legacy_paths {
		if strings.HasPrefix(path, legacy_path_prefix) {
			return drupal_legacy_url
		}
	}

	// Evaluate shared path requests like `/views/ajax`. If the request path
	// starts with one of the `shared_paths` and the `referer` path starts with
	// one of the `drupal_legacy_paths`, use Drupal-legacy. Note that this test
	// will fail for browsers that disable the `referer` header.
	for _, shared_path_prefix := range shared_paths {
		if strings.HasPrefix(path, shared_path_prefix) {
			for _, legacy_path_prefix := range drupal_legacy_paths {
				if strings.HasPrefix(referer, legacy_path_prefix) {
					return drupal_legacy_url
				}
			}
		}
	}

	// All other request paths use Drupal-latest.
	return drupal_latest_url
}

func main() {
	// Configure and check for required environment variables.
	env_drupal_latest_url, ok_latest := os.LookupEnv("DRUPAL_LATEST_URL")
	env_drupal_legacy_url, ok_legacy := os.LookupEnv("DRUPAL_LEGACY_URL")
	env_listen, ok_listen := os.LookupEnv("LISTEN")
	env_port, ok_port := os.LookupEnv("PORT")

	if !ok_latest {
		log.Fatal("Error loading DRUPAL_LATEST_URL env var.")
	}

	if !ok_legacy {
		log.Fatal("Error loading DRUPAL_LEGACY_URL env var.")
	}

	if !ok_listen {
		env_listen = "0.0.0.0"
	}

	if !ok_port {
		log.Fatal("Error loading PORT env var.")
	}

	drupal_latest_url, _ = url.Parse(env_drupal_latest_url)
	drupal_legacy_url, _ = url.Parse(env_drupal_legacy_url)

	director := func(req *http.Request) {
		referer, _ := url.Parse(req.Header.Get("Referer"))

		// Choose the target host based on the incoming request path.
		target := getPathTarget(req.URL.Path, referer.Path)

		// Check if there is an incoming X-Forwarded-Proto header value. This could
		// happen if this application is running behind reverse proxy. We can use
		// this info to get the protocol (e.g. https) from the original request,
		// since req will not have it.
		forwarded_proto := req.Header.Get("X-Forwarded-Proto")

		if forwarded_proto == "" {
			forwarded_proto = "http"
		}

		// req.Header.Add("X-Forwarded-Host", req.Host)
		// req.Header.Add("X-Forwarded-Proto", forwarded_proto)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
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
	log.Printf("Drupal Latest served from %s.", drupal_latest_url)
	log.Printf("Listening on %s:%s", env_listen, env_port)
	log.Fatal(http.ListenAndServe(env_listen+":"+env_port, nil))
}
