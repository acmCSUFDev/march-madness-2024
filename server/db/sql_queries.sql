-- name: CreateTeam :one
INSERT INTO teams (team_name, invite_code) VALUES (?, ?) RETURNING *;

-- name: JoinTeam :one
REPLACE INTO team_members (team_name, user_name, is_leader) VALUES (?, ?, ?) RETURNING *;

-- name: LeaveTeam :one
DELETE FROM team_members WHERE team_name = ? AND user_name = ? RETURNING *;

-- name: IsLeader :one
SELECT is_leader FROM team_members WHERE team_name = ? AND user_name = ?;

-- name: RecordSubmission :one
INSERT INTO team_submit_attempts (team_name, submitted_by, problem_id, correct) VALUES (?, ?, ?, ?) RETURNING *;

-- name: HasSolved :one
SELECT COUNT(*) FROM team_submit_attempts WHERE team_name = ? AND problem_id = ? AND correct = TRUE;

-- name: ListSubmissions :many
SELECT * FROM team_submit_attempts WHERE team_name = ? AND problem_id = ?
	ORDER BY submitted_at ASC;

-- name: CountIncorrectSubmissions :one
SELECT COUNT(*) FROM team_submit_attempts WHERE team_name = ? AND problem_id = ? AND correct = FALSE;

-- name: LastSubmissionTime :one
SELECT submitted_at FROM team_submit_attempts
	WHERE team_name = ? AND problem_id = ?
	ORDER BY submitted_at DESC
	LIMIT 1;

-- name: AddPoints :one
INSERT INTO team_points (team_name, points, reason) VALUES (?, ?, ?) RETURNING *;

-- name: TeamPoints :many
SELECT
		teams.team_name,
		team_points.reason,
		SUM(team_points.points) AS points
	FROM team_points
	RIGHT JOIN teams ON teams.team_name = team_points.team_name
	GROUP BY teams.team_name, team_points.reason
	ORDER BY COALESCE(SUM(team_points.points), 0) DESC;

-- name: TeamPointsHistory :many
SELECT * FROM team_points ORDER BY added_at ASC;

-- name: ListTeams :many
SELECT * FROM teams;

-- name: ListTeamAndMembers :many
SELECT * FROM team_members;

-- name: FindTeamWithInviteCode :one
SELECT * FROM teams WHERE invite_code = ? AND accepting_members = TRUE;

-- name: FindTeam :one
SELECT * FROM teams WHERE team_name = ?;
