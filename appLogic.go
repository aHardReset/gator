package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aHardReset/gator/internal/config"
	"github.com/aHardReset/gator/internal/database"
	"github.com/google/uuid"
)

type state struct {
	db     *database.Queries
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (cmds *commands) run(s *state, cmd command) error {
	call, ok := cmds.commands[cmd.name]
	if !ok {
		return fmt.Errorf("%s not recognized as command\n", cmd.name)
	}
	err := call(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (cmds *commands) register(name string, f func(*state, command) error) {
	exists := false
	if _, ok := cmds.commands[name]; ok {
		exists = true
	}
	cmds.commands[name] = f
	if exists {
		fmt.Printf("Warning: %s overwritted", name)
	}
}

func handleLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("Username required to login")
	}
	user, err := s.db.GetUserByName(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}
	err = s.config.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("%s logged in\n", s.config.CurrentUserName)
	return nil
}

func handleRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("Username required to login")
	}
	uuid, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	params := database.CreateUserParams{
		ID:        uuid,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}
	user, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		return err
	}
	err = s.config.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("User '%s' created\n", user.Name)
	return nil
}

func handleListUsers(s *state, cmd command) error {
	users, err := s.db.ListUsers(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}
	for _, user := range users {
		fmt.Print(user.Name)
		if user.IsLoggedIn == true {
			fmt.Print(" (current)")
		}
		fmt.Println()
	}
	return nil
}

func handleReset(s *state, _ command) error {
	if err := s.db.FlushTables(context.Background()); err != nil {
		return err
	}
	fmt.Printf("Tables flushed \n")
	return nil
}
