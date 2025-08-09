package main

import (
	"fmt"
	"github/wycliff-ochieng/cmd/api"
	"github/wycliff-ochieng/internal/config"
	"log"
)

func main() {
	fmt.Println("Main function:::")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("ERROR loading ENV variables:%v", err)
	}

	server := api.NewAPIServer(":3000", cfg)
	server.Run()

}
