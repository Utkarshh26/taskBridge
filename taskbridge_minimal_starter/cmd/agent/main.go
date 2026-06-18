package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taskbridge/internal/agent"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "TaskBridge server URL")
	agentID := flag.String("id", "", "agent identifier (assigned on registration if empty)")
	capabilities := flag.String("capabilities", "http_check,tcp_check,file_exists,checksum,copy_file,write_file,wait", "comma-separated job capabilities")
	pollInterval := flag.Duration("poll-interval", 3*time.Second, "job polling interval")
	heartbeatInterval := flag.Duration("heartbeat-interval", 10*time.Second, "heartbeat interval")
	flag.Parse()

	caps := agent.ParseCapabilities(*capabilities)
	client := agent.NewClient(agent.Config{
		ServerURL:    *serverURL,
		AgentID:      *agentID,
		Capabilities: caps,
	}, nil)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	registered, err := client.Register(ctx, agent.Hostname(), caps)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("TaskBridge agent registered: id=%s hostname=%s\n", registered.ID, registered.Hostname)

	if err := client.Run(ctx, *pollInterval, *heartbeatInterval); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
	fmt.Println("agent stopped")
	os.Exit(0)
}
