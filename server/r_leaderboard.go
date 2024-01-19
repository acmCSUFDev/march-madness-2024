package server

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"
	"libdb.so/february-frenzy/server/db"
	"libdb.so/february-frenzy/server/frontend"
)

func (s *Server) routeLeaderboard(r chi.Router) {
	r.Get("/", s.leaderboard)
}

type leaderboardPageData struct {
	frontend.ComponentContext
	StartedAt time.Time

	database *db.Database
	logger   *slog.Logger
	ctx      context.Context
}

type leaderboardTeamPointsTable struct {
	Reasons []string
	Teams   []string
	Points  [][]float64
}

func (d leaderboardPageData) TeamPointsTable() (leaderboardTeamPointsTable, error) {
	var table leaderboardTeamPointsTable

	pointsRows, err := d.database.TeamPoints(d.ctx)
	if err != nil {
		return table, err
	}

	for _, row := range pointsRows {
		// table.Reasons = append(table.Reasons, row.Reason)
		table.Teams = append(table.Teams, row.TeamName)
	}

	// table.Reasons = slices.Compact(table.Reasons)
	table.Reasons = []string{"week of code", "hackathon"}
	table.Teams = slices.Compact(table.Teams)

	table.Points = make([][]float64, len(table.Teams))
	for i := range table.Points {
		table.Points[i] = make([]float64, len(table.Reasons))
	}

	for _, row := range pointsRows {
		i := slices.Index(table.Teams, row.TeamName)
		j := slices.Index(table.Reasons, row.Reason.String)
		if i != -1 && j != -1 {
			table.Points[i][j] = row.Points.Float64
		} else {
			d.logger.WarnContext(d.ctx,
				"leaderboard: unexpected team or reason",
				"team", row.TeamName,
				"reason", row.Reason)
		}
	}

	return table, nil
}

type leaderboardTeamPointsEvent struct {
	TeamName string    `json:"team_name"`
	AddedAt  time.Time `json:"added_at"`
	Reason   string    `json:"reason"`
	Points   float64   `json:"points"`
}

func (d leaderboardPageData) TeamPointsEvents() ([]leaderboardTeamPointsEvent, error) {
	var events []leaderboardTeamPointsEvent

	rows, err := d.database.TeamPointsHistory(d.ctx)
	if err != nil {
		return events, err
	}

	for _, row := range rows {
		events = append(events, leaderboardTeamPointsEvent{
			TeamName: row.TeamName,
			AddedAt:  row.AddedAt,
			Reason:   row.Reason,
			Points:   row.Points,
		})
	}

	return events, nil
}

func (s *Server) leaderboard(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	s.renderTemplate(w, "leaderboard", leaderboardPageData{
		ComponentContext: frontend.ComponentContext{
			TeamName: u.TeamName,
			Username: u.Username,
		},
		StartedAt: s.problems.StartedAt(),
		database:  s.database,
		logger:    s.logger,
		ctx:       r.Context(),
	})
}
