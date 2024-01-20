package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"libdb.so/february-frenzy/server/frontend"
)

func (s *Server) routeHackathon(r chi.Router) {
	r.Get("/", s.hackathonPage)
}

type hackathonPageData struct {
	frontend.ComponentContext
}

func (s *Server) hackathonPage(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	s.renderTemplate(w, "hackathon", hackathonPageData{
		ComponentContext: frontend.ComponentContext{
			Username: u.Username,
			TeamName: u.TeamName,
		},
	})
}
