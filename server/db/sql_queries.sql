-- name: CreateTeam :one
INSERT INTO teams (team_name, invite_code) VALUES (?, ?) RETURNING *;

-- name: JoinTeam :one
INSERT INTO team_members (team_name, user_name, is_leader) VALUES (?, ?, ?) RETURNING *;

-- name: LeaveTeam :exec
DELETE FROM team_members WHERE team_name = ? AND user_name = ?;

-- name: IsLeader :one
SELECT is_leader FROM team_members WHERE team_name = ? AND user_name = ?;

-- name: RecordSubmission :one
INSERT INTO team_submit_attempts (team_name, submitted_by, problem_id, correct) VALUES (?, ?, ?, ?) RETURNING *;

-- name: HasSolved :one
SELECT COUNT(*) FROM team_submit_attempts WHERE team_name = ? AND problem_id = ? AND correct = TRUE;

-- name: ListSubmissions :many
SELECT * FROM team_submit_attempts WHERE team_name = ? AND problem_id = ?
	ORDER BY submitted_at ASC;

-- name: ListAllCorrectSubmissions :many
SELECT *
	FROM team_submit_attempts
	WHERE correct = TRUE
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

-- name: RemovePointsByReason :many
DELETE FROM team_points WHERE team_name = ? AND reason = ? RETURNING *;

-- name: RemovePointsByTime :one
DELETE FROM team_points WHERE team_name = ? AND added_at = ? RETURNING *;

-- name: TeamPointsTotal :many
SELECT team_name, SUM(points) AS points
	FROM team_points
	GROUP BY team_name
	ORDER BY COALESCE(SUM(points), 0) DESC;

-- name: TeamPointsEach :many
SELECT team_name, reason, points
	FROM team_points
	GROUP BY team_name, reason;

-- name: TeamPointsHistory :many
SELECT *
	FROM (
		SELECT team_name, added_at, points FROM team_points
		UNION ALL
		SELECT team_name, MIN(joined_at) AS added_at, 0 AS points
			FROM team_members
			GROUP BY team_name
	) AS history
	ORDER BY added_at ASC;

-- name: ListTeams :many
SELECT team_name, created_at, accepting_members FROM teams;

-- name: ListTeamAndMembers :many
SELECT team_name, user_name FROM team_members ORDER BY joined_at ASC;

-- name: FindTeamWithInviteCode :one
SELECT team_name, created_at, accepting_members FROM teams WHERE invite_code = ? AND accepting_members = TRUE;

-- name: FindTeam :one
SELECT team_name, created_at, accepting_members FROM teams WHERE team_name = ?;

-- name: TeamInviteCode :one
SELECT invite_code FROM teams WHERE team_name = ?;

-- name: ListTeamMembers :many
SELECT * FROM team_members WHERE team_name = ? ORDER BY joined_at ASC;

-- name: DropTeam :one
DELETE FROM teams WHERE team_name = ? RETURNING *;

-- name: SetHackathonSubmission :exec
REPLACE INTO hackathon_submissions (team_name, project_url, project_description, category) VALUES (?, ?, ?, ?);

-- name: SetHackathonWinner :exec
UPDATE hackathon_submissions SET won_rank = ? WHERE team_name = ?;

-- name: HackathonSubmissions :many
SELECT * FROM hackathon_submissions ORDER BY submitted_at ASC;

-- name: HackathonSubmission :one
SELECT * FROM hackathon_submissions WHERE team_name = ?;

-- name: HackathonWinners :many
SELECT * FROM hackathon_submissions WHERE won_rank IS NOT NULL ORDER BY won_rank ASC;
