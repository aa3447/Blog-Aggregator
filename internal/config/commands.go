package config

import (
	"fmt"
)

type State struct {
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
		return fmt.Errorf("login command requires argument")
	}
	if s.CurrentState.Current_user_name == cmd.Args[0] {
		fmt.Println("User is already set to:", cmd.Args[0])
		return nil
	}
	
	err := s.CurrentState.SetUser(cmd.Args[0])
	if err != nil{
		return fmt.Errorf("error setting user: %v", err)
	}

	fmt.Println("User set to:", cmd.Args[0])
	return nil
}

func (c *Commands) Run(s *State, cmd Command) error {
	if handler, exists := c.CommandsMap[cmd.Name]; exists {
		return handler(s, cmd)
	}
	return fmt.Errorf("unknown command: %s", cmd.Name)
}

func (c *Commands) register(name string, f func(*State, Command) error){
	if c.CommandsMap == nil {
		c.CommandsMap = make(map[string]func(*State, Command) error)
	}
	c.CommandsMap[name] = f
}

func (c *Commands) Init() {
	c.register("login", handlerLogin)
}