package main

import (
	"bytes"
	"testing"

	"github.com/ollama/ollama/cmd"
	"github.com/spf13/cobra"
)

func TestCLI(t *testing.T) {
	cli := cmd.NewCLI()
	b := bytes.NewBufferString("")
	cli.SetOut(b)
	cli.SetArgs([]string{"--version"})
	err := cli.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	out := b.String()
	if out == "" {
		t.Fatalf("Expected version output, got empty string")
	}
}

func TestCommands(t *testing.T) {
	cli := cmd.NewCLI()
	commands := []string{"create", "show", "run", "stop", "serve", "pull", "push", "list", "ps", "cp", "rm"}

	for _, cmd := range commands {
		if findSubCommand(cli, cmd) == nil {
			t.Errorf("Command '%s' not found", cmd)
		}
	}
}

func findSubCommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, c := range cmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}
