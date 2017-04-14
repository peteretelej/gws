package gws

import (
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// Using a global tmpl variable caches the template (tree) in-memory for fast use
var tmpl = template.Must(template.ParseGlob("tmpl/*.gohtml"))

// Serve launches web server listening on listenAddr
func Serve(listenAddr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleHome)
	mux.Handle("/about", GZIP(http.HandlerFunc(handleAbout))) // example middleware usage

	mux.HandleFunc("/favicon.ico", handleFavicon)

	// serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", GZIP(CACHE(http.StripPrefix("/static", fs)))) // example middleware chaining (GZIP & CACHE)

	svr := &http.Server{
		Addr:           listenAddr,
		Handler:        mux,              // You can wrap the handle in CSRF protection (see github.com/gorilla/csrf)
		ReadTimeout:    15 * time.Second, // ReadTimeout and WriteTimeout are essential to avoiding leaked file descriptors by slow or disappearing clients. You may increase this time in case of long running operations
		WriteTimeout:   20 * time.Second, // Read more here: https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
		MaxHeaderBytes: 1 << 20,
	}
	// Related article on why you should use a custome http.Server: https://blog.cloudflare.com/exposing-go-on-the-internet/

	return svr.ListenAndServe()
	// Since ListenAndServe is being immediately returned here, Serve() is not a long running function,
	// hence do not place go routines here unless properly protected e.g. by sync.WaitGroup
	// however this returns when the server stops and maybe exactly what is desired
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// GZIP is a http gzip compression middleware
func GZIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer func() {
			_ = gz.Close() //ignore this error, in case response status does not allow body (204)
		}()
		gzwr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzwr, r)
	})
}

// CACHE is the caching middleware
func CACHE(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Using recommended long time expires (at least a year). Remember to freshen changed assets (e.g rename)
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", 370*24*3600))

		expires := time.Now().Add(367 * 24 * time.Hour).Format("Mon, 02 Jan 2006 15:04:05 MST")
		// Expires header must be represented in GMT (https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.21)
		// TODO: This silly little conversion below should be replaced with something better and timezone neutral
		expires = strings.Replace(expires, "EAT", "GMT", 1) // fix EAT->GMT ( 3hr diff)
		w.Header().Set("Expires", expires)
		next.ServeHTTP(w, r)
	})
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Page Not Found", http.StatusNotFound)
		return
	}
	data := struct {
		Title string
	}{"GWS: Go Web Server"}
	renderTemplate(w, "home", data)
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
		About string
	}{Title: "About GWS"}
	data.About = "GWS is a Golang web server"
	renderTemplate(w, "about", data)
}

func handleFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/favicon.ico")
}

func renderTemplate(w http.ResponseWriter, page string, data interface{}) {
	secureHeaders(w)
	err := tmpl.ExecuteTemplate(w, page, data)
	if err != nil {
		log.Print(err.Error())
	}
}

func secureHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Add("X-Content-Type-Options", "nosniff")
	w.Header().Add("X-XSS-Protection", "1; mode=block")
	w.Header().Add("X-Frame-Options", "SAMEORIGIN")
	w.Header().Add("X-UA-Compatible", "IE=edge")

	// For HTTPS ONLY domains (recommended), the only reason not to use this is if the server is running locally
	// e.g. 127.0.0.1 and reverse proxied by a webserver that has HTTPS and has this header set
	//w.Header().Add("Strict-Transport-Security", "max-age=16070400; includeSubDomains")
}

// ClientIP returns the client IP address of a request.
// Returns empty string if no IP is found
func ClientIP(r *http.Request) string {
	if val := r.Header.Get("x-forwarded-for"); val != "" {
		return val
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip // empty string if ignored err !=nil
}
