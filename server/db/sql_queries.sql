-- name: CreateTeam :one
INSERT INTO teams (team_name, invite_code) VALUES (?, ?) RETURNING *;

-- name: JoinTeam :one
REPLACE INTO team_members (team_name, user_name, is_leader) VALUES (?, ?, ?) RETURNING *;

-- name: LeaveTeam :one
DELETE FROM team_members WHERE team_name = ? AND user_name = ? RETURNING *;

-- name: IsLeader :one
SELECT is_leader FROM team_members WHERE team_name = ? AND user_name = ?;

-- name: RecordSubmission :one
INSERT INTO team_submit_attempts (team_name, problem_id, correct) VALUES (?, ?, ?) RETURNING *;

-- name: AddPoints :one
UPDATE teams SET points = points + ? WHERE team_name = ? RETURNING *;

-- name: ListTeams :many
SELECT * FROM teams;

-- name: FindTeamWithInviteCode :one
SELECT * FROM teams WHERE invite_code = ? AND accepting_members = TRUE;

-- name: FindTeam :one
SELECT * FROM teams WHERE team_name = ?;
