package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	blue   = color.New(color.FgCyan).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
)

// Add this struct to store test details
type TestDetail struct {
	name     string
	duration string
	status   string // "PASS", "FAIL", or "SKIP"
	subtests []string
}

type TestStats struct {
	totalTests    int
	passedTests   int
	failedTests   int
	skippedTests  int
	totalDuration time.Duration
	mainTests     map[string]bool
	testDetails   []TestDetail // Add this field
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	stats := TestStats{
		mainTests:   make(map[string]bool),
		testDetails: make([]TestDetail, 0), // Initialize testDetails
	}

	var currentTest *TestDetail

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "=== RUN"):
			parts := strings.SplitN(line, "   ", 2)
			if len(parts) == 2 {
				testName := parts[1]
				if !strings.Contains(testName, "/") {
					stats.mainTests[testName] = true
					stats.totalTests++
					currentTest = &TestDetail{
						name:     testName,
						subtests: make([]string, 0),
					}
				} else if currentTest != nil {
					currentTest.subtests = append(currentTest.subtests, testName)
				}
				fmt.Printf("%s   %s\n", blue(parts[0]), parts[1])
			}

		case strings.HasPrefix(line, "--- PASS"):
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				testInfo := strings.Split(parts[1], " ")
				testName := testInfo[0]
				duration := ""
				if len(testInfo) > 1 {
					duration = testInfo[1]
				}

				if !strings.Contains(testName, "/") {
					stats.passedTests++
					if currentTest != nil {
						currentTest.status = "PASS"
						currentTest.duration = duration
						stats.testDetails = append(stats.testDetails, *currentTest)
						currentTest = nil
					}
				}
				fmt.Printf("%s: %s\n", green(parts[0]), green(parts[1]))
			}

		case strings.HasPrefix(line, "--- FAIL"):
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				// Only count main test functions
				testName := parts[1]
				if !strings.Contains(strings.Split(testName, " ")[0], "/") {
					stats.failedTests++
				}
				fmt.Printf("%s: %s\n", red(parts[0]), red(parts[1]))
			} else {
				fmt.Println(line)
			}

		case strings.HasPrefix(line, "--- SKIP"):
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				// Only count main test functions
				testName := parts[1]
				if !strings.Contains(strings.Split(testName, " ")[0], "/") {
					stats.skippedTests++
				}
				fmt.Printf("%s: %s\n", yellow(parts[0]), yellow(parts[1]))
			} else {
				fmt.Println(line)
			}

		default:
			// Try to parse duration
			if strings.Contains(line, "ok") && strings.Contains(line, "s)") {
				if d, err := time.ParseDuration(strings.Split(strings.Split(line, "(")[1], ")")[0]); err == nil {
					stats.totalDuration += d
				}
			}
			fmt.Println(line)
		}
	}

	// Print summary
	fmt.Println("\n" + blue("Test Summary:"))
	fmt.Printf("Total Tests: %d\n", stats.totalTests)
	fmt.Printf("Passed: %s (%d%%)\n", green(stats.passedTests), calculatePercentage(stats.passedTests, stats.totalTests))
	if stats.failedTests > 0 {
		fmt.Printf("Failed: %s (%d%%)\n", red(stats.failedTests), calculatePercentage(stats.failedTests, stats.totalTests))
	}
	if stats.skippedTests > 0 {
		fmt.Printf("Skipped: %s (%d%%)\n", yellow(stats.skippedTests), calculatePercentage(stats.skippedTests, stats.totalTests))
	}
	fmt.Printf("Total Duration: %s\n", blue(stats.totalDuration.Round(time.Millisecond)))

	// Print detailed test results
	fmt.Println("\n" + blue("Detailed Test Results:"))
	for _, test := range stats.testDetails {
		switch test.status {
		case "PASS":
			fmt.Printf("%s %s %s\n", green("✔"), test.name, blue(test.duration))
		case "FAIL":
			fmt.Printf("%s %s %s\n", red("✘"), test.name, blue(test.duration))
		case "SKIP":
			fmt.Printf("%s %s %s\n", yellow("⚠"), test.name, blue(test.duration))
		}

		// Print subtests if any
		for _, subtest := range test.subtests {
			fmt.Printf("  %s\n", subtest)
		}
	}

	// Print final status
	fmt.Println()
	if stats.failedTests > 0 {
		fmt.Printf("%s\n", red("✘ Some tests failed"))
		os.Exit(1)
	} else {
		fmt.Printf("%s\n", green("✔ All tests passed"))
	}
}

func calculatePercentage(part, total int) int {
	if total == 0 {
		return 0
	}
	return (part * 100) / total
}
