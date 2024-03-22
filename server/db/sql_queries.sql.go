// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: sql_queries.sql

package db

import (
	"context"
	"database/sql"
)

const addPoints = `-- name: AddPoints :one
INSERT INTO team_points (team_name, points, reason) VALUES (?, ?, ?) RETURNING team_name, added_at, points, reason
`

type AddPointsParams struct {
	TeamName string
	Points   float64
	Reason   string
}

func (q *Queries) AddPoints(ctx context.Context, arg AddPointsParams) (TeamPoint, error) {
	row := q.queryRow(ctx, q.addPointsStmt, addPoints, arg.TeamName, arg.Points, arg.Reason)
	var i TeamPoint
	err := row.Scan(
		&i.TeamName,
		&i.AddedAt,
		&i.Points,
		&i.Reason,
	)
	return i, err
}

const countIncorrectSubmissions = `-- name: CountIncorrectSubmissions :one
SELECT COUNT(*) FROM team_submit_attempts WHERE team_name = ? AND problem_id = ? AND correct = FALSE
`

type CountIncorrectSubmissionsParams struct {
	TeamName  string
	ProblemID string
}

func (q *Queries) CountIncorrectSubmissions(ctx context.Context, arg CountIncorrectSubmissionsParams) (int64, error) {
	row := q.queryRow(ctx, q.countIncorrectSubmissionsStmt, countIncorrectSubmissions, arg.TeamName, arg.ProblemID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createTeam = `-- name: CreateTeam :one
INSERT INTO teams (team_name, invite_code) VALUES (?, ?) RETURNING team_name, created_at, invite_code, accepting_members
`

type CreateTeamParams struct {
	TeamName   string
	InviteCode string
}

func (q *Queries) CreateTeam(ctx context.Context, arg CreateTeamParams) (Team, error) {
	row := q.queryRow(ctx, q.createTeamStmt, createTeam, arg.TeamName, arg.InviteCode)
	var i Team
	err := row.Scan(
		&i.TeamName,
		&i.CreatedAt,
		&i.InviteCode,
		&i.AcceptingMembers,
	)
	return i, err
}

const dropTeam = `-- name: DropTeam :one
DELETE FROM teams WHERE team_name = ? RETURNING team_name, created_at, invite_code, accepting_members
`

func (q *Queries) DropTeam(ctx context.Context, teamName string) (Team, error) {
	row := q.queryRow(ctx, q.dropTeamStmt, dropTeam, teamName)
	var i Team
	err := row.Scan(
		&i.TeamName,
		&i.CreatedAt,
		&i.InviteCode,
		&i.AcceptingMembers,
	)
	return i, err
}

const findTeam = `-- name: FindTeam :one
SELECT team_name, created_at, accepting_members FROM teams WHERE team_name = ?
`

type FindTeamRow struct {
	TeamName         string
	CreatedAt        DateTime
	AcceptingMembers bool
}

func (q *Queries) FindTeam(ctx context.Context, teamName string) (FindTeamRow, error) {
	row := q.queryRow(ctx, q.findTeamStmt, findTeam, teamName)
	var i FindTeamRow
	err := row.Scan(&i.TeamName, &i.CreatedAt, &i.AcceptingMembers)
	return i, err
}

const findTeamWithInviteCode = `-- name: FindTeamWithInviteCode :one
SELECT team_name, created_at, accepting_members FROM teams WHERE invite_code = ? AND accepting_members = TRUE
`

type FindTeamWithInviteCodeRow struct {
	TeamName         string
	CreatedAt        DateTime
	AcceptingMembers bool
}

func (q *Queries) FindTeamWithInviteCode(ctx context.Context, inviteCode string) (FindTeamWithInviteCodeRow, error) {
	row := q.queryRow(ctx, q.findTeamWithInviteCodeStmt, findTeamWithInviteCode, inviteCode)
	var i FindTeamWithInviteCodeRow
	err := row.Scan(&i.TeamName, &i.CreatedAt, &i.AcceptingMembers)
	return i, err
}

const hackathonSubmission = `-- name: HackathonSubmission :one
SELECT team_name, submitted_at, project_url, project_description, category, won_rank FROM hackathon_submissions WHERE team_name = ?
`

func (q *Queries) HackathonSubmission(ctx context.Context, teamName string) (HackathonSubmission, error) {
	row := q.queryRow(ctx, q.hackathonSubmissionStmt, hackathonSubmission, teamName)
	var i HackathonSubmission
	err := row.Scan(
		&i.TeamName,
		&i.SubmittedAt,
		&i.ProjectUrl,
		&i.ProjectDescription,
		&i.Category,
		&i.WonRank,
	)
	return i, err
}

const hackathonSubmissions = `-- name: HackathonSubmissions :many
SELECT team_name, submitted_at, project_url, project_description, category, won_rank FROM hackathon_submissions ORDER BY submitted_at ASC
`

func (q *Queries) HackathonSubmissions(ctx context.Context) ([]HackathonSubmission, error) {
	rows, err := q.query(ctx, q.hackathonSubmissionsStmt, hackathonSubmissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []HackathonSubmission
	for rows.Next() {
		var i HackathonSubmission
		if err := rows.Scan(
			&i.TeamName,
			&i.SubmittedAt,
			&i.ProjectUrl,
			&i.ProjectDescription,
			&i.Category,
			&i.WonRank,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const hackathonWinners = `-- name: HackathonWinners :many
SELECT team_name, submitted_at, project_url, project_description, category, won_rank FROM hackathon_submissions WHERE won_rank IS NOT NULL ORDER BY won_rank ASC
`

func (q *Queries) HackathonWinners(ctx context.Context) ([]HackathonSubmission, error) {
	rows, err := q.query(ctx, q.hackathonWinnersStmt, hackathonWinners)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []HackathonSubmission
	for rows.Next() {
		var i HackathonSubmission
		if err := rows.Scan(
			&i.TeamName,
			&i.SubmittedAt,
			&i.ProjectUrl,
			&i.ProjectDescription,
			&i.Category,
			&i.WonRank,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const hasSolved = `-- name: HasSolved :one
SELECT COUNT(*) FROM team_submit_attempts WHERE team_name = ? AND problem_id = ? AND correct = TRUE
`

type HasSolvedParams struct {
	TeamName  string
	ProblemID string
}

func (q *Queries) HasSolved(ctx context.Context, arg HasSolvedParams) (int64, error) {
	row := q.queryRow(ctx, q.hasSolvedStmt, hasSolved, arg.TeamName, arg.ProblemID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const isLeader = `-- name: IsLeader :one
SELECT is_leader FROM team_members WHERE team_name = ? AND user_name = ?
`

type IsLeaderParams struct {
	TeamName string
	Username string
}

func (q *Queries) IsLeader(ctx context.Context, arg IsLeaderParams) (bool, error) {
	row := q.queryRow(ctx, q.isLeaderStmt, isLeader, arg.TeamName, arg.Username)
	var is_leader bool
	err := row.Scan(&is_leader)
	return is_leader, err
}

const joinTeam = `-- name: JoinTeam :one
INSERT INTO team_members (team_name, user_name, is_leader) VALUES (?, ?, ?) RETURNING team_name, user_name, joined_at, is_leader
`

type JoinTeamParams struct {
	TeamName string
	Username string
	IsLeader bool
}

func (q *Queries) JoinTeam(ctx context.Context, arg JoinTeamParams) (TeamMember, error) {
	row := q.queryRow(ctx, q.joinTeamStmt, joinTeam, arg.TeamName, arg.Username, arg.IsLeader)
	var i TeamMember
	err := row.Scan(
		&i.TeamName,
		&i.Username,
		&i.JoinedAt,
		&i.IsLeader,
	)
	return i, err
}

const lastSubmissionTime = `-- name: LastSubmissionTime :one
SELECT submitted_at FROM team_submit_attempts
	WHERE team_name = ? AND problem_id = ?
	ORDER BY submitted_at DESC
	LIMIT 1
`

type LastSubmissionTimeParams struct {
	TeamName  string
	ProblemID string
}

func (q *Queries) LastSubmissionTime(ctx context.Context, arg LastSubmissionTimeParams) (DateTime, error) {
	row := q.queryRow(ctx, q.lastSubmissionTimeStmt, lastSubmissionTime, arg.TeamName, arg.ProblemID)
	var submitted_at DateTime
	err := row.Scan(&submitted_at)
	return submitted_at, err
}

const leaveTeam = `-- name: LeaveTeam :exec
DELETE FROM team_members WHERE team_name = ? AND user_name = ?
`

type LeaveTeamParams struct {
	TeamName string
	Username string
}

func (q *Queries) LeaveTeam(ctx context.Context, arg LeaveTeamParams) error {
	_, err := q.exec(ctx, q.leaveTeamStmt, leaveTeam, arg.TeamName, arg.Username)
	return err
}

const listAllCorrectSubmissions = `-- name: ListAllCorrectSubmissions :many
SELECT team_name, problem_id, submitted_at, correct, submitted_by
	FROM team_submit_attempts
	WHERE correct = TRUE
	ORDER BY submitted_at ASC
`

func (q *Queries) ListAllCorrectSubmissions(ctx context.Context) ([]TeamSubmitAttempt, error) {
	rows, err := q.query(ctx, q.listAllCorrectSubmissionsStmt, listAllCorrectSubmissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TeamSubmitAttempt
	for rows.Next() {
		var i TeamSubmitAttempt
		if err := rows.Scan(
			&i.TeamName,
			&i.ProblemID,
			&i.SubmittedAt,
			&i.Correct,
			&i.SubmittedBy,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listSubmissions = `-- name: ListSubmissions :many
SELECT team_name, problem_id, submitted_at, correct, submitted_by FROM team_submit_attempts WHERE team_name = ? AND problem_id = ?
	ORDER BY submitted_at ASC
`

type ListSubmissionsParams struct {
	TeamName  string
	ProblemID string
}

func (q *Queries) ListSubmissions(ctx context.Context, arg ListSubmissionsParams) ([]TeamSubmitAttempt, error) {
	rows, err := q.query(ctx, q.listSubmissionsStmt, listSubmissions, arg.TeamName, arg.ProblemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TeamSubmitAttempt
	for rows.Next() {
		var i TeamSubmitAttempt
		if err := rows.Scan(
			&i.TeamName,
			&i.ProblemID,
			&i.SubmittedAt,
			&i.Correct,
			&i.SubmittedBy,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listTeamAndMembers = `-- name: ListTeamAndMembers :many
SELECT team_name, user_name FROM team_members ORDER BY joined_at ASC
`

type ListTeamAndMembersRow struct {
	TeamName string
	Username string
}

func (q *Queries) ListTeamAndMembers(ctx context.Context) ([]ListTeamAndMembersRow, error) {
	rows, err := q.query(ctx, q.listTeamAndMembersStmt, listTeamAndMembers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListTeamAndMembersRow
	for rows.Next() {
		var i ListTeamAndMembersRow
		if err := rows.Scan(&i.TeamName, &i.Username); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listTeamMembers = `-- name: ListTeamMembers :many
SELECT team_name, user_name, joined_at, is_leader FROM team_members WHERE team_name = ? ORDER BY joined_at ASC
`

func (q *Queries) ListTeamMembers(ctx context.Context, teamName string) ([]TeamMember, error) {
	rows, err := q.query(ctx, q.listTeamMembersStmt, listTeamMembers, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TeamMember
	for rows.Next() {
		var i TeamMember
		if err := rows.Scan(
			&i.TeamName,
			&i.Username,
			&i.JoinedAt,
			&i.IsLeader,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listTeams = `-- name: ListTeams :many
SELECT team_name, created_at, accepting_members FROM teams
`

type ListTeamsRow struct {
	TeamName         string
	CreatedAt        DateTime
	AcceptingMembers bool
}

func (q *Queries) ListTeams(ctx context.Context) ([]ListTeamsRow, error) {
	rows, err := q.query(ctx, q.listTeamsStmt, listTeams)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListTeamsRow
	for rows.Next() {
		var i ListTeamsRow
		if err := rows.Scan(&i.TeamName, &i.CreatedAt, &i.AcceptingMembers); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const recordSubmission = `-- name: RecordSubmission :one
INSERT INTO team_submit_attempts (team_name, submitted_by, problem_id, correct) VALUES (?, ?, ?, ?) RETURNING team_name, problem_id, submitted_at, correct, submitted_by
`

type RecordSubmissionParams struct {
	TeamName    string
	SubmittedBy sql.NullString
	ProblemID   string
	Correct     bool
}

func (q *Queries) RecordSubmission(ctx context.Context, arg RecordSubmissionParams) (TeamSubmitAttempt, error) {
	row := q.queryRow(ctx, q.recordSubmissionStmt, recordSubmission,
		arg.TeamName,
		arg.SubmittedBy,
		arg.ProblemID,
		arg.Correct,
	)
	var i TeamSubmitAttempt
	err := row.Scan(
		&i.TeamName,
		&i.ProblemID,
		&i.SubmittedAt,
		&i.Correct,
		&i.SubmittedBy,
	)
	return i, err
}

const removePointsByReason = `-- name: RemovePointsByReason :many
DELETE FROM team_points WHERE team_name = ? AND reason = ? RETURNING team_name, added_at, points, reason
`

type RemovePointsByReasonParams struct {
	TeamName string
	Reason   string
}

func (q *Queries) RemovePointsByReason(ctx context.Context, arg RemovePointsByReasonParams) ([]TeamPoint, error) {
	rows, err := q.query(ctx, q.removePointsByReasonStmt, removePointsByReason, arg.TeamName, arg.Reason)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TeamPoint
	for rows.Next() {
		var i TeamPoint
		if err := rows.Scan(
			&i.TeamName,
			&i.AddedAt,
			&i.Points,
			&i.Reason,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const removePointsByTime = `-- name: RemovePointsByTime :one
DELETE FROM team_points WHERE team_name = ? AND added_at = ? RETURNING team_name, added_at, points, reason
`

type RemovePointsByTimeParams struct {
	TeamName string
	AddedAt  DateTime
}

func (q *Queries) RemovePointsByTime(ctx context.Context, arg RemovePointsByTimeParams) (TeamPoint, error) {
	row := q.queryRow(ctx, q.removePointsByTimeStmt, removePointsByTime, arg.TeamName, arg.AddedAt)
	var i TeamPoint
	err := row.Scan(
		&i.TeamName,
		&i.AddedAt,
		&i.Points,
		&i.Reason,
	)
	return i, err
}

const setHackathonSubmission = `-- name: SetHackathonSubmission :exec
REPLACE INTO hackathon_submissions (team_name, project_url, project_description, category) VALUES (?, ?, ?, ?)
`

type SetHackathonSubmissionParams struct {
	TeamName           string
	ProjectUrl         string
	ProjectDescription sql.NullString
	Category           string
}

func (q *Queries) SetHackathonSubmission(ctx context.Context, arg SetHackathonSubmissionParams) error {
	_, err := q.exec(ctx, q.setHackathonSubmissionStmt, setHackathonSubmission,
		arg.TeamName,
		arg.ProjectUrl,
		arg.ProjectDescription,
		arg.Category,
	)
	return err
}

const setHackathonWinner = `-- name: SetHackathonWinner :exec
UPDATE hackathon_submissions SET won_rank = ? WHERE team_name = ?
`

type SetHackathonWinnerParams struct {
	WonRank  sql.NullInt64
	TeamName string
}

func (q *Queries) SetHackathonWinner(ctx context.Context, arg SetHackathonWinnerParams) error {
	_, err := q.exec(ctx, q.setHackathonWinnerStmt, setHackathonWinner, arg.WonRank, arg.TeamName)
	return err
}

const teamInviteCode = `-- name: TeamInviteCode :one
SELECT invite_code FROM teams WHERE team_name = ?
`

func (q *Queries) TeamInviteCode(ctx context.Context, teamName string) (string, error) {
	row := q.queryRow(ctx, q.teamInviteCodeStmt, teamInviteCode, teamName)
	var invite_code string
	err := row.Scan(&invite_code)
	return invite_code, err
}

const teamPointsEach = `-- name: TeamPointsEach :many
SELECT team_name, reason, SUM(points) AS points
	FROM team_points
	GROUP BY team_name, reason
`

type TeamPointsEachRow struct {
	TeamName string
	Reason   string
	Points   sql.NullFloat64
}

func (q *Queries) TeamPointsEach(ctx context.Context) ([]TeamPointsEachRow, error) {
	rows, err := q.query(ctx, q.teamPointsEachStmt, teamPointsEach)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TeamPointsEachRow
	for rows.Next() {
		var i TeamPointsEachRow
		if err := rows.Scan(&i.TeamName, &i.Reason, &i.Points); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const teamPointsHistory = `-- name: TeamPointsHistory :many
SELECT team_name, added_at, points
	FROM (
		SELECT team_name, added_at, points FROM team_points
		UNION ALL
		SELECT team_name, MIN(joined_at) AS added_at, 0 AS points
			FROM team_members
			GROUP BY team_name
	) AS history
	ORDER BY added_at ASC
`

type TeamPointsHistoryRow struct {
	TeamName string
	AddedAt  DateTime
	Points   float64
}

func (q *Queries) TeamPointsHistory(ctx context.Context) ([]TeamPointsHistoryRow, error) {
	rows, err := q.query(ctx, q.teamPointsHistoryStmt, teamPointsHistory)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TeamPointsHistoryRow
	for rows.Next() {
		var i TeamPointsHistoryRow
		if err := rows.Scan(&i.TeamName, &i.AddedAt, &i.Points); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const teamPointsTotal = `-- name: TeamPointsTotal :many
SELECT team_name, SUM(points) AS points
	FROM team_points
	GROUP BY team_name
	ORDER BY COALESCE(SUM(points), 0) DESC
`

type TeamPointsTotalRow struct {
	TeamName string
	Points   sql.NullFloat64
}

func (q *Queries) TeamPointsTotal(ctx context.Context) ([]TeamPointsTotalRow, error) {
	rows, err := q.query(ctx, q.teamPointsTotalStmt, teamPointsTotal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TeamPointsTotalRow
	for rows.Next() {
		var i TeamPointsTotalRow
		if err := rows.Scan(&i.TeamName, &i.Points); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
