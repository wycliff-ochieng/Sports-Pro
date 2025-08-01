package main

import (
	"fmt"
	"log"

	"github.com/wycliff-ochieng/cmd/api"
	"github.com/wycliff-ochieng/internal/config"
)

func main() {
	fmt.Println("<<<<working on authentication service")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load:%v", err)
	}
	server := api.NewAPIServer(":9000", cfg)
	server.Run()
}
