package models_test

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	green = color.New(color.FgGreen).SprintFunc()
	red   = color.New(color.FgRed).SprintFunc()
	blue  = color.New(color.FgCyan).SprintFunc()
)

func colorizeTestResult(name string, passed bool) string {
	if passed {
		return fmt.Sprintf("%s: %s", blue("=== RUN"), green(name))
	}
	return fmt.Sprintf("%s: %s", blue("=== RUN"), red(name))
}

func colorizeTestPass(name string, duration string) string {
	return fmt.Sprintf("%s: %s %s", green("--- PASS"), green(name), blue(duration))
}

func colorizeTestFail(name string, duration string) string {
	return fmt.Sprintf("%s: %s %s", red("--- FAIL"), red(name), blue(duration))
}
