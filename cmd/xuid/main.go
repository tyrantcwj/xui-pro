package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"xui-next/internal/master"
	"xui-next/internal/reality"
	"xui-next/internal/version"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("xui-pro master " + version.String())
		return
	}

	addr := env("XUI_LISTEN", ":8080")
	libraryPath := env("XUI_REALITY_LIBRARY", "reality/domains.json")

	lib, err := reality.Load(libraryPath)
	if err != nil {
		log.Fatalf("load reality library: %v", err)
	}

	srv := master.NewServer(master.NewStore(), lib)
	log.Printf("xuid master listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
