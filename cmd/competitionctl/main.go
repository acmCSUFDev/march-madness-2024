package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/spf13/pflag"
	"libdb.so/february-frenzy/internal/config"
	"libdb.so/february-frenzy/server/db"
)

var (
	configPath = "config.json"
	verbose    = false
)

func main() {
	pflag.StringVarP(&configPath, "config", "c", configPath, "path to config file")
	pflag.BoolVarP(&verbose, "verbose", "v", verbose, "enable verbose logging")
	pflag.Usage = func() {
		log.SetFlags(0)
		log.Println("Usage:")
		log.Println("  competitionctl [flags] <command> [args...]")
		log.Println()
		log.Println("Flags:")
		pflag.PrintDefaults()
		log.Println()
		log.Println("Commands:")
		for _, cmd := range commandsHelp {
			log.Println("  " + cmd)
		}
		log.Println()
	}
	pflag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatalln(err)
	}
}

type Context struct {
	context.Context
	config   *config.Config
	database *db.Database
}

func run(ctx context.Context) error {
	config, err := config.ParseFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	db, err := db.Open(config.Paths.Database)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	context := Context{
		Context:  ctx,
		config:   config,
		database: db,
	}
	switch pflag.Arg(0) {
	case "hackathon-submissions":
		return hackathonSubmissions(context)
	case "hackathon-set-winner":
		return hackathonSetWinner(context)
	case "teams-list":
		return teamsList(context)
	case "teams-delete":
		return teamsDelete(context)
	case "points-list":
		return pointsList(context)
	default:
		return fmt.Errorf("unknown command: %q", pflag.Arg(0))
	}
}

var commandsHelp = []string{
	"hackathon-submissions                  list hackathon submissions",
	"hackathon-set-winner [team] [0|1|2|3]  set hackathon winner (0 = no winner)",
	"teams-list                             list teams",
	"teams-delete [team]                    delete team",
	"points-list                            list points",
}

func hackathonSubmissions(ctx Context) error {
	submissions, err := ctx.database.HackathonSubmissions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get hackathon submissions: %w", err)
	}
	return dumpAsJSON(submissions)
}

func hackathonSetWinner(ctx Context) error {
	team := pflag.Arg(1)

	place, err := strconv.Atoi(pflag.Arg(2))
	if err != nil {
		return fmt.Errorf("invalid place: %w", err)
	}
	if place < 0 || place > 3 {
		return fmt.Errorf("invalid place: %d", place)
	}

	return ctx.database.SetHackathonWinner(ctx, db.SetHackathonWinnerParams{
		TeamName: team,
		WonRank: sql.NullInt64{
			Int64: int64(place),
			Valid: place != 0,
		},
	})
}

func teamsList(ctx Context) error {
	teams, err := ctx.database.ListTeams(ctx)
	if err != nil {
		return fmt.Errorf("failed to list teams: %w", err)
	}
	return dumpAsJSON(teams)
}

func teamsDelete(ctx Context) error {
	team := pflag.Arg(1)

	t, err := ctx.database.DropTeam(ctx, team)
	if err != nil {
		return fmt.Errorf("failed to drop team: %w", err)
	}

	fmt.Printf("dropped team %q created at %v\n", t.TeamName, t.CreatedAt)
	return nil
}

func pointsList(ctx Context) error {
	points, err := ctx.database.TeamPointsHistory(ctx)
	if err != nil {
		return fmt.Errorf("failed to get points: %w", err)
	}
	for _, pt := range points {
		fmt.Printf("%v: %s, %s: %.0f\n", pt.AddedAt.In(time.Local), pt.Reason, pt.TeamName, pt.Points)
	}
	return nil
}

func dumpAsJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
