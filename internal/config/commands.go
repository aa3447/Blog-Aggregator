package config

import (
	"fmt"
	"context"
	"database/sql"
	"time"
	"log"
	"github.com/google/uuid"
	"home/aa3447/workspace/github.com/aa3447/blog-aggregator/internal/database"
)

type State struct {
	Db *database.Queries
	CurrentState *Config
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	CommandsMap map[string]func(*State, Command) error
}

func handlerLogin(s *State, cmd Command) error{
	if len(cmd.Args) == 0 {
		return fmt.Errorf("login command requires username argument")
	}
	
	user ,err := s.Db.GetUserByName(context.Background(), cmd.Args[0])
	if err != nil || user == (database.User{}) {
		return fmt.Errorf("user %s does not exist, please register first", cmd.Args[0])
	}	
	
	if s.CurrentState.Current_user_name == cmd.Args[0] {
		fmt.Println("User is already set to:", cmd.Args[0])
		return nil
	}
	
	err = s.CurrentState.SetUser(cmd.Args[0])
	if err != nil{
		return fmt.Errorf("error setting user: %v", err)
	}

	fmt.Println("User set to:", cmd.Args[0])
	return nil
}

func handlerRegister(s *State, cmd Command) error{
	if len(cmd.Args) == 0 {
		return fmt.Errorf("register command requires username argument")
	}
	userName := cmd.Args[0]
	ctx := context.Background()

	existingUser, err := s.Db.GetUserByName(ctx, userName)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error checking existing user: %v", err)
	}
	if existingUser != (database.User{}) {
		return fmt.Errorf("user %s already exists", userName)
	}

	newUser := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      userName,
	}

	_, err = s.Db.CreateUser(ctx, newUser)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}

	fmt.Println("User registered:", userName)
	log.Printf("User Info: %#v\n", newUser)
	
	handlerLogin(s, Command{Name: "login", Args: []string{userName}})
	return nil
}

func handlerReset(s *State, cmd Command) error {
	ctx := context.Background()
	err := s.Db.ResetUsers(ctx)
	if err != nil {
		return fmt.Errorf("error resetting users: %v", err)
	}
	
	s.CurrentState.Current_user_name = ""
	err = s.CurrentState.SetUser("")
	if err != nil {
		return fmt.Errorf("error saving config: %v", err)
	}

	fmt.Println("All users have been reset and current user cleared.")
	return nil
}

func (c *Commands) Run(s *State, cmd Command) error {
	if handler, exists := c.CommandsMap[cmd.Name]; exists {
		return handler(s, cmd)
	}
	return fmt.Errorf("unknown command: %s", cmd.Name)
}

func (c *Commands) registerCommand(name string, f func(*State, Command) error){
	if c.CommandsMap == nil {
		c.CommandsMap = make(map[string]func(*State, Command) error)
	}
	c.CommandsMap[name] = f
}

func (c *Commands) Init() {
	c.registerCommand("login", handlerLogin)
	c.registerCommand("register", handlerRegister)
	c.registerCommand("reset", handlerReset)
}