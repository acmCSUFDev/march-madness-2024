package server

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"libdb.so/february-frenzy/server/db"
	"libdb.so/february-frenzy/server/frontend"
	"libdb.so/february-frenzy/server/problem"
)

func (s *Server) routeProblems(r chi.Router) {
	r.Get("/", s.listProblems)
	r.Get("/{problemID}", s.viewProblem)

	r.Group(func(r chi.Router) {
		r.Use(s.requireAuth)
		r.Get("/{problemID}/input", s.viewProblemInput)
		r.With(parseForm).Post("/{problemID}/submit", s.submitProblem)
	})
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
		Problems: s.problems.ProblemSet,
	})
}

type problemPageData struct {
	frontend.ComponentContext
	Problem       *problem.Problem
	ID            int
	PointsPerPart int
	SolvedPart1   bool
	SolvedPart2   bool
}

func (s *Server) viewProblem(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	ctx := r.Context()

	p, id, err := s.getProblemFromRequest(r)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	var p1solves, p2solves int64
	if u.TeamName != "" {
		p1solves, _ = s.database.HasSolved(ctx, db.HasSolvedParams{
			TeamName:  u.TeamName,
			ProblemID: s.problems.ProblemIDs[id] + "/part1",
		})
		p2solves, _ = s.database.HasSolved(ctx, db.HasSolvedParams{
			TeamName:  u.TeamName,
			ProblemID: s.problems.ProblemIDs[id] + "/part2",
		})
	}

	s.renderTemplate(w, "problem", problemPageData{
		ComponentContext: frontend.ComponentContext{
			TeamName: u.TeamName,
			Username: u.Username,
		},
		Problem:       p,
		ID:            id,
		PointsPerPart: problem.PointsPerPart,
		SolvedPart1:   p1solves > 0,
		SolvedPart2:   p2solves > 0,
	})
}

func (s *Server) viewProblemInput(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	ctx := r.Context()

	p, _, err := s.getProblemFromRequest(r)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	input, err := p.Input.GenerateInput(ctx, problem.StringToSeed(u.TeamName))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, input.Input)
}

type problemResultPageData struct {
	frontend.ComponentContext
	ID           int
	Cooldown     time.Duration
	CooldownTime time.Time
	Correct      bool
}

func (s *Server) submitProblem(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	ctx := r.Context()

	p, id, err := s.getProblemFromRequest(r)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	var data struct {
		Answer int64 `schema:"answer"`
		Part   int   `schema:"part"`
	}
	if err := decoder.Decode(&data, r.PostForm); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if data.Part != 1 && data.Part != 2 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid part %d", data.Part))
		return
	}

	problemID := s.problems.ProblemIDs[id] + fmt.Sprintf("/part%d", data.Part)

	var numSolves int64
	var numAttempts int64
	var lastAttempt time.Time

	err = s.database.Tx(func(q *db.Queries) (err error) {
		numSolves, err = s.database.HasSolved(ctx, db.HasSolvedParams{
			TeamName:  u.TeamName,
			ProblemID: problemID,
		})
		if err != nil {
			return fmt.Errorf("failed to check if solved: %w", err)
		}
		if numSolves > 0 {
			return fmt.Errorf("problem is already solved")
		}

		numAttempts, err = s.database.CountIncorrectSubmissions(ctx, db.CountIncorrectSubmissionsParams{
			TeamName:  u.TeamName,
			ProblemID: problemID,
		})
		if err != nil {
			return fmt.Errorf("failed to count incorrect submissions: %w", err)
		}

		if numAttempts > 0 {
			lastAttempt, err = s.database.LastSubmissionTime(ctx, db.LastSubmissionTimeParams{
				TeamName:  u.TeamName,
				ProblemID: problemID,
			})
			if err != nil {
				return fmt.Errorf("failed to get last submission time: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	now := time.Now()
	cooldownTime := problem.CalculateCooldownEnd(int(numAttempts), lastAttempt, now)

	cooldown := max(0, cooldownTime.Sub(now))
	var correct bool

	if cooldown == 0 {
		input, err := p.Input.GenerateInput(ctx, problem.StringToSeed(u.TeamName))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		switch data.Part {
		case 1:
			correct = input.Part1 == data.Answer
		case 2:
			correct = input.Part2 == data.Answer
		}

		err = s.database.Tx(func(q *db.Queries) error {
			_, err := s.database.RecordSubmission(ctx, db.RecordSubmissionParams{
				TeamName: u.TeamName,
				SubmittedBy: sql.NullString{
					String: u.Username,
					Valid:  true,
				},
				ProblemID: problemID,
				Correct:   correct,
			})
			if err != nil {
				return fmt.Errorf("failed to record submission: %w", err)
			}

			_, err = s.database.AddPoints(ctx, db.AddPointsParams{
				TeamName: u.TeamName,
				Points:   problem.PointsPerPart,
				Reason:   "week of code",
			})
			if err != nil {
				return fmt.Errorf("failed to add points: %w", err)
			}

			return nil
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}

	s.renderTemplate(w, "problem_result", problemResultPageData{
		ComponentContext: frontend.ComponentContext{
			TeamName: u.TeamName,
			Username: u.Username,
		},
		ID:           id,
		Correct:      correct,
		Cooldown:     cooldown,
		CooldownTime: cooldownTime,
	})
}

func (s *Server) getProblemFromRequest(r *http.Request) (*problem.Problem, int, error) {
	id, err := strconv.Atoi(chi.URLParam(r, "problemID"))
	if err != nil {
		return nil, 0, err
	}

	p := s.problems.Problem(id - 1)
	if p == nil {
		return nil, 0, fmt.Errorf("problem %d not found", id)
	}

	return p, id, nil
}
