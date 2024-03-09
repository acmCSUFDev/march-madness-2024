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
	r.Get("/{problemDay}", s.viewProblem)

	r.Group(func(r chi.Router) {
		r.Use(s.requireAuth)
		r.Get("/{problemDay}/input", s.viewProblemInput)
		r.With(parseForm).Post("/{problemDay}/submit", s.submitProblem)
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
		Problems: s.problems,
	})
}

type problemPageData struct {
	frontend.ComponentContext
	Problem       problem.Problem
	Day           problemDay
	PointsPerPart int
	SolvedPart1   bool
	SolvedPart2   bool
}

func (s *Server) viewProblem(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	ctx := r.Context()

	p, day, err := s.getProblemFromRequest(r)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	var p1solves, p2solves int64
	if u.TeamName != "" {
		p1solves, _ = s.database.HasSolved(ctx, db.HasSolvedParams{
			TeamName:  u.TeamName,
			ProblemID: s.problemID(day, false),
		})
		p2solves, _ = s.database.HasSolved(ctx, db.HasSolvedParams{
			TeamName:  u.TeamName,
			ProblemID: s.problemID(day, true),
		})
	}

	s.renderTemplate(w, "problem", problemPageData{
		ComponentContext: frontend.ComponentContext{
			TeamName: u.TeamName,
			Username: u.Username,
		},
		Problem:       p,
		Day:           day,
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

	input, err := p.Input(ctx, problem.StringToSeed(u.TeamName))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, input)
}

type problemResultPageData struct {
	frontend.ComponentContext
	Day          problemDay
	Cooldown     time.Duration
	CooldownTime time.Time
	Correct      bool
}

func (s *Server) submitProblem(w http.ResponseWriter, r *http.Request) {
	u := getAuthentication(r)
	ctx := r.Context()

	p, day, err := s.getProblemFromRequest(r)
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

	problemID := s.problemID(day, data.Part == 2)
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
		seed := problem.StringToSeed(u.TeamName)

		var answer int64
		switch data.Part {
		case 1:
			answer, err = p.Part1Solution(ctx, seed)
		case 2:
			answer, err = p.Part2Solution(ctx, seed)
		default:
			err = fmt.Errorf("invalid part %d", data.Part)
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		correct = answer == data.Answer

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
		Day:          day,
		Correct:      correct,
		Cooldown:     cooldown,
		CooldownTime: cooldownTime,
	})
}

type problemDay int

func (p problemDay) index() int {
	return int(p) - 1
}

func (s *Server) getProblemFromRequest(r *http.Request) (problem.Problem, problemDay, error) {
	day, err := strconv.Atoi(chi.URLParam(r, "problemDay"))
	if err != nil {
		return nil, 0, err
	}

	p := s.problems.Problem(day - 1)
	if p == nil {
		return nil, 0, fmt.Errorf("problem %d not found", day)
	}

	return p, problemDay(day), nil
}

func (s *Server) problemID(day problemDay, part2 bool) string {
	problem := s.config.Problems.Problem(day.index())
	if problem == nil {
		return ""
	}
	id := problem.ID()
	if part2 {
		id += "/part2"
	} else {
		id += "/part1"
	}
	return id
}
