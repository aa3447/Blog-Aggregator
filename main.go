package main

import (
	"fmt"
	"home/aa3447/workspace/github.com/aa3447/blog-aggregator/internal/config"
)

func main() {
	configFile, err := config.ReadConfig()
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}
	fmt.Println("Database URL:", configFile.Db_url)
	fmt.Println("Current user name:", configFile.Current_user_name)
	
	err = configFile.SetUser("aa3447")
	if err != nil {
		fmt.Println("Error setting user:", err)
		return
	}
	configFile, err = config.ReadConfig()
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}
	fmt.Println("Database URL:", configFile.Db_url)
	fmt.Println("Current user name:", configFile.Current_user_name)
}