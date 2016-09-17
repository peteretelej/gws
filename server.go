package gws

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

var tmpl = template.Must(template.ParseGlob("tmpl/*.gohtml"))

// Serve launches web server listening on listenAddr
func Serve(listenAddr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/about", handleAbout)
	mux.HandleFunc("/favicon.ico", handleFavicon)

	// serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static", fs))

	svr := &http.Server{
		Addr:           listenAddr,
		Handler:        mux,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return svr.ListenAndServe()
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

	// For HTTPS ONLY domains (recommended)
	//w.Header().Add("Strict-Transport-Security", "max-age=16070400; includeSubDomains")
}
