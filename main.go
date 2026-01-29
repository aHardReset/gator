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

func main() {
	godotenv.Load()
	s := loadState()
	cmds := commands{
		commands: map[string]func(*state, command) error{},
	}
	cmds.register("login", handleLogin)
	cmds.register("register", handleRegister)
	cmds.register("reset", handleReset)
	cmds.register("users", handleListUsers)
	cmds.register("agg", handleAgg)
	cmds.register("addfeed", handleAddFeed)
	cmds.register("feeds", handleListFeeds)

	allArgs := os.Args
	if len(allArgs) < 2 {
		fmt.Println("Not enough args")
		os.Exit(1)
	}
	cmd := command{
		name: allArgs[1],
		args: allArgs[2:],
	}
	err := cmds.run(&s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
