package main

import (
"context"
"flag"
"fmt"
"os"
"os/signal"
"syscall"

"wazmeow/internal/app"
"wazmeow/internal/app/config"

"github.com/rs/zerolog/log"
)

var (
versionFlag = flag.Bool("version", false, "Display version information and exit")
)

const version = "2.0.0"

func init() {
flag.Parse()

if *versionFlag {
fmt.Printf("WazMeow version %s\n", version)
os.Exit(0)
}
}

func main() {
// Load configuration
cfg, err := config.Load()
if err != nil {
log.Fatal().Err(err).Msg("Failed to load configuration")
}

// Setup logger
cfg.SetupLogger()

log.Info().Str("version", version).Msg("Starting WazMeow")

// Create application container
container, err := app.NewContainer(cfg)
if err != nil {
log.Fatal().Err(err).Msg("Failed to create application container")
}
defer container.Close()

// Create and start server
server := app.NewServer(container)

// Setup graceful shutdown
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Handle shutdown signals
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

go func() {
<-sigChan
log.Info().Msg("Shutdown signal received")
cancel()
}()

// Start server
if err := server.Start(ctx); err != nil {
log.Fatal().Err(err).Msg("Server failed to start")
}

log.Info().Msg("WazMeow stopped gracefully")
}
