package server

import (
	"database/sql"
	"net/http"
	"slices"
	"time"

	"dev.acmcsuf.com/march-madness-2024/internal/config"
	"dev.acmcsuf.com/march-madness-2024/server/db"
	"dev.acmcsuf.com/march-madness-2024/server/frontend"
	"github.com/go-chi/chi/v5"
)

func (s *Server) routeHackathon(r chi.Router) {
	r.Get("/", s.hackathonPage)
	r.With(s.requireAuth).Post("/submit", s.submitHackathon)
}

type hackathonForm struct {
	ProjectURL         string `schema:"project_url"`
	ProjectDescription string `schema:"project_description"`
	Category           string `schema:"category"`
}

type hackathonPageData struct {
	frontend.ComponentContext
	config.HackathonConfig
	Submission hackathonForm
}

func (s *Server) hackathonPage(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	ctx := r.Context()

	var form hackathonForm
	if u.TeamName != "" {
		submission, err := s.database.HackathonSubmission(ctx, u.TeamName)
		if err == nil {
			form.ProjectURL = submission.ProjectUrl
			form.ProjectDescription = submission.ProjectDescription.String
			form.Category = submission.Category
		}
	}

	s.renderTemplate(w, "hackathon", hackathonPageData{
		ComponentContext: frontend.ComponentContext{
			Username: u.Username,
			TeamName: u.TeamName,
		},
		HackathonConfig: s.config.HackathonConfig,
		Submission:      form,
	})
}

var categories = []string{
	"interactive",
	"lazy",
	"otherworldly",
	"non-ai",
}

func (s *Server) submitHackathon(w http.ResponseWriter, r *http.Request) {
	if !s.config.HackathonConfig.IsOpen(time.Now()) {
		http.Error(w, "hackathon is not open", http.StatusBadRequest)
		return
	}

	u := getAuthentication(r)
	ctx := r.Context()

	var submission hackathonForm
	if err := unmarshalForm(r, &submission); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !slices.Contains(categories, submission.Category) {
		http.Error(w, "invalid category", http.StatusBadRequest)
		return
	}

	if submission.ProjectURL == "" {
		http.Error(w, "project url is required", http.StatusBadRequest)
		return
	}

	if err := s.database.SetHackathonSubmission(ctx, db.SetHackathonSubmissionParams{
		TeamName:   u.TeamName,
		ProjectUrl: submission.ProjectURL,
		ProjectDescription: sql.NullString{
			String: submission.ProjectDescription,
			Valid:  submission.ProjectDescription != "",
		},
		Category: submission.Category,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.renderTemplate(w, "hackathon_submitted", frontend.ComponentContext{
		Username: u.Username,
		TeamName: u.TeamName,
	})
}
