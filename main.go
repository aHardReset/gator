package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/aHardReset/gator/internal/config"
	"github.com/aHardReset/gator/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func loadState() state {

	c, err := config.Read()
	if err != nil {
		fmt.Println("%w", err)
		os.Exit(1)
	}
	db, err := sql.Open("postgres", c.DbUrl)
	if err != nil {
		fmt.Println("%w", err)
		os.Exit(1)
	}
	s := state{
		config: &c,
		db:     database.New(db),
	}
	return s
}

func loadCommand() command {
	allArgs := os.Args
	if len(allArgs) < 2 {
		fmt.Println("Not enough args")
		os.Exit(1)
	}
	return command{
		name: allArgs[1],
		args: allArgs[2:],
	}
}

func main() {
	godotenv.Load()
	s := loadState()
	cmd := loadCommand()
	cmds := commands{
		commands: map[string]func(*state, command) error{},
	}
	cmds.register("login", handleLogin)
	cmds.register("register", handleRegister)
	cmds.register("reset", handleReset)
	cmds.register("users", handleListUsers)
	cmds.register("agg", handleAgg)
	cmds.register("addfeed", middlewareLoggedIn(handleAddFeed))
	cmds.register("feeds", handleListFeeds)
	cmds.register("follow", middlewareLoggedIn(handleFollow))
	cmds.register("unfollow", middlewareLoggedIn(handleUnfollow))
	cmds.register("following", middlewareLoggedIn(handleListFollow))
	cmds.register("browse", middlewareLoggedIn(handleBrowse))

	err := cmds.run(&s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
