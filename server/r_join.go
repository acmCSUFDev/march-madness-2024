package server

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"libdb.so/february-frenzy/server/db"
	"libdb.so/february-frenzy/server/frontend"
)

func (s *Server) routeJoin(r chi.Router) {
	r.Get("/", s.joinPage)
	r.With(parseForm).Post("/", s.join)
	r.With(parseForm, s.requireAuth).Patch("/", s.join)
}

type joinPageData struct {
	frontend.ComponentContext
	OpenRegistrationTime time.Time
	FillingUsername      string
	FillingTeamName      string
	Error                string
}

func (s *Server) joinPage(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	s.renderTemplate(w, "join", joinPageData{
		ComponentContext: frontend.ComponentContext{
			TeamName: u.TeamName,
			Username: u.Username,
		},
		OpenRegistrationTime: s.config.OpenRegistrationTime,
	})
}

var (
	reUsername = regexp.MustCompile(`^[a-zA-Z0-9-_ ]{2,32}$`)
	reTeamName = regexp.MustCompile(`^[a-zA-Z0-9-_ ]{2,32}$`)
)

func (s *Server) join(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	isAuthenticated := isAuthenticated(r)

	ctx := r.Context()
	pageData := joinPageData{
		ComponentContext: frontend.ComponentContext{
			TeamName: u.TeamName,
			Username: u.Username,
		},
		OpenRegistrationTime: s.config.OpenRegistrationTime,
	}

	writeError := func(err error) {
		pageData.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		s.renderTemplate(w, "join", pageData)
	}

	var data struct {
		Username string `schema:"username"`
		TeamName string `schema:"team_name"`
	}
	if err := unmarshalForm(r, &data); err != nil {
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

	if isAuthenticated {
		data.Username = u.Username
	}

	if data.TeamName == "" && isAuthenticated {
		data.TeamName = u.TeamName
	}

	err := s.database.Tx(func(q *db.Queries) error {
		if isAuthenticated {
			isLeader, err := q.IsLeader(ctx, db.IsLeaderParams{
				TeamName: u.TeamName,
				Username: u.Username,
			})
			if err != nil {
				return fmt.Errorf("failed to check if leader: %w", err)
			}
			if isLeader {
				return fmt.Errorf("you cannot leave your own team")
			}

			_, err = q.LeaveTeam(ctx, db.LeaveTeamParams{
				TeamName: u.TeamName,
				Username: u.Username,
			})
			if err != nil {
				return fmt.Errorf("failed to leave team: %w", err)
			}
		}

		team, err := q.FindTeamWithInviteCode(ctx, data.TeamName)
		if err == nil {
			_, err := q.JoinTeam(ctx, db.JoinTeamParams{
				TeamName: team.TeamName,
				Username: data.Username,
			})
			if err != nil {
				return fmt.Errorf("failed to join team: %w", err)
			}
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
			Username: data.Username,
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