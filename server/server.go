package server

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/schema"
	"libdb.so/etagcache"
	"libdb.so/february-frenzy/server/db"
	"libdb.so/february-frenzy/server/frontend"
	"libdb.so/february-frenzy/server/problem"
	"libdb.so/tmplutil"
)

type Server struct {
	*chi.Mux
	secretKey paseto.V4AsymmetricSecretKey
	template  *tmplutil.Templater
	problems  *problem.ProblemSet
	database  *db.Database
	logger    *slog.Logger
}

// New creates a new server.
func New(
	frontendDir fs.FS,
	secretKey SecretKey,
	problems *problem.ProblemSet,
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
	r.Use(middleware.Timeout(30 * time.Second))

	r.Use(s.authMiddleware)
	r.Use(etagcache.UseAutomatic)

	r.Group(func(r chi.Router) {
		r.Use(middleware.SetHeader("Cache-Control", "private, must-revalidate"))
		r.Get("/", s.index)

		r.Route("/join", func(r chi.Router) {
			r.Get("/", s.joinPage)
			r.With(parseForm).Post("/", s.join)
			r.With(parseForm, s.requireAuth).Patch("/", s.join)
		})

		r.Route("/problems", func(r chi.Router) {
			r.Get("/", s.listProblems)
			r.Get("/{problemID}", s.viewProblem)
			r.Get("/{problemID}/input", s.viewProblemInput)
			r.With(s.requireAuth).Post("/{problemID}/submit", s.submitProblem)
		})
	})

	r.Route("/static", func(r chi.Router) {
		r.Use(middleware.Compress(5))
		r.Use(middleware.SetHeader("Cache-Control", "public, max-age=3600, must-revalidate"))
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

type joinPageData struct {
	frontend.ComponentContext
	FillingUsername string
	FillingTeamName string
	Error           string
}

func (s *Server) joinPage(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	s.renderTemplate(w, "join", joinPageData{
		ComponentContext: frontend.ComponentContext{
			TeamName: u.TeamName,
			Username: u.Username,
		},
	})
}

var decoder = schema.NewDecoder()

var (
	reUsername = regexp.MustCompile(`^[a-zA-Z0-9-_ ]{2,32}$`)
	reTeamName = regexp.MustCompile(`^[a-zA-Z0-9-_ ]{2,32}$`)
)

func (s *Server) join(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	isAuthenticated := isAuthenticated(r)

	ctx := r.Context()
	pageData := joinPageData{}

	writeError := func(err error) {
		pageData.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		s.renderTemplate(w, "join", pageData)
	}

	var data struct {
		Username string `schema:"username"`
		TeamName string `schema:"team_name"`
	}
	if err := decoder.Decode(&data, r.PostForm); err != nil {
		writeError(err)
		return
	}

	pageData.FillingUsername = data.Username
	pageData.FillingTeamName = data.TeamName

	if data.Username != "" || !isAuthenticated {
		if !reUsername.MatchString(data.Username) {
			writeError(fmt.Errorf("invalid username"))
			return
		}
	}

	if data.TeamName != "" || !isAuthenticated {
		if !reTeamName.MatchString(data.TeamName) {
			writeError(fmt.Errorf("invalid team name"))
			return
		}
	}

	if data.Username == "" && isAuthenticated {
		data.Username = u.Username
	}

	if data.TeamName == "" && isAuthenticated {
		data.TeamName = u.TeamName
	}

	err := s.database.Tx(func(q *db.Queries) error {
		if isAuthenticated {
			isLeader, err := q.IsLeader(ctx, db.IsLeaderParams{
				TeamName: u.TeamName,
				UserName: u.Username,
			})
			if err != nil {
				return fmt.Errorf("failed to check if leader: %w", err)
			}
			if isLeader {
				return fmt.Errorf("cannot leave team as leader")
			}

			_, err = q.LeaveTeam(ctx, db.LeaveTeamParams{
				TeamName: u.TeamName,
				UserName: u.Username,
			})
			if err != nil {
				return fmt.Errorf("failed to leave team: %w", err)
			}
		}

		team, err := q.FindTeamWithInviteCode(ctx, data.TeamName)
		if err == nil {
			_, err := q.JoinTeam(ctx, db.JoinTeamParams{
				TeamName: team.TeamName,
				UserName: data.Username,
			})
			if err != nil {
				return fmt.Errorf("failed to join team: %w", err)
			}
			data.TeamName = team.TeamName
			return nil
		}

		_, err = q.CreateTeam(ctx, db.CreateTeamParams{
			TeamName:   data.TeamName,
			InviteCode: generateInviteCode(),
		})
		if err != nil {
			return fmt.Errorf("failed to create team: %w", err)
		}

		_, err = q.JoinTeam(ctx, db.JoinTeamParams{
			TeamName: data.TeamName,
			UserName: data.Username,
			IsLeader: true,
		})
		if err != nil {
			return fmt.Errorf("failed to join team: %w", err)
		}

		return nil
	})
	if err != nil {
		writeError(err)
		return
	}

	s.setTokenCookie(w, authenticatedUser{
		Username: data.Username,
		TeamName: data.TeamName,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type problemsPageData struct {
	frontend.ComponentContext
	Problems *problem.ProblemSet
}

func (s *Server) listProblems(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	s.renderTemplate(w, "problems", problemsPageData{
		ComponentContext: frontend.ComponentContext{
			TeamName: u.TeamName,
			Username: u.Username,
		},
		Problems: s.problems,
	})
}

func (s *Server) viewProblem(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) viewProblemInput(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) submitProblem(w http.ResponseWriter, r *http.Request) {
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
