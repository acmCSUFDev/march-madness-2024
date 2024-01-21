package server

import (
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"
	"libdb.so/february-frenzy/server/frontend"
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
	Reasons          []string
	Teams            []string
	TeamMembers      [][]string
	Points           [][]float64
	WeekOfCodeSolves [][]int8 // list of teams, each containing N days
}

type leaderboardTeamPointsEvent struct {
	TeamName string    `json:"team_name"`
	AddedAt  time.Time `json:"added_at"`
	Reason   string    `json:"reason"`
	Points   float64   `json:"points"`
}

func (s *Server) leaderboard(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	ctx := r.Context()

	var table leaderboardTeamPointsTable

	pointsRows, err := s.database.TeamPoints(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	membersRows, err := s.database.ListTeamAndMembers(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	membersForTeam := func(teamName string) []string {
		var members []string
		for _, row := range membersRows {
			if row.TeamName == teamName {
				members = append(members, row.Username)
			}
		}
		return members
	}

	for _, row := range pointsRows {
		// table.Reasons = append(table.Reasons, row.Reason)
		table.Teams = append(table.Teams, row.TeamName)
	}

	table.Reasons = []string{"week of code", "hackathon"}
	table.Teams = slices.Compact(table.Teams)

	table.TeamMembers = make([][]string, len(table.Teams))
	for i, teamName := range table.Teams {
		table.TeamMembers[i] = membersForTeam(teamName)
	}

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
			s.logger.WarnContext(ctx,
				"leaderboard: unexpected team or reason",
				"team", row.TeamName,
				"reason", row.Reason)
		}
	}

	weekOfCodeSolves, err := s.database.ListAllCorrectSubmissions(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	table.WeekOfCodeSolves = make([][]int8, len(table.Teams))
	for i := range table.WeekOfCodeSolves {
		table.WeekOfCodeSolves[i] = make([]int8, s.config.Problems.TotalProblems())
	}
	for _, row := range weekOfCodeSolves {
		day, part2, ok := s.parseProblemID(row.ProblemID)
		if ok {
			// We assume that part 2 is always solved after part 1.
			i := slices.Index(table.Teams, row.TeamName)
			p := int8(1)
			if part2 {
				p = 2
			}
			table.WeekOfCodeSolves[i][day-1] = p
		}
	}

	rows, err := s.database.TeamPointsHistory(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	events := make([]leaderboardTeamPointsEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, leaderboardTeamPointsEvent{
			TeamName: row.TeamName,
			AddedAt:  row.AddedAt,
			Reason:   row.Reason,
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
