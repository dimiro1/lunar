package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/dimiro1/lunar/lunar-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	flagServer string
	flagToken  string

	// resolved after PersistentPreRunE
	serverURL string
	apiToken  string
)

// envConfig holds CLI settings sourced from environment variables. They override
// the config file but are themselves overridden by command-line flags.
type envConfig struct {
	Server string `env:"LUNAR_SERVER"`
	Token  string `env:"LUNAR_TOKEN"`
}

var rootCmd = &cobra.Command{
	Use:   "lunar-cli",
	Short: "Lunar CLI – manage your Lunar FaaS functions",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		serverURL = cfg.Server
		apiToken = cfg.Token

		// Environment variables override the config file.
		envCfg, _ := env.ParseAs[envConfig]()
		if envCfg.Server != "" {
			serverURL = envCfg.Server
		}
		if envCfg.Token != "" {
			apiToken = envCfg.Token
		}

		// Flags override everything.
		if cmd.Root().PersistentFlags().Changed("server") {
			serverURL = flagServer
		}
		if cmd.Root().PersistentFlags().Changed("token") {
			apiToken = flagToken
		}
		return nil
	},
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Run executes the CLI with the given arguments, writing output to w.
// Returns the error from the command, if any. Intended for testing.
func Run(args []string, w io.Writer) error {
	prev := outputWriter
	outputWriter = w
	defer func() { outputWriter = prev }()

	prevOut := rootCmd.OutOrStdout()
	prevErr := rootCmd.ErrOrStderr()
	rootCmd.SetOut(w)
	rootCmd.SetErr(w)
	defer func() {
		rootCmd.SetOut(prevOut)
		rootCmd.SetErr(prevErr)
	}()

	resetAllFlags(rootCmd)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

// resetAllFlags resets every flag in cmd and its subcommands to its default
// value, so that successive Run calls do not leak flag state.
func resetAllFlags(c *cobra.Command) {
	c.Flags().VisitAll(resetFlag)
	c.PersistentFlags().VisitAll(resetFlag)
	for _, sub := range c.Commands() {
		resetAllFlags(sub)
	}
}

func resetFlag(f *pflag.Flag) {
	f.Changed = false
	// StringArray/StringSlice flags use append semantics on Set(), so the
	// generic Set(DefValue) path would push "[]" onto the slice instead of
	// clearing it. Use the SliceValue interface to replace with an empty slice.
	if sv, ok := f.Value.(pflag.SliceValue); ok {
		_ = sv.Replace([]string{})
		return
	}
	_ = f.Value.Set(f.DefValue)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagServer, "server", "", "Lunar server URL (env: LUNAR_SERVER)")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "API token (env: LUNAR_TOKEN)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format: pretty (default) or json")
	rootCmd.PersistentFlags().BoolVar(&showCode, "show-code", false, "Print code fields when displaying functions or versions")
}

// printJSON renders a JSON response body using the active output format.
// When --output is not set it auto-detects: pretty on a terminal, json when piped.
func printJSON(data []byte) error {
	format := outputFormat
	if format == "" {
		if isTerminal() {
			format = "pretty"
		} else {
			format = "json"
		}
	}
	if format == "json" {
		return printRawJSON(data)
	}
	return printPretty(data)
}

func printAPIResponse(status int, body []byte) error {
	if status < 200 || status >= 300 {
		return apiResponseError(status, body)
	}
	return printJSON(body)
}

func apiResponseError(status int, body []byte) error {
	trimmed := strings.TrimSpace(string(bytes.TrimSpace(body)))
	if trimmed == "" {
		return fmt.Errorf("request failed with HTTP %d", status)
	}
	return fmt.Errorf("request failed with HTTP %d: %s", status, trimmed)
}
