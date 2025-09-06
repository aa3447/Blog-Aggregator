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
	c.registerCommand("feeds", handlerGetFeeds)
	c.registerCommand("addfeed", userLoggedInWrapper(handlerAddFeed))
	c.registerCommand("follow", userLoggedInWrapper(handlerFollowFeed))
	c.registerCommand("following", userLoggedInWrapper(handlerFollowing))
	c.registerCommand("unfollow", userLoggedInWrapper(handlerUnfollowFeed))
}

// Login sets the current user in the config if the user exists in the database.
// Requires one argument: username.
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


// RegisterUser creates a new user in the database and sets it as the current user in the config.
// Requires one argument: username.
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



// GetUsers retrieves and prints all users from the database, indicating the current user.
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
		userData += fmt.Sprintf("- %s (CreatedAt: %s)", user.Name, user.CreatedAt.Format(time.RFC3339))
		
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

// GetFeeds retrieves and prints all feeds from the database, including their creators.
func handlerGetFeeds(s *State, cmd Command) error {
	ctx := context.Background()
	feeds, err := s.Db.GetAllFeeds(ctx)

	if err != nil {
		return fmt.Errorf("error retrieving feeds: %v", err)
	}
	if len(feeds) == 0 {
		fmt.Println("No feeds found.")
		return nil
	}
	
	fmt.Println("Feeds:")
	for _, feed := range feeds {
		var userName string
		if !feed.UserID.Valid {
			return fmt.Errorf("feed %s has no associated user", feed.Name)
		}

		user, err := s.Db.GetUserByID(ctx, feed.UserID.UUID)
		if err != nil {
			return fmt.Errorf("error retrieving user for feed: %v", err)
		}
		userName = user.Name

		var feedData string
		feedData += fmt.Sprintf("- %s ( CreatedBy: %s, URL: %s, CreatedAt: %s )", feed.Name, userName, feed.Url, feed.CreatedAt.Format(time.RFC3339))
		
		if feed.UpdatedAt != feed.CreatedAt {
			feedData += fmt.Sprintf("  UpdatedAt: %s", feed.UpdatedAt.Format(time.RFC3339))
		}
		
		fmt.Println(feedData)
	}
	
	return nil
}

// FetchFeed fetches and prints the RSS feed from the specified URL.
// Requires one argument: feed URL.
func handlerFetchFeed(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("follow command requires feed URL argument")
	}

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

// userLoggedInWrapper is a helper function that ensures a user is logged in before executing the handler.
func userLoggedInWrapper(handler func(s *State, cmd Command, user database.User) error)  func(*State, Command) error {
	return func(s *State, cmd Command) error {
		if s.CurrentState.Current_user_name == "" {
			return fmt.Errorf("no user logged in, please login first")
		}
		
		userName := s.CurrentState.Current_user_name
		ctx := context.Background()

		existingUser, err := s.Db.GetUserByName(ctx, userName)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("error checking existing user: %v", err)
		}
		
		return handler(s, cmd, existingUser)
	}
}

// AddFeed creates a new feed in the database and automatically follows it for the current user.
// Requires two arguments: feed name and feed URL.
// Wrapped with userLoggedInWrapper to ensure a user is logged in.
func handlerAddFeed(s *State, cmd Command, user database.User) error{
	if len(cmd.Args) < 2 {
		return fmt.Errorf("register command requires name and url argument")
	}
	
	name := cmd.Args[0]
	url := cmd.Args[1]
	ctx := context.Background()


	newFeed := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
	}

	currentFeed, err := s.Db.CreateFeed(ctx, newFeed)
	if err != nil {
		return fmt.Errorf("error creating feed: %v", err)
	}

	err = handlerFollowFeed(s, Command{Name: "follow", Args: []string{url}}, user)
	if err != nil {
		return fmt.Errorf("error following feed after creation: %v", err)
	}

	fmt.Println("User feed:", currentFeed.Url)
	log.Printf("Feed Info: %#v\n", currentFeed)
	
	return nil
}

// FollowFeed creates a new feed follow relationship in the database for the current user and the specified feed URL.
// Requires one argument: feed URL.
// Wrapped with userLoggedInWrapper to ensure a user is logged in.
func handlerFollowFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("follow command requires feed URL argument")
	}
	
	url := cmd.Args[0]
	ctx := context.Background()


	existingFeed, err := s.Db.GetFeedByUrl(ctx, url)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error checking existing feeds: %v", err)
	}


	newFeedFollow := database.CreateFeedFollowParams{
		ID:		uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID:    uuid.NullUUID{UUID: existingFeed.ID, Valid: true},
	}

	currentFeedFollow, err := s.Db.CreateFeedFollow(ctx, newFeedFollow)
	if err != nil {
		return fmt.Errorf("error creating feedFollow: %v", err)
	}
	
	fmt.Printf("%s followed feed: %s\n", currentFeedFollow.Username, currentFeedFollow.FeedName)
	log.Printf("FeedFollow Info: %#v\n", currentFeedFollow)
	
	return nil
}

// Following lists all feeds followed by the current user.
// Wrapped with userLoggedInWrapper to ensure a user is logged in.
func handlerFollowing(s *State, cmd Command, user database.User) error {
	ctx := context.Background()

	followedFeeds, err := s.Db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("error retrieving followed feeds: %v", err)
	}
	if len(followedFeeds) == 0 {
		return fmt.Errorf("user %s is not following any feeds", user.Name)
	}

	for _, feed := range followedFeeds {
		fmt.Printf("- %s (URL: %s)\n", feed.FeedName, feed.FeedUrl)
	}
	
	return nil
}

// UnfollowFeed removes a feed follow relationship in the database for the current user and the specified feed URL.
// Requires one argument: feed URL.
// Wrapped with userLoggedInWrapper to ensure a user is logged in.
func handlerUnfollowFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("unfollow command requires feed URL argument")
	}
	
	url := cmd.Args[0]
	ctx := context.Background()

	existingFeed, err := s.Db.GetFeedByUrl(ctx, url)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error checking existing feeds: %v", err)
	}
	if existingFeed == (database.Feed{}) {
		return fmt.Errorf("feed with URL %s does not exist", url)
	}

	params := database.DeleteFeedFollowParams{
		UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID: uuid.NullUUID{UUID: existingFeed.ID, Valid: true},
	}

	err = s.Db.DeleteFeedFollow(ctx, params)
	if err != nil {
		return fmt.Errorf("error unfollowing feed: %v", err)
	}

	fmt.Printf("User %s unfollowed feed: %s\n", user.Name, existingFeed.Name)
	log.Printf("Unfollowed Feed Info: UserID=%s, FeedID=%s\n", user.ID, existingFeed.ID)
	return nil
}


// Reset removes all users from the database and clears the current user in the config.
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

