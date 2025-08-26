package config

import (
	"fmt"
	"context"
	"database/sql"
	"time"
	"log"
	
	"github.com/google/uuid"
	
	"home/aa3447/workspace/github.com/aa3447/blog-aggregator/internal/database"
	"home/aa3447/workspace/github.com/aa3447/blog-aggregator/internal/rss"
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

func (c *Commands) Init() {
	c.registerCommand("login", handlerLogin)
	c.registerCommand("register", handlerRegisterUser)
	c.registerCommand("reset", handlerReset)
	c.registerCommand("users", handlerGetUsers)
	c.registerCommand("agg", handlerFetchFeed)
	c.registerCommand("addfeed", handlerAddFeed)
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

func handlerRegisterUser(s *State, cmd Command) error{
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

	currentUser, err := s.Db.CreateUser(ctx, newUser)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}

	fmt.Println("User registered:", currentUser.Name)
	log.Printf("User Info: %#v\n", currentUser)
	
	handlerLogin(s, Command{Name: "login", Args: []string{currentUser.Name}})
	return nil
}

func handlerAddFeed(s *State, cmd Command) error{
	if len(cmd.Args) < 2 {
		return fmt.Errorf("register command requires name and url argument")
	}
	if s.CurrentState.Current_user_name == "" {
		return fmt.Errorf("no user logged in, please login first")
	}
	
	name := cmd.Args[0]
	url := cmd.Args[1]
	userName := s.CurrentState.Current_user_name
	ctx := context.Background()

	existingUser, err := s.Db.GetUserByName(ctx, userName)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error checking existing user: %v", err)
	}

	newFeed := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    uuid.NullUUID{UUID: existingUser.ID, Valid: true},
	}

	currentFeed, err := s.Db.CreateFeed(ctx, newFeed)
	if err != nil {
		return fmt.Errorf("error creating feed: %v", err)
	}

	fmt.Println("User feed:", currentFeed.Url)
	log.Printf("Feed Info: %#v\n", currentFeed)
	
	return nil
}

func handlerGetUsers(s *State, cmd Command) error {
	ctx := context.Background()
	users, err := s.Db.GetAllUsers(ctx)
	
	if err != nil {
		return fmt.Errorf("error retrieving users: %v", err)
	}
	if len(users) == 0 {
		fmt.Println("No users found.")
		return nil
	}
	
	fmt.Println("Registered Users:")
	for _, user := range users {
		var userData string
		userData += fmt.Sprintf("- %s (ID: %s, CreatedAt: %s)", user.Name, user.ID, user.CreatedAt.Format(time.RFC3339))
		
		if user.UpdatedAt != user.CreatedAt {
			userData += fmt.Sprintf("  UpdatedAt: %s", user.UpdatedAt.Format(time.RFC3339))
		}
		if s.CurrentState.Current_user_name == user.Name {
			userData += " (current)"
		}
		
		fmt.Println(userData)
	}
	
	return nil
}

func handlerFetchFeed(s *State, cmd Command) error {
	ctx := context.Background()
	xml, err := rss.FetchFeed(ctx, cmd.Args[0])
	if err != nil {
		return fmt.Errorf("error fetching feed: %v", err)
	}
	fmt.Println(xml)
	for _, item := range xml.Channel.Item {
		fmt.Println(item)	
	}
	
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

