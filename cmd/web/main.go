package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"dungeons/internal/llm"
	"dungeons/internal/web"
)

func main() {
	// Parse flags
	port := flag.Int("port", 8085, "Port to listen on")
	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	// Load OpenRouter configuration (OPENROUTER_API_KEY is required; the legacy
	// ANTHROPIC_API_KEY is only consulted by the not-yet-migrated nested agents).
	cfg, err := llm.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		fmt.Fprintln(os.Stderr, "Please set it in your .envrc file or export it")
		os.Exit(1)
	}

	// Create server config
	webCfg := web.Config{
		Port:         *port,
		LLMConfig:    cfg,
		TemplatesDir: "web/templates",
		StaticDir:    "web/static",
		Debug:        *debug,
	}

	// Create and start server
	server := web.NewServer(webCfg)

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		fmt.Println("\nShutting down server...")
		server.Stop()
		os.Exit(0)
	}()

	// Run server
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
