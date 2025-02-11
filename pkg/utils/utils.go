package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// FormatOutput formats the output for display.
func FormatOutput(data interface{}) error {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

// ParseInput reads input from the command line and returns it as a string.
func ParseInput() (string, error) {
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		return "", err
	}
	return input, nil
}