package server

import (
	"bytes"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"libdb.so/february-frenzy/server/problem"
	"libdb.so/tmplutil"
)

type Server struct {
	*chi.Mux
	secretKey paseto.V4AsymmetricSecretKey
	template  *tmplutil.Templater
	problems  *problem.ProblemSet
	logger    *slog.Logger
}

// New creates a new server.
func New(
	frontendDir fs.FS,
	secretKey SecretKey,
	problems *problem.ProblemSet,
	logger *slog.Logger,
) *Server {
	s := &Server{
		template: tmplutil.Preregister(&tmplutil.Templater{
			FileSystem: frontendDir,
			Includes: map[string]string{
				"head": "components/head.html",
			},
		}),
		secretKey: secretKey,
		problems:  problems,
		logger:    logger,
	}

	s.Mux = chi.NewRouter()

	s.Use(middleware.RealIP)
	s.Use(middleware.Recoverer)
	s.Use(middleware.Timeout(30 * time.Second))

	s.Use(s.authMiddleware)

	s.Get("/", s.index)

	s.Route("/join", func(r chi.Router) {
		r.Get("/", s.joinPage)
		r.Post("/", s.join)
	})

	s.Route("/problems", func(r chi.Router) {
		r.Get("/", s.listProblems)
		r.Get("/{problemID}", s.viewProblem)
		r.Get("/{problemID}/input", s.viewProblemInput)
		r.With(s.requireAuth).Post("/{problemID}/submit", s.submitProblem)
	})

	return s
}

func (s *Server) renderTemplate(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	start := time.Now()

	var out bytes.Buffer
	out.Grow(512)
	if err := s.template.Execute(&out, name, data); err != nil {
		s.logger.Error(
			"failed to render template",
			"name", name,
			"err", err)

		writeError(w, http.StatusInternalServerError, err)
		return
	}

	taken := time.Since(start)
	s.logger.Debug(
		"rendered template",
		"took", taken,
		"size", out.Len())

	w.Write(out.Bytes())
}

func writeError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Error(w, err.Error(), code)
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "index", nil)
}

func (s *Server) joinPage(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) join(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) listProblems(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) viewProblem(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) viewProblemInput(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) submitProblem(w http.ResponseWriter, r *http.Request) {
}
