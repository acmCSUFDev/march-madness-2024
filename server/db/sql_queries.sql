-- name: CreateTeam :one
INSERT INTO teams (team_name, invite_code) VALUES (?, ?) RETURNING *;

-- name: JoinTeam :one
INSERT INTO team_members (team_name, user_name, is_leader) VALUES (?, ?, ?) RETURNING *;

-- name: RecordSubmission :one
INSERT INTO team_submit_attempts (team_name, problem_id, correct) VALUES (?, ?, ?) RETURNING *;

-- name: AddPoints :one
UPDATE teams SET points = points + ? WHERE team_name = ? RETURNING *;

-- name: ListTeams :many
SELECT * FROM teams;
