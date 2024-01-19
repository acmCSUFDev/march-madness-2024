CREATE TABLE teams (
	team_name TEXT PRIMARY KEY,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	invite_code TEXT NOT NULL,
	accepting_members BOOLEAN NOT NULL DEFAULT TRUE,
	points REAL NOT NULL DEFAULT 0);

CREATE TABLE team_members (
	team_name TEXT NOT NULL,
	user_name TEXT NOT NULL,
	joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	is_leader BOOLEAN NOT NULL DEFAULT FALSE,
	PRIMARY KEY (team_name, user_name),
	FOREIGN KEY (team_name) REFERENCES teams (team_name));

CREATE TABLE team_submit_attempts (
	team_name TEXT NOT NULL,
	problem_id TEXT NOT NULL,
	submitted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	correct BOOLEAN NOT NULL,
	FOREIGN KEY (team_name) REFERENCES teams (team_name));

--------------------------------- NEW VERSION ---------------------------------

-- Add UNIQUE into teams.invite_code
CREATE UNIQUE INDEX teams_invite_code_idx ON teams (invite_code);


--------------------------------- NEW VERSION ---------------------------------

-- Add UNIQUE into team_members.user_name so that we can't have duplicate users
CREATE UNIQUE INDEX team_members_user_name_idx ON team_members (user_name);

--------------------------------- NEW VERSION ---------------------------------

-- Replace the teams.points column with a separate table that tracks points
-- over time.
ALTER TABLE teams DROP COLUMN points;

CREATE TABLE team_points (
	team_name TEXT NOT NULL,
	added_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	points REAL NOT NULL,
	reason TEXT NOT NULL,
	PRIMARY KEY (team_name, added_at),
	FOREIGN KEY (team_name) REFERENCES teams (team_name));

CREATE INDEX team_points_team_name_idx ON team_points (team_name);
