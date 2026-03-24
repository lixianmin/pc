package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand_Help(t *testing.T) {
	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := buf.String()

	// Check for expected content
	if !strings.Contains(output, "PersonalClaw") {
		t.Error("Help output should contain 'PersonalClaw'")
	}

	if !strings.Contains(output, "Usage:") {
		t.Error("Help output should contain 'Usage:'")
	}
}

func TestVersionCommand(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := strings.TrimSpace(buf.String())
	expected := "PersonalClaw version dev"
	if output != expected {
		t.Errorf("version output = %q, want %q", output, expected)
	}
}

func TestGatewayCommand_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"gateway", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "gateway") {
		t.Error("Help output should contain 'gateway'")
	}

	if !strings.Contains(output, "start") {
		t.Error("Help output should contain 'start'")
	}

	if !strings.Contains(output, "stop") {
		t.Error("Help output should contain 'stop'")
	}

	if !strings.Contains(output, "status") {
		t.Error("Help output should contain 'status'")
	}
}

func TestGatewayStartCommand_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"gateway", "start", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "start") {
		t.Error("Help output should contain 'start'")
	}
}

func TestGatewayStopCommand_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"gateway", "stop", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "stop") {
		t.Error("Help output should contain 'stop'")
	}
}

func TestGatewayStatusCommand_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"gateway", "status", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "status") {
		t.Error("Help output should contain 'status'")
	}
}
