package frontend

import (
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/Masterminds/sprig/v3"
	"libdb.so/tmplutil"
)

// Initialize node_modules.
//go:generate npm i --silent --package-lock-only --no-audit --no-fund

// Handle generating CSS from SCSS files.
//go:generate sass -I node_modules/@picocss/pico/scss -s compressed ./styles/styles.scss ./static/styles.css

// ComponentContext is the context passed to all components.
type ComponentContext struct {
	TeamName string
	Username string
}

// NewTemplater returns a new templater with the given filesystem.
func NewTemplater(fs fs.FS) *tmplutil.Templater {
	t := &tmplutil.Templater{
		FileSystem: fs,
		Includes: map[string]string{
			"head":   "components/head.html",
			"header": "components/header.html",
			"footer": "components/footer.html",
		},
		Functions: joinFuncMaps(
			sprig.FuncMap(),
			template.FuncMap{
				"rfc3339": func(t time.Time) string {
					return t.Format(time.RFC3339)
				},
			},
		),
	}
	if err := t.Preregister("pages"); err != nil {
		panic(err)
	}
	return t
}

func joinFuncMaps(maps ...map[string]any) map[string]any {
	out := make(map[string]any)
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

// StaticHandler returns a handler for serving static files.
func StaticHandler(fs_ fs.FS) http.Handler {
	fs_, _ = fs.Sub(fs_, "static")
	if fs_ == nil {
		panic("static files not found")
	}
	return http.StripPrefix("/static", http.FileServer(http.FS(fs_)))
}
