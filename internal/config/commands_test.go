package config

import (
	"database/sql"
	"fmt"
	"home/aa3447/workspace/github.com/aa3447/blog-aggregator/internal/database"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func Setup(t *testing.T) (*State, *Commands){
	configFile, err := ReadConfig()
	if err != nil {
		fmt.Println("Error reading config:", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", configFile.Db_url)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		os.Exit(1)
	}
	
	newState := &State{
		Db: database.New(db),
		CurrentState: configFile,
	}
	cmds := &Commands{}
	cmds.Init()

	return newState, cmds
}

func TestReset(t *testing.T) {
	state, commands := Setup(t)

	// First, register a user to ensure there's something to reset
	err := commands.Run(state, Command{Name: "register", Args: []string{"resetuser"}})
	if err != nil {
		t.Fatalf("Error registering user: %v", err)
	}

	// Now, reset the state for that user
	err = commands.Run(state, Command{Name: "reset", Args: []string{"resetuser"}})
	if err != nil {
		t.Errorf("Error resetting state: %v", err)
	}

	// Attempt to login again to ensure the user was removed
	err = commands.Run(state, Command{Name: "login", Args: []string{"resetuser"}})
	if err == nil {
		t.Errorf("Expected error logging in after reset, but got none")
	}
}

func TestCommandsNotNeedingLogin(t *testing.T) {
	state, commands := Setup(t)
	tests := []struct {
		name string
		cmd Command
		expectedError bool
		shouldReset bool
	}{
		{
			name: "Test users command with no users",
			cmd: Command{Name: "users", Args: []string{} },
			expectedError: true,
		},
		{
			name: "Test login with valid credentials not registered user",
			cmd: Command{Name: "login", Args: []string{"testuser"}},
			expectedError: true,
		},
		{
			name: "Test registering user",
			cmd: Command{Name: "register", Args: []string{"testuser"}},
			expectedError: false,
		},
		{
			name: "Test login with while logged in user",
			cmd: Command{Name: "login", Args: []string{"testuser"}},
			expectedError: true,
		},
		{
			name: "Test registering user that is already registered",
			cmd: Command{Name: "register", Args: []string{"testuser"}},
			expectedError: true,
		},
		{
			name: "Test users command",
			cmd: Command{Name: "users", Args: []string{} },
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := commands.Run(state, tt.cmd)
			if (err != nil) != tt.expectedError {
				t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
			}
		})
	}

		resetCmd := Command{Name: "reset", Args: []string{"testuser"}}
		err := commands.Run(state, resetCmd)
		if err != nil {
			t.Errorf("Error resetting state: %v", err)
		}
}

func TestCommandsNeedingLogin(t *testing.T) {
	state, commands := Setup(t)
	// First, register and login a user to ensure we have a logged-in state
	err := commands.Run(state, Command{Name: "register", Args: []string{"testuser"}})
	if err != nil {
		t.Fatalf("Error registering user: %v", err)
	}

	tests := []struct {
		name string
		cmd Command
		expectedError bool
		shouldReset bool
	}{
		{	
			name: "Test list feeds with no feeds",
			cmd: Command{Name: "feeds", Args: []string{}},
			expectedError: true,
		},
		{	
			name: "Test list followed feeds with no feeds",
			cmd: Command{Name: "following", Args: []string{}},
			expectedError: true,
		},
		{	
			name: "Test add feed",
			cmd: Command{Name: "addfeed", Args: []string{"Hacker News", "https://news.ycombinator.com/rss"}},
			expectedError: false,
		},
		{	
			name: "Test add same feed",
			cmd: Command{Name: "addfeed", Args: []string{"Hacker News", "https://news.ycombinator.com/rss"}},
			expectedError: true,
		},
		{	
			name: "Test list feeds",
			cmd: Command{Name: "feeds", Args: []string{}},
			expectedError: false,
		},
		{	
			name: "Test list followed feeds",
			cmd: Command{Name: "following", Args: []string{}},
			expectedError: false,
		},
		{	
			name: "Test unfollow feeds",
			cmd: Command{Name: "unfollow", Args: []string{"https://news.ycombinator.com/rss"}},
			expectedError: false,
		},
		{	
			name: "Test unfollow same feed again",
			cmd: Command{Name: "following", Args: []string{}},
			expectedError: true,
		},
	}	
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := commands.Run(state, tt.cmd)
			if (err != nil) != tt.expectedError {
				t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
			}
		})
	}

	resetCmd := Command{Name: "reset", Args: []string{"testuser"}}
	err = commands.Run(state, resetCmd)
	if err != nil {
		t.Errorf("Error resetting state: %v", err)
	}

}

