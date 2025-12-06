package main

import (
	"fmt"
	"log"

	"github.com/wycliff-ochieng/cmd/api"
	"github.com/wycliff-ochieng/internal/config"
)

func main() {
	fmt.Println("Main function:::")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("ERROR loading ENV variables:%v", err)
	}

	server := api.NewAPIServer(":7000", cfg)
	server.Run()

}
