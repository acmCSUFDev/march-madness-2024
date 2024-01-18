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
