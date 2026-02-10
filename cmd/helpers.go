package cmd

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// readPassword reads a password from the terminal without echoing.
func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)

	fd := int(os.Stdin.Fd())

	password, err := term.ReadPassword(fd)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}

	_, _ = fmt.Fprintln(os.Stderr)

	return string(password), nil
}
