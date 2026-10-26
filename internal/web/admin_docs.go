package web

import "net/http"

func adminDocsCoding(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "docs-coding.html", nil)
}

func adminDocsModules(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "docs-modules.html", nil)
}

func adminDocsAWS(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "docs-aws.html", nil)
}
