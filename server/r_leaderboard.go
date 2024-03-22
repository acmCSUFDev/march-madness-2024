package server

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"dev.acmcsuf.com/march-madness-2024/server/frontend"
	"github.com/go-chi/chi/v5"
)

func (s *Server) routeLeaderboard(r chi.Router) {
	r.Get("/", s.leaderboard)
}

type leaderboardPageData struct {
	frontend.ComponentContext
	StartedAt time.Time
	Table     leaderboardTeamPointsTable
	Events    []leaderboardTeamPointsEvent
}

// TODO: this is awful, refactor it maybe
type leaderboardTeamPointsTable struct {
	Teams            []string
	TeamTotals       []float64
	TeamMembers      [][]string
	TeamPoints       [][]teamPoints
	WeekOfCodeSolves [][]int8 // list of teams, each containing N days
}

type teamPoints struct {
	Reason string
	Points float64
}

func (t leaderboardTeamPointsTable) TeamPointsTooltip(teamIx int) string {
	vals := make([]string, len(t.TeamPoints[teamIx]))
	for i, p := range t.TeamPoints[teamIx] {
		vals[i] = fmt.Sprintf("%s: %.0f", p.Reason, math.Floor(p.Points))
	}
	return strings.Join(vals, ", ")
}

func (t leaderboardTeamPointsTable) TeamMembersTooltip(teamIx int) string {
	return strings.Join(t.TeamMembers[teamIx], ", ")
}

type leaderboardTeamPointsEvent struct {
	TeamName string    `json:"team_name"`
	AddedAt  time.Time `json:"added_at"`
	Points   float64   `json:"points"`
}

func (s *Server) leaderboard(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	ctx := r.Context()

	var table leaderboardTeamPointsTable

	/*
	 * Scan for team names and team totals
	 */

	totals, err := s.database.TeamPointsTotal(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to scan team points total", "err", err)

		writeError(w, http.StatusInternalServerError, err)
		return
	}

	teamIndices := make(map[string]int, len(totals))
	table.Teams = make([]string, len(totals))
	table.TeamTotals = make([]float64, len(totals))
	for i, row := range totals {
		teamIndices[row.TeamName] = i
		table.Teams[i] = row.TeamName
		table.TeamTotals[i] = row.Points.Float64
	}

	/*
	 * Scan for team points breakdown
	 */

	points, err := s.database.TeamPointsEach(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to scan team points each", "err", err)

		writeError(w, http.StatusInternalServerError, err)
		return
	}

	table.TeamPoints = make([][]teamPoints, len(table.Teams))
	for _, row := range points {
		ti, ok := teamIndices[row.TeamName]
		if !ok {
			continue
		}
		table.TeamPoints[ti] = append(table.TeamPoints[ti], teamPoints{
			Reason: row.Reason,
			Points: row.Points.Float64,
		})
	}

	/*
	 * Scan for team members
	 */

	members, err := s.database.ListTeamAndMembers(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to scan team members", "err", err)

		writeError(w, http.StatusInternalServerError, err)
		return
	}

	table.TeamMembers = make([][]string, len(table.Teams))
	for _, member := range members {
		ti, ok := teamIndices[member.TeamName]
		if !ok {
			continue
		}
		table.TeamMembers[ti] = append(table.TeamMembers[ti], member.Username)
	}

	/*
	 * Scan for week of code solves
	 */

	weekOfCodeSolves, err := s.database.ListAllCorrectSubmissions(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to scan week of code solves", "err", err)

		writeError(w, http.StatusInternalServerError, err)
		return
	}

	table.WeekOfCodeSolves = make([][]int8, len(table.Teams))
	for i := range table.WeekOfCodeSolves {
		table.WeekOfCodeSolves[i] = make([]int8, s.config.Problems.TotalProblems())
	}
	for _, row := range weekOfCodeSolves {
		day, part2, ok := s.parseProblemID(row.ProblemID)
		if !ok {
			continue
		}

		ti, ok := teamIndices[row.TeamName]
		if !ok {
			continue
		}

		p := int8(1)
		if part2 {
			p = 2
		}
		table.WeekOfCodeSolves[ti][day.index()] = p
	}

	rows, err := s.database.TeamPointsHistory(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to scan team points history", "err", err)

		writeError(w, http.StatusInternalServerError, err)
		return
	}

	events := make([]leaderboardTeamPointsEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, leaderboardTeamPointsEvent{
			TeamName: row.TeamName,
			AddedAt:  row.AddedAt.Time(),
			Points:   row.Points,
		})
	}

	s.renderTemplate(w, "leaderboard", leaderboardPageData{
		ComponentContext: frontend.ComponentContext{
			TeamName: u.TeamName,
			Username: u.Username,
		},
		StartedAt: s.problems.StartedAt(),
		Table:     table,
		Events:    events,
	})
}

func (s *Server) parseProblemID(id string) (day problemDay, part2 bool, ok bool) {
	switch {
	case strings.HasSuffix(id, "/part1"):
		part2 = false
		id = strings.TrimSuffix(id, "/part1")
	case strings.HasSuffix(id, "/part2"):
		part2 = true
		id = strings.TrimSuffix(id, "/part2")
	default:
		return
	}
	for i, problem := range s.config.Problems.Problems() {
		if problem.ID == id {
			day = problemDay(i + 1)
			ok = true
			return
		}
	}
	return
}
