package utils

import (
	"fmt"
	"testing"
)

type TestReport struct {
	TotalTests  int
	PassedTests int
	FailedTests int
	TestResults map[string]struct {
		Passed bool
		Error  string
	}
}

var Report = TestReport{
	TestResults: make(map[string]struct {
		Passed bool
		Error  string
	}),
}

func RecordTest(t *testing.T, name string, passed bool, errMsg string) {
	Report.TotalTests++
	Report.TestResults[name] = struct {
		Passed bool
		Error  string
	}{
		Passed: passed,
		Error:  errMsg,
	}
	if passed {
		Report.PassedTests++
	} else {
		Report.FailedTests++
		t.Errorf("%s: %s", name, errMsg)
	}
}

func PrintReport() {
	fmt.Println("\n=== Test Report ===")
	fmt.Printf("Total Tests: %d\n", Report.TotalTests)
	fmt.Printf("Passed: %d\n", Report.PassedTests)
	fmt.Printf("Failed: %d\n", Report.FailedTests)
	fmt.Println("\nDetailed Results:")
	fmt.Println("----------------")
	for name, result := range Report.TestResults {
		status := "✓ PASS"
		if !result.Passed {
			status = "✗ FAIL"
			if result.Error != "" {
				status += fmt.Sprintf(" (%s)", result.Error)
			}
		}
		fmt.Printf("%s: %s\n", name, status)
	}
	fmt.Println("----------------")
}
