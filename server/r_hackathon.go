package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"libdb.so/february-frenzy/server/frontend"
)

func (s *Server) routeHackathon(r chi.Router) {
	r.Get("/", s.hackathonPage)
}

type hackathonPageData struct {
	frontend.ComponentContext
	StartTime      time.Time
	Duration       time.Duration
	SubmissionLink string
}

func (d hackathonPageData) EndTime() time.Time {
	return d.StartTime.Add(d.Duration)
}

func (s *Server) hackathonPage(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	s.renderTemplate(w, "hackathon", hackathonPageData{
		ComponentContext: frontend.ComponentContext{
			Username: u.Username,
			TeamName: u.TeamName,
		},
		StartTime:      s.config.HackathonStart,
		Duration:       s.config.HackathonDuration,
		SubmissionLink: s.config.HackathonSubmissionLink,
	})
}
