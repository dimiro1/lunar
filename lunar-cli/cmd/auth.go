package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"

	"github.com/dimiro1/lunar/lunar-cli/config"
	"github.com/spf13/cobra"
)

// Device authorization flow stays REST: it runs before the CLI is
// authenticated and maps poorly to GraphQL (it sets up the very token a
// GraphQL request would need). These two endpoints live under /api/auth/*.

// deviceRequestResponse mirrors the server's POST /api/auth/device-request body.
type deviceRequestResponse struct {
	DeviceCode  string `json:"device_code"`
	UserCode    string `json:"user_code"`
	ApprovalURL string `json:"approval_url"`
	ExpiresIn   int    `json:"expires_in"`
	Interval    int    `json:"interval"`
}

// deviceTokenResponse mirrors the server's GET /api/auth/device-token body.
type deviceTokenResponse struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate via device authorization flow",
	RunE:  runLogin,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove the stored authentication token",
	RunE:  runLogout,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	req, err := requestDeviceCode(cmd.Context())
	if err != nil {
		return fmt.Errorf("device request: %w", err)
	}

	fmt.Fprintf(out, "Your verification code: %s\n", req.UserCode)
	fmt.Fprintf(out, "Open this URL to approve: %s\n", req.ApprovalURL)
	openBrowser(req.ApprovalURL)
	fmt.Fprint(out, "Waiting for approval")

	interval := time.Duration(req.Interval) * time.Second
	expires := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)

	for time.Now().Before(expires) {
		time.Sleep(interval)

		tokenResp, err := pollDeviceToken(cmd.Context(), req.DeviceCode)
		if err != nil {
			fmt.Fprintln(out)
			return fmt.Errorf("polling: %w", err)
		}

		switch tokenResp.Status {
		case "approved":
			if tokenResp.Token == "" {
				fmt.Fprintln(out)
				return fmt.Errorf("approved but no token returned")
			}
			cfg, _ := config.Load()
			cfg.Token = tokenResp.Token
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("saving token: %w", err)
			}
			fmt.Fprintln(out, "\nAuthentication successful. Token saved.")
			return nil
		case "denied":
			fmt.Fprintln(out)
			return fmt.Errorf("authorization denied")
		default:
			fmt.Fprint(out, ".")
		}
	}
	fmt.Fprintln(out)
	return fmt.Errorf("authorization timed out")
}

func runLogout(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.Token = ""
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Logged out.")
	return nil
}

// requestDeviceCode starts a device authorization flow via
// POST /api/auth/device-request.
func requestDeviceCode(ctx context.Context) (*deviceRequestResponse, error) {
	if serverURL == "" {
		return nil, fmt.Errorf("no server configured (use --server or LUNAR_SERVER)")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL+"/api/auth/device-request", nil)
	if err != nil {
		return nil, err
	}
	var resp deviceRequestResponse
	if err := doJSON(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// pollDeviceToken checks the status of a pending device authorization via
// GET /api/auth/device-token?code=<device_code>.
func pollDeviceToken(ctx context.Context, deviceCode string) (*deviceTokenResponse, error) {
	endpoint := serverURL + "/api/auth/device-token?code=" + url.QueryEscape(deviceCode)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	var resp deviceTokenResponse
	if err := doJSON(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// doJSON sends req and decodes a 2xx JSON body into v, turning non-2xx
// responses into an error carrying the server's message.
func doJSON(req *http.Request, v any) error {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return apiResponseError(res.StatusCode, body)
	}
	if err := json.Unmarshal(bytes.TrimSpace(body), v); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

var openBrowser = func(url string) {
	var name string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		name, args = "open", []string{url}
	case "windows":
		name, args = "cmd", []string{"/c", "start", url}
	default:
		name, args = "xdg-open", []string{url}
	}
	_ = exec.Command(name, args...).Start()
}
