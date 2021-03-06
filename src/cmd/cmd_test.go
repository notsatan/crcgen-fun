package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/notsatan/crcgen/src/logger"
)

const (
	envProd  = "production_mode"
	envDebug = "debug_mode"
	envBuild = "__BUILD_MODE__"
)

func resetEnv() {
	// Unset all environment variables
	_ = os.Unsetenv(envProd)
	_ = os.Unsetenv(envDebug)
	_ = os.Unsetenv(envBuild)
}

func reset() {
	exit = os.Exit
	initLogger = logger.Log
	execCmd = Root.Execute
	cmdUsage = (*cobra.Command).Usage
	closeLogger = logger.Stop
}

func TestMain(m *testing.M) {
	// Run all tests, unset env variables, and exit
	resetEnv()

	val := m.Run()

	resetEnv()
	os.Exit(val)
}

func TestRun_DebugMode(t *testing.T) {
	// Mock debug mode through environment variable - ensure logger gets enabled

	resetEnv()
	reset()

	t.Setenv(debugMode, envDebug)

	calls := 0
	execCmd = func() error { return nil }
	initLogger = func(file bool) (err error) {
		if file {
			calls++
		}

		calls++
		return nil
	}

	assert.NoError(t, Run())

	// `calls` should be exactly 1 - less implies logger was never run, more implies
	// logs were attempted to be written to the file as well (should not happen)
	assert.Equal(t, 1, calls)
}

func TestRun_LoggerFail(t *testing.T) {
	// Error should return if the logger fails to open

	reset()
	resetEnv()

	t.Setenv(debugMode, envDebug)
	initLogger = func(bool) (err error) {
		return fmt.Errorf("(%s/TestRun_LoggerFail): test error", pkgName)
	}

	assert.Error(t, Run())
}

func TestRun_ProdMode(t *testing.T) {
	reset()
	resetEnv()

	calls := 0
	execCmd = func() error { return nil }
	initLogger = func(bool) (err error) {
		calls++
		return nil
	}

	assert.NoError(t, Run())
	assert.Equal(t, calls, 0, "logging enabled in production mode")
}

func TestRun_CmdFail(t *testing.T) {
	// Test to ensure an error is returned if `cmd.Root` fails

	reset()
	resetEnv()

	execCmd = func() error {
		return fmt.Errorf("(%s/TestRun_CmdFail): test error", pkgName)
	}

	assert.Error(t, Run())
}

func TestCloseRes(t *testing.T) {
	reset()

	// `closeRes` is designed to run as an independent function - it should consume
	// errors internally

	closeLogger = func() error { return nil }
	assert.NotPanics(t, func() { closeRes() })

	closeLogger = func() error { return fmt.Errorf("(%s/Run): test error", pkgName) }
	assert.NotPanics(t, func() { closeRes() })
}

func TestRootCmd(t *testing.T) {
	reset()
	resetEnv()

	assert.NoError(t, Root.Help())
	assert.NoError(t, Root.Usage())
	assert.NoError(t, Root.Execute())

	// Ensure os.Exit is called if command usage can't be printed
	calls := 0
	exit = func(int) { calls++ }
	cmdUsage = func(*cobra.Command) error {
		return fmt.Errorf("(%s/TestRootCmd): test", pkgName)
	}

	// Run the command - will attempt to print `usage`, which fails, and should result
	// in a call to `exit`
	assert.NoError(t, Root.Execute())
	assert.Equal(t, 1, calls)
}
