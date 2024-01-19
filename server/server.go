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
	"libdb.so/february-frenzy/server/db"
	"libdb.so/february-frenzy/server/frontend"
	"libdb.so/february-frenzy/server/problem"
	"libdb.so/tmplutil"
)

type Server struct {
	*chi.Mux
	secretKey paseto.V4AsymmetricSecretKey
	problems  ProblemSet
	template  *tmplutil.Templater
	database  *db.Database
	logger    *slog.Logger
}

type ProblemSet struct {
	*problem.ProblemSet
	// ProblemIDs is a list of problem IDs corresponding to the problems in
	// the ProblemSet. This is solely for internal use and should not be
	// exposed to the user.
	ProblemIDs []string
}

// New creates a new server.
func New(
	frontendDir fs.FS,
	secretKey SecretKey,
	problems ProblemSet,
	database *db.Database,
	logger *slog.Logger,
) *Server {
	s := &Server{
		template:  frontend.NewTemplater(frontendDir),
		secretKey: secretKey,
		problems:  problems,
		database:  database,
		logger:    logger,
	}

	s.Mux = chi.NewRouter()
	r := s.Mux

	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Use(s.authMiddleware)
	// r.Use(etagcache.UseAutomatic)

	r.Group(func(r chi.Router) {
		r.Use(middleware.SetHeader("Cache-Control", "private, must-revalidate"))
		r.Get("/", s.index)
		r.Route("/join", s.routeJoin)
		r.Route("/problems", s.routeProblems)
		r.Route("/leaderboard", s.routeLeaderboard)
	})

	r.Route("/static", func(r chi.Router) {
		r.Use(middleware.Compress(5))
		r.Use(middleware.SetHeader("Cache-Control", "public, must-revalidate"))
		r.Mount("/", frontend.StaticHandler(frontendDir))
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

type indexPageData struct {
	frontend.ComponentContext
	InviteCode string
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)

	var inviteCode string
	if u.TeamName != "" {
		team, err := s.database.FindTeam(r.Context(), u.TeamName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, err)
			return
		}
		inviteCode = team.InviteCode
	}

	s.renderTemplate(w, "index", indexPageData{
		ComponentContext: frontend.ComponentContext{
			TeamName: u.TeamName,
			Username: u.Username,
		},
		InviteCode: inviteCode,
	})
}

func parseForm(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}
