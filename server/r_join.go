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
	FillingTeamCode      string
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
	reTeamCode = regexp.MustCompile(`^[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}$`)
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
		TeamCode string `schema:"team_code"`
	}
	if err := unmarshalForm(r, &data); err != nil {
		writeError(err)
		return
	}

	pageData.FillingUsername = data.Username
	pageData.FillingTeamName = data.TeamName
	pageData.FillingTeamCode = data.TeamCode

	if data.TeamName != "" && data.TeamCode != "" {
		writeError(fmt.Errorf("cannot provide both team name and team code"))
		return
	}

	if isAuthenticated {
		// If no username is provided, default to the current username.
		if data.Username == "" {
			data.Username = u.Username
		}

		// If neither team name nor team code is provided, assume the user
		// remains on their current team.
		if data.TeamName == "" && data.TeamCode == "" {
			data.TeamName = u.TeamName
		}
	} else {
		if !reUsername.MatchString(data.Username) {
			writeError(fmt.Errorf("invalid username"))
			return
		}

		// If neither team name nor team code is provided, default to the
		// username.
		if data.TeamName == "" && data.TeamCode == "" {
			data.TeamName = data.Username
		}
	}

	if data.TeamName != "" && !reTeamName.MatchString(data.TeamName) {
		writeError(fmt.Errorf("invalid team name"))
		return
	}

	if data.TeamCode != "" && !reTeamCode.MatchString(data.TeamCode) {
		writeError(fmt.Errorf("invalid team code"))
		return
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

		var isLeader bool
		if data.TeamCode != "" {
			t, err := q.FindTeamWithInviteCode(ctx, data.TeamCode)
			if err != nil {
				return fmt.Errorf("failed to find team: %w", err)
			}
			data.TeamName = t.TeamName
		} else {
			_, err := q.CreateTeam(ctx, db.CreateTeamParams{
				TeamName:   data.TeamName,
				InviteCode: generateInviteCode(),
			})
			if err != nil {
				return fmt.Errorf("failed to create team: %w", err)
			}
			isLeader = true
		}

		_, err := q.JoinTeam(ctx, db.JoinTeamParams{
			TeamName: data.TeamName,
			Username: data.Username,
			IsLeader: isLeader,
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
