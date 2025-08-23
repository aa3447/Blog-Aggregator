package main

import (
	"database/sql"
	"fmt"
	"os"
	"home/aa3447/workspace/github.com/aa3447/blog-aggregator/internal/config"
	"home/aa3447/workspace/github.com/aa3447/blog-aggregator/internal/database"
	
	_ "github.com/lib/pq"
)

func main() {
	configFile, err := config.ReadConfig()
	if err != nil {
		fmt.Println("Error reading config:", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", configFile.Db_url)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		os.Exit(1)
	}
	defer db.Close()
	
	newState := &config.State{
		Db: database.New(db),
		CurrentState: configFile,
	}
	cmds := &config.Commands{}
	cmds.Init()

	
	commandArgs := os.Args[1:]
	if len(commandArgs) == 0 {
		fmt.Println("No command provided")
		os.Exit(1)
	}

	err = cmds.Run(newState, config.Command{
		Name: commandArgs[0],
		Args: commandArgs[1:],})
	if err != nil {
		fmt.Println("Error running command:", err)
		os.Exit(1)
	}
	
	fmt.Println("Command executed successfully")
}