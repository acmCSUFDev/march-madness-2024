package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"dev.acmcsuf.com/march-madness-2024/internal/config"
	"dev.acmcsuf.com/march-madness-2024/server/db"
	"github.com/spf13/pflag"
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
	case "hackathon-set-winner":
		return hackathonSetWinner(context)
	case "award-points":
		return awardPoints(context)
	case "list-teams":
		return teamsList(context)
	case "delete-team":
		return teamsDelete(context)
	case "invite-code":
		return teamInviteCode(context)
	case "list-points":
		return pointsList(context)
	default:
		pflag.Usage()
		return fmt.Errorf("missing or invalid command %q", pflag.Arg(0))
	}
}

var commandsHelp = []string{
	"hackathon-set-winner [team] [0|1|2|3] [points] set hackathon winner (0 = no winner)",
	"award-points [team] [points] [reason]          award or remove points",
	"list-teams                                     list teams",
	"delete-team [team]                             delete team",
	"invite-code [team]                             get invite code for team",
	"list-points                                    list points",
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

	var points float64
	if place != 0 {
		var err error

		points, err = strconv.ParseFloat(pflag.Arg(3), 64)
		if err != nil {
			return fmt.Errorf("invalid points: %w", err)
		}
	}

	if err := ctx.database.SetHackathonWinner(ctx, db.SetHackathonWinnerParams{
		TeamName: team,
		WonRank: sql.NullInt64{
			Int64: int64(place),
			Valid: place != 0,
		},
	}); err != nil {
		return fmt.Errorf("failed to set hackathon winner: %w", err)
	}

	if place == 0 {
		_, err := ctx.database.RemovePointsByReason(ctx, db.RemovePointsByReasonParams{
			TeamName: team,
			Reason:   "hackathon",
		})
		if err != nil {
			return fmt.Errorf("failed to remove points: %w", err)
		}
	} else {
		_, err := ctx.database.AddPoints(ctx, db.AddPointsParams{
			TeamName: team,
			Points:   points,
			Reason:   "hackathon",
		})
		if err != nil {
			return fmt.Errorf("failed to add points: %w", err)
		}
	}

	return nil
}

func awardPoints(ctx Context) error {
	team := pflag.Arg(1)

	points, err := strconv.ParseFloat(pflag.Arg(2), 64)
	if err != nil {
		return fmt.Errorf("invalid points: %w", err)
	}

	reason := pflag.Arg(3)
	if reason == "hackathon" {
		return fmt.Errorf("hackathon-specific points must be set with hackathon-set-winner")
	}

	deletedPoints, deleteErr := ctx.database.RemovePointsByReason(ctx, db.RemovePointsByReasonParams{
		TeamName: team,
		Reason:   reason,
	})
	if deleteErr != nil {
		return fmt.Errorf("failed to remove points: %w", deleteErr)
	}

	if points > 0 {
		if len(deletedPoints) > 0 {
			log.Printf("team %q has %d points to delete for reason %q\n", team, len(deletedPoints), reason)
		}
		_, err = ctx.database.AddPoints(ctx, db.AddPointsParams{
			TeamName: team,
			Points:   points,
			Reason:   reason,
		})
		if err != nil {
			return fmt.Errorf("failed to add points: %w", err)
		}
	} else {
		if len(deletedPoints) == 0 {
			log.Printf("team %q has no points to delete for reason %q\n", team, reason)
		}
	}

	return nil
}

func teamsList(ctx Context) error {
	teams, err := ctx.database.ListTeams(ctx)
	if err != nil {
		return fmt.Errorf("failed to list teams: %w", err)
	}

	teamPoints, err := ctx.database.TeamPointsTotal(ctx)
	if err != nil {
		return fmt.Errorf("failed to get team points: %w", err)
	}

	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Team\tMembers\tPoints\tCreated At\n")
	fmt.Fprintf(w, "----\t-------\t------\t----------\n")

	for _, team := range teams {
		var membersString string
		if members, err := ctx.database.ListTeamMembers(ctx, team.TeamName); err != nil {
			membersString = fmt.Sprintf("(error: %v)", err)
		} else {
			strs := make([]string, len(members))
			for i, member := range members {
				strs[i] = member.Username
				if member.IsLeader {
					strs[i] += " (leader)"
				}
			}
			membersString = strings.Join(strs, ", ")
		}

		pointsIx := slices.IndexFunc(teamPoints, func(r db.TeamPointsTotalRow) bool {
			return r.TeamName == team.TeamName
		})
		points := teamPoints[pointsIx].Points.Float64

		fmt.Fprintf(w,
			"%s\t%s\t%.0f\t%v\n",
			team.TeamName, membersString, points, team.CreatedAt.Time().In(time.Local))
	}

	w.Flush()
	fmt.Print(b.String())

	return nil
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

func teamInviteCode(ctx Context) error {
	team := pflag.Arg(1)

	code, err := ctx.database.TeamInviteCode(ctx, team)
	if err != nil {
		return fmt.Errorf("failed to get invite code: %w", err)
	}

	fmt.Println(code)
	return nil
}

func pointsList(ctx Context) error {
	points, err := ctx.database.TeamPointsHistory(ctx)
	if err != nil {
		return fmt.Errorf("failed to get points: %w", err)
	}
	for _, pt := range points {
		fmt.Printf(
			"%v: %s +%f\n",
			pt.AddedAt.Time().In(time.Local),
			pt.TeamName,
			math.Floor(pt.Points))
	}
	return nil
}
