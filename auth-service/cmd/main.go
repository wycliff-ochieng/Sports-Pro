package main

import (
	"fmt"
	"log"
	cmd "sports/authservice/cmd/api"
	"sports/authservice/internal/config"
)

func main() {
	fmt.Println("<<<<working on user service")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load:%v", err)
	}
	server := cmd.NewAPIServer(":8000", cfg)
	server.Run()
}
