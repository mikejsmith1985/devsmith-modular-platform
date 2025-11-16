package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/metrics"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	collector := metrics.NewCollector()
	command := os.Args[1]

	switch command {
	case "test":
		handleTestCommand(collector, os.Args[2:])
	case "deploy":
		handleDeployCommand(collector, os.Args[2:])
	case "cert":
		handleCertCommand(collector, os.Args[2:])
	case "violation":
		handleViolationCommand(collector, os.Args[2:])
	case "health":
		handleHealthCommand(collector, os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: metrics <command> [args...]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  test <passed> <failed> <duration_ms>          - Record test run\n")
	fmt.Fprintf(os.Stderr, "  deploy <service> <success> <duration_ms>      - Record deployment\n")
	fmt.Fprintf(os.Stderr, "  cert <success>                                - Record certificate generation\n")
	fmt.Fprintf(os.Stderr, "  violation <rule> <severity>                   - Record rule violation\n")
	fmt.Fprintf(os.Stderr, "  health <service> <available> <response_ms>    - Record service health\n")
}

func handleTestCommand(collector *metrics.Collector, args []string) {
	if len(args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: metrics test <passed> <failed> <duration_ms>\n")
		os.Exit(1)
	}

	passed, _ := strconv.Atoi(args[0])
	failed, _ := strconv.Atoi(args[1])
	durationMs, _ := strconv.ParseFloat(args[2], 64)
	duration := time.Duration(durationMs) * time.Millisecond

	if err := collector.RecordTestRun(passed, failed, duration); err != nil {
		fmt.Fprintf(os.Stderr, "Error recording test run: %v\n", err)
		os.Exit(1)
	}

	total := passed + failed
	passRate := 0.0
	if total > 0 {
		passRate = float64(passed) / float64(total) * 100
	}
	fmt.Printf("✓ Recorded test run: %d/%d passed (%.1f%%)\n", passed, total, passRate)
}

func handleDeployCommand(collector *metrics.Collector, args []string) {
	if len(args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: metrics deploy <service> <success> <duration_ms>\n")
		os.Exit(1)
	}

	service := args[0]
	success := args[1] == "true" || args[1] == "1"
	durationMs, _ := strconv.ParseFloat(args[2], 64)
	duration := time.Duration(durationMs) * time.Millisecond

	if err := collector.RecordDeployment(service, success, duration); err != nil {
		fmt.Fprintf(os.Stderr, "Error recording deployment: %v\n", err)
		os.Exit(1)
	}

	status := "✓ success"
	if !success {
		status = "✗ failed"
	}
	fmt.Printf("✓ Recorded deployment: %s (%s, %.1fms)\n", service, status, durationMs)
}

func handleCertCommand(collector *metrics.Collector, args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: metrics cert <success>\n")
		os.Exit(1)
	}

	success := args[0] == "true" || args[0] == "1"

	if err := collector.RecordCertificateGeneration(success); err != nil {
		fmt.Fprintf(os.Stderr, "Error recording certificate: %v\n", err)
		os.Exit(1)
	}

	status := "✓ success"
	if !success {
		status = "✗ failed"
	}
	fmt.Printf("✓ Recorded certificate generation: %s\n", status)
}

func handleViolationCommand(collector *metrics.Collector, args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: metrics violation <rule> <severity>\n")
		os.Exit(1)
	}

	rule := args[0]
	severity := args[1]

	if err := collector.RecordRuleViolation(rule, severity); err != nil {
		fmt.Fprintf(os.Stderr, "Error recording violation: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Recorded rule violation: %s (%s)\n", rule, severity)
}

func handleHealthCommand(collector *metrics.Collector, args []string) {
	if len(args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: metrics health <service> <available> <response_ms>\n")
		os.Exit(1)
	}

	service := args[0]
	available := args[1] == "true" || args[1] == "1"
	responseMs, _ := strconv.ParseFloat(args[2], 64)
	responseTime := time.Duration(responseMs) * time.Millisecond

	if err := collector.RecordServiceHealth(service, available, responseTime); err != nil {
		fmt.Fprintf(os.Stderr, "Error recording health: %v\n", err)
		os.Exit(1)
	}

	status := "✓ available"
	if !available {
		status = "✗ unavailable"
	}
	fmt.Printf("✓ Recorded service health: %s (%s, %.1fms)\n", service, status, responseMs)
}
