package main

import (
  "log"
  "net/http"
  "net/http/httputil"
  "net/url"
  "github.com/joho/godotenv"
  "os"
  "strings"
)

var drupal_latest_url, drupal_legacy_url *url.URL

// These paths are always served from the latest Drupal version (e.g. Drupal 9).
var drupal_latest_paths = []string {
  "/programs/professional-education",
  "/programs/professional-programs",
  "/programs/graduate-degree-programs/blog",
  "/programs/graduate-degree-programs/master-industrial-and-labor-relations-milr",
  "/alumni",
  "/buffalo-co-lab",
  "/cornell-debate",
  "/cjei",
  "/coronavirus",
  "/course",
  "/diversity-equity-and-inclusion",
  "/work-and-coronavirus",
  "/new-york-city",
  "/news",
  "/public-impact",
  "/worker-institute",
  "/scheinman-institute",
  "/scr-summer-school",
  "/current-students",
  "/blog",
  "/ilrie",
  "/ada30",
  "/75",
  "/ithaca-co-lab",
  "/new-conversations-project",
  "/labor-dynamics-institute",
  "/persona",
  "/core",
  "/libraries/union",
  "/themes/custom/union_marketing",
  "/sites/default/files-d8",
  "/system/files/webform",
  "/media/oembed",
  "/modules/contrib",
  "/modules/custom",
}

var shared_paths = []string {
  "/views/ajax",
}

// Return a target URL for a given path. `path` should be the path only, with no
// host or query string.
func getPathTarget(path string, referer string) *url.URL {
  // Always use Drupal-latest for the home page, but not automatically for paths
  // starting with just a slash. Note that even if the user omits the initial
  // `/`, the request will include it, as required by RFC 2616 section 5.1.2.
  if path == "/" || strings.HasPrefix(path, "/?") {
    return drupal_latest_url;
  }

  // Use Drupal-latest for any request path that starts with a value from
  // `drupal_latest_paths`.
  for _, latest_path_prefix := range drupal_latest_paths {
    if strings.HasPrefix(path, latest_path_prefix) {
      return drupal_latest_url
    }
  }

  // Evaluate shared path requests like `/views/ajax`. If the request path
  // starts with one of the `shared_paths` and the `referer` path starts with
  // one of the `drupal_latest_paths`, use Drupal-latest. Note that this test
  // will fail for browsers that disable the `referer` header.
  for _, shared_path_prefix := range shared_paths {
    if strings.HasPrefix(path, shared_path_prefix) {
      for _, latest_path_prefix := range drupal_latest_paths {
        if strings.HasPrefix(referer, latest_path_prefix) {
          return drupal_latest_url
        }
      }
    }
  }

  // All other request paths use Drupal-legacy.
  return drupal_legacy_url
}

func main() {
  if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
    log.Fatalln("Error loading .env")
  }

  // Configure and check for required environment variables.
  env_drupal_latest_url, ok_latest := os.LookupEnv("DRUPAL_LATEST_URL")
  env_drupal_legacy_url, ok_legacy := os.LookupEnv("DRUPAL_LEGACY_URL")
  env_port, ok_port := os.LookupEnv("PORT")

  if !ok_latest {
    log.Fatal("Error loading DRUPAL_LATEST_URL env var.")
  }

  if !ok_legacy {
    log.Fatal("Error loading DRUPAL_LEGACY_URL env var.")
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

    req.Header.Add("X-Forwarded-Host", req.Host)
    req.Header.Add("X-Forwarded-Proto", forwarded_proto)
    req.URL.Scheme = target.Scheme
    req.URL.Host = target.Host
    req.Host = target.Host
  }

  // Modify the response to include source/target hostname.
  modifyResponse := func(response *http.Response) error {
    response.Header.Set("X-ILR-Proxy-Source", response.Request.Host)
    return nil
  }

  proxy := &httputil.ReverseProxy{Director: director, ModifyResponse: modifyResponse}

  http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    proxy.ServeHTTP(w, req)
  })

  log.Printf("Drupal Legacy served from %s.", drupal_legacy_url);
  log.Printf("Drupal Latest served from %s.", drupal_latest_url);
  log.Printf("Listening on port %s", env_port)
  log.Fatal(http.ListenAndServe(":" + env_port, nil))
}
