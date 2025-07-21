package actionskit

import (
	"fmt"
	"os"
	"strings"
)

// GetInput retrieves the value of an input by name from the environment.
// GitHub Actions sets input values as environment variables with the prefix INPUT_
// and converts the input name to uppercase with dashes replaced by underscores.
//
// For example, an input named "github-token" would be available as environment
// variable "INPUT_GITHUB_TOKEN".
//
// Returns the input value or an empty string if the input is not set.
func GetInput(name string) string {
	// Convert input name to environment variable format
	envName := "INPUT_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
	return os.Getenv(envName)
}

// GetInputRequired retrieves the value of a required input by name from the environment.
// If the input is not set or is empty, it returns an error.
func GetInputRequired(name string) (string, error) {
	value := GetInput(name)
	if value == "" {
		return "", fmt.Errorf("input required and not supplied: %s", name)
	}
	return value, nil
}

// SetOutput sets an output parameter for the action.
// GitHub Actions reads outputs from the GITHUB_OUTPUT environment file.
func SetOutput(name, value string) error {
	outputFile := os.Getenv("GITHUB_OUTPUT")
	if outputFile == "" {
		// For local testing, output to stdout
		fmt.Printf("::set-output name=%s::%s\n", name, value)
		return nil
	}

	// Write to GitHub Actions output file
	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening output file: %v", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s=%s\n", name, value)
	if err != nil {
		return fmt.Errorf("error writing to output file: %v", err)
	}

	return nil
}

// IsDebug returns true if the runner debug mode is enabled.
func IsDebug() bool {
	return os.Getenv("RUNNER_DEBUG") == "1"
}

// Info writes an info message to the log.
func Info(message string) {
	fmt.Println(message)
}

// Debug writes a debug message to the log if debug mode is enabled.
func Debug(message string) {
	if IsDebug() {
		fmt.Printf("::debug::%s\n", message)
	}
}

// Warning writes a warning message to the log.
func Warning(message string) {
	fmt.Printf("::warning::%s\n", message)
}

// Error writes an error message to the log.
func Error(message string) {
	fmt.Fprintf(os.Stderr, "::error::%s\n", message)
}