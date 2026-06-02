package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"xui-next/internal/agent"
	"xui-next/internal/version"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("xui-pro agent " + version.String())
		return
	}

	masterURL := os.Getenv("XUI_MASTER")
	if masterURL == "" {
		masterURL = "http://127.0.0.1:8080"
	}

	a := agent.New(masterURL, agent.NodeFromEnv())
	if err := a.Register(); err != nil {
		log.Fatalf("register node: %v", err)
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		if err := a.Heartbeat(); err != nil {
			log.Printf("heartbeat failed: %v", err)
		}
		<-ticker.C
	}
}
