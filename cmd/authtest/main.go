// cmd/authtest is a throwaway prototype that proves the Kite Connect auth flow
// end-to-end and inspects the raw historical data response format.
//
// This binary is NOT production code. It is deleted or left unused once the
// Zerodha provider (TASK-0008) is implemented.
//
// Usage:
//
//	cp .env.example .env          # first time only
//	# fill in KITE_API_KEY and KITE_API_SECRET in .env
//	go run ./cmd/authtest
//
// Credentials are loaded from .env at the project root. Real environment
// variables take precedence over .env values.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	kiteBaseURL    = "https://kite.zerodha.com"
	kiteAPIBaseURL = "https://api.kite.trade"
	kiteVersion    = "3"

	// NIFTY 50 index instrument token (NSE). Used for the candle smoke-test.
	// Verify this is correct for your account from the instruments CSV.
	nifty50Token = "256265"
)

func main() {
	loadDotEnv(".env")

	apiKey := mustEnv("KITE_API_KEY")
	apiSecret := mustEnv("KITE_API_SECRET")

	// ── Step 1: Login URL ─────────────────────────────────────────────────────

	loginURL := fmt.Sprintf("%s/connect/login?v=%s&api_key=%s", kiteBaseURL, kiteVersion, apiKey)

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("STEP 1 — Open this URL in your browser and log in:")
	fmt.Println()
	fmt.Println(" ", loginURL)
	fmt.Println()
	fmt.Println("After login, Kite will redirect to your registered redirect_url.")
	fmt.Println("The browser will show a 'connection refused' error — that is expected.")
	fmt.Println("Copy the value of the 'request_token' query parameter from the address bar.")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// ── Step 2: Read request_token ────────────────────────────────────────────

	requestToken := prompt("Paste request_token here: ")
	if requestToken == "" {
		fatal("request_token cannot be empty")
	}

	// ── Step 3: Compute checksum ──────────────────────────────────────────────

	h := sha256.New()
	h.Write([]byte(apiKey + requestToken + apiSecret))
	checksum := fmt.Sprintf("%x", h.Sum(nil))

	fmt.Printf("\nChecksum (SHA-256): %s\n", checksum)

	// ── Step 4: Exchange request_token for access_token ───────────────────────

	fmt.Println("\nSTEP 2 — Exchanging request_token for access_token...")

	sessionResp, err := postSession(apiKey, requestToken, checksum)
	if err != nil {
		fatal("session exchange failed: %v", err)
	}

	fmt.Printf("\n✓ access_token: %s\n", sessionResp.AccessToken)
	fmt.Printf("  user_id:       %s\n", sessionResp.UserID)
	fmt.Printf("  login_time:    %s\n", sessionResp.LoginTime)
	fmt.Printf("  expires at:    6:00 AM IST tomorrow (00:30 UTC)\n")

	// ── Step 5: Smoke-test — fetch last 5 daily candles for NIFTY 50 ─────────

	fmt.Println("\nSTEP 3 — Fetching last 5 daily candles for NIFTY 50 (token 256265)...")

	to := time.Now()
	from := to.AddDate(0, 0, -10) // 10 calendar days back; expect ~5–7 trading days

	candles, rawJSON, err := fetchCandles(apiKey, sessionResp.AccessToken, nifty50Token, "day", from, to)
	if err != nil {
		fatal("historical fetch failed: %v", err)
	}

	fmt.Printf("\n✓ Raw JSON response:\n%s\n", rawJSON)

	fmt.Printf("\n✓ Parsed candles (%d):\n", len(candles))
	fmt.Printf("  %-28s  %8s  %8s  %8s  %8s  %10s\n", "Timestamp", "Open", "High", "Low", "Close", "Volume")
	fmt.Printf("  %s\n", strings.Repeat("─", 80))
	for _, c := range candles {
		fmt.Printf("  %-28s  %8.2f  %8.2f  %8.2f  %8.2f  %10.0f\n",
			c.Timestamp, c.Open, c.High, c.Low, c.Close, c.Volume)
	}

	// ── Step 6: Instrument token lookup smoke-test ────────────────────────────

	fmt.Println("\nSTEP 4 — Fetching instruments CSV (first 5 rows)...")
	if err := printInstrumentsSample(apiKey, sessionResp.AccessToken); err != nil {
		fmt.Printf("  ⚠  instruments fetch failed: %v\n", err)
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Auth prototype complete. Record findings in TASK-0007 decision doc.")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// ── Session exchange ──────────────────────────────────────────────────────────

type sessionResponse struct {
	AccessToken string `json:"access_token"`
	UserID      string `json:"user_id"`
	LoginTime   string `json:"login_time"`
}

func postSession(apiKey, requestToken, checksum string) (sessionResponse, error) {
	form := url.Values{}
	form.Set("api_key", apiKey)
	form.Set("request_token", requestToken)
	form.Set("checksum", checksum)

	req, err := http.NewRequest(http.MethodPost,
		kiteAPIBaseURL+"/session/token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return sessionResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Kite-Version", kiteVersion)

	body, err := doRequest(req)
	if err != nil {
		return sessionResponse{}, err
	}

	var envelope struct {
		Status string          `json:"status"`
		Data   sessionResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return sessionResponse{}, fmt.Errorf("unmarshal session response: %w\nraw: %s", err, body)
	}
	return envelope.Data, nil
}

// ── Historical candle fetch ───────────────────────────────────────────────────

type candle struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

func fetchCandles(apiKey, accessToken, token, interval string, from, to time.Time) ([]candle, string, error) {
	endpoint := fmt.Sprintf("%s/instruments/historical/%s/%s", kiteAPIBaseURL, token, interval)

	req, err := http.NewRequest(http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, "", err
	}

	q := req.URL.Query()
	q.Set("from", from.Format("2006-01-02 15:04:05"))
	q.Set("to", to.Format("2006-01-02 15:04:05"))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", fmt.Sprintf("token %s:%s", apiKey, accessToken))
	req.Header.Set("X-Kite-Version", kiteVersion)

	body, err := doRequest(req)
	if err != nil {
		return nil, "", err
	}

	// Pretty-print the raw JSON for inspection.
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, body, "  ", "  "); err != nil {
		pretty.Write(body) // fallback to raw if indent fails
	}

	// Parse the array-of-arrays response:
	// { "data": { "candles": [[ts, o, h, l, c, v], ...] } }
	var envelope struct {
		Data struct {
			Candles [][]interface{} `json:"candles"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, pretty.String(), fmt.Errorf("unmarshal candles: %w", err)
	}

	ist, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		// Fall back to fixed +05:30 offset if timezone data is unavailable.
		ist = time.FixedZone("IST", 5*3600+30*60)
	}

	candles := make([]candle, 0, len(envelope.Data.Candles))
	for _, row := range envelope.Data.Candles {
		if len(row) < 6 {
			continue
		}
		tsStr, ok := row[0].(string)
		if !ok {
			continue
		}
		// Kite returns "2024-01-01T09:15:00+0530" — parse with IST location.
		ts, err := time.ParseInLocation("2006-01-02T15:04:05-0700", tsStr, ist)
		if err != nil {
			// Try alternate format without explicit offset.
			ts, err = time.Parse(time.RFC3339, tsStr)
			if err != nil {
				fmt.Printf("  ⚠  could not parse timestamp %q: %v\n", tsStr, err)
				continue
			}
		}
		candles = append(candles, candle{
			Timestamp: ts.UTC(), // store as UTC
			Open:      toFloat(row[1]),
			High:      toFloat(row[2]),
			Low:       toFloat(row[3]),
			Close:     toFloat(row[4]),
			Volume:    toFloat(row[5]),
		})
	}

	return candles, pretty.String(), nil
}

// ── Instruments CSV sample ────────────────────────────────────────────────────

func printInstrumentsSample(apiKey, accessToken string) error {
	req, err := http.NewRequest(http.MethodGet, kiteAPIBaseURL+"/instruments", http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s:%s", apiKey, accessToken))
	req.Header.Set("X-Kite-Version", kiteVersion)

	body, err := doRequest(req)
	if err != nil {
		return err
	}

	// Print CSV header + first 5 rows.
	lines := strings.Split(string(body), "\n")
	limit := 6 // header + 5 rows
	if len(lines) < limit {
		limit = len(lines)
	}
	for _, l := range lines[:limit] {
		fmt.Printf("  %s\n", l)
	}
	fmt.Printf("  ... (%d total rows)\n", len(lines)-1)
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func doRequest(req *http.Request) ([]byte, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}
	return body, nil
}

func toFloat(v interface{}) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

func prompt(label string) string {
	fmt.Print(label)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fatal("environment variable %s is not set", key)
	}
	return v
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

// loadDotEnv reads a .env file and sets any key=value pair as an environment
// variable, skipping lines that are blank or start with #. Real environment
// variables already set take precedence and are never overwritten.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // .env is optional; silently skip if missing
	}
	defer f.Close() //nolint:errcheck // read-only file; close error is non-fatal

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if os.Getenv(key) == "" { // real env vars take precedence
			os.Setenv(key, value) //nolint:errcheck // best-effort; failure is non-fatal
		}
	}
}
