package server

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
)

func Run(blog_site templ.Component) {
	http.Handle("/", templ.Handler(blog_site))
	slog.Info("Server running", "ip", "0.0.0.0", "port", 3000)
	http.ListenAndServe("0.0.0.0:3000", nil)
}
