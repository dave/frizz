package handler

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"errors"

	"github.com/dave/frizz/config"
	"github.com/dave/jsgo/server/store"
)

func (h *Handler) Page(w http.ResponseWriter, req *http.Request) {
	if err := h.handlePage(w, req); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (h *Handler) handlePage(w http.ResponseWriter, req *http.Request) error {

	ctx, cancel := context.WithTimeout(req.Context(), config.PageTimeout)
	defer cancel()

	var url string
	if config.DEV {
		url = "/_script.js"
	} else {
		found, c, err := store.Package(ctx, h.Database, "github.com/dave/frizz/ed")
		if err != nil {
			return err
		}
		if !found {
			return errors.New("ed package not found")
		}
		url = fmt.Sprintf("https://pkg.jsgo.io/github.com/dave/frizz/ed.%s.js", c.Min.Main)
	}

	v := struct {
		Script string
		Prod   bool
	}{
		Script: url,
		Prod:   !config.DEV,
	}

	if err := pageTemplate.Execute(w, v); err != nil {
		return err
	}
	return nil
}

var pageTemplate = template.Must(template.New("main").Funcs(template.FuncMap{
	"Asset": func(url string) string {
		if config.LOCAL {
			return "/_local" + url[strings.LastIndex(url, "/"):]
		}
		return url
	},
}).Parse(`<html>
	<head>
		<meta charset="utf-8">
		{{ if .Prod -}}
		<script async src="https://www.googletagmanager.com/gtag/js?id=UA-118676357-1"></script>
        <script>
            window.dataLayer = window.dataLayer || [];
            function gtag(){dataLayer.push(arguments);}
            gtag('js', new Date());
            gtag('config', 'UA-118676357-1');
        </script>
		{{- end }}
        <link href="{{ Asset "https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css" }}" rel="stylesheet">
        <script src="{{ Asset "https://code.jquery.com/jquery-3.2.1.slim.min.js" }}"></script>
        <script src="{{ Asset "https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.12.9/umd/popper.min.js" }}"></script>
        <script src="{{ Asset "https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/js/bootstrap.min.js" }}"></script>
		<script>
			window.jsgoProgress = function(count, total) {
				var value = (count * 100.0) / (total * 1.0);
				var bar = document.getElementById("progress-bar");
				bar.style.width = value+"%";
				bar.setAttribute('aria-valuenow', value);
				if (count === total) {
					document.getElementById("progress-holder").style.display = "none";
				}
			}
		</script>
    	<script src="{{ .Script }}"></script>
	</head>
	<body id="wrapper" style="margin: 0;">
		<div id="progress-holder" style="width: 100%; padding: 25%;">
			<div class="progress">
				<div id="progress-bar" class="progress-bar" role="progressbar" style="width: 0%" aria-valuenow="0" aria-valuemin="0" aria-valuemax="100"></div>
			</div>
		</div>
	</body>
</html>`))
