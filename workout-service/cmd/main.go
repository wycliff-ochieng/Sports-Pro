package main

import (
	"fmt"
	"log"

	"github.com/wycliff-ochieng/cmd/api"
	"github.com/wycliff-ochieng/internal/config"
)

func main() {
	fmt.Println("Spinning up server")

	cfg, err := config.Load()
	if err != nil {
		log.Printf("error loading configurations due to: %s", err)
	}

	server := api.NewAPIServer(":3000", cfg)
	server.Run()
}
