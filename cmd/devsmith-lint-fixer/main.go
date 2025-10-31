// Package main implements the devsmith-lint-fixer CLI tool.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	fixPtr := flag.Bool("fix", false, "Apply fixes to files")
	reportPtr := flag.Bool("report", false, "Generate linting report")
	pathPtr := flag.String("path", ".", "Directory or file to analyze")
	flag.Parse()

	if !*fixPtr && !*reportPtr {
		fmt.Println("devsmith-lint-fixer: Automated linting tool for DevSmith platform")
		fmt.Println("Usage: devsmith-lint-fixer -fix -path ./internal")
		fmt.Println("       devsmith-lint-fixer -report -path ./internal")
		os.Exit(1)
	}

	lf := NewLintFixer(*pathPtr)

	if *reportPtr {
		report := lf.AnalyzeDirectory()
		fmt.Println(report.String())
	}

	if *fixPtr {
		errors, fixes := lf.FixDirectory()
		fmt.Printf("Found %d issues, applied %d fixes\n", errors, fixes)
	}
}
