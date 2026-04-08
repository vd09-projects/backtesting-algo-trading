package zerodha

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestParseInstrumentsCSV_basic(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "instruments.csv"))
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	tokens, err := parseInstrumentsCSV(data)
	if err != nil {
		t.Fatalf("parseInstrumentsCSV: %v", err)
	}

	want := map[string]int64{
		"NSE:NIFTY 50":  256265,
		"BSE:SENSEX":    260105,
		"NSE:RELIANCE":  738561,
		"NSE:INFY":      5633,
		"MCX:GOLDPETAL": 225537,
	}

	for key, wantToken := range want {
		if got := tokens[key]; got != wantToken {
			t.Errorf("tokens[%q] = %d, want %d", key, got, wantToken)
		}
	}
	if len(tokens) != len(want) {
		t.Errorf("token map has %d entries, want %d", len(tokens), len(want))
	}
}

func TestParseInstrumentsCSV_empty(t *testing.T) {
	header := []byte("instrument_token,exchange_token,tradingsymbol,name,last_price,expiry,strike,tick_size,lot_size,instrument_type,segment,exchange\n")
	tokens, err := parseInstrumentsCSV(header)
	if err != nil {
		t.Fatalf("parseInstrumentsCSV on header-only CSV: %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("want empty map, got %d entries", len(tokens))
	}
}

func TestParseInstrumentsCSV_missing_columns(t *testing.T) {
	bad := []byte("foo,bar,baz\n1,2,3\n")
	_, err := parseInstrumentsCSV(bad)
	if err == nil {
		t.Fatal("want error for missing required columns, got nil")
	}
}

func TestParseInstrumentsCSV_skips_bad_token(t *testing.T) {
	csv := []byte("instrument_token,exchange_token,tradingsymbol,name,last_price,expiry,strike,tick_size,lot_size,instrument_type,segment,exchange\n" +
		"notanumber,1,RELIANCE,Reliance,0,,0,0.05,1,EQ,NSE,NSE\n" +
		"738561,2885,RELIANCE,Reliance,0,,0,0.05,1,EQ,NSE,NSE\n")
	tokens, err := parseInstrumentsCSV(csv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 1 {
		t.Errorf("want 1 entry (bad row skipped), got %d", len(tokens))
	}
	if tokens["NSE:RELIANCE"] != 738561 {
		t.Errorf("tokens[NSE:RELIANCE] = %d, want 738561", tokens["NSE:RELIANCE"])
	}
}

func TestLoadInstrumentsCSV_success(t *testing.T) {
	csvData, err := os.ReadFile(filepath.Join("testdata", "instruments.csv"))
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/instruments" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write(csvData)
	}))
	defer srv.Close()

	tokens, err := loadInstrumentsCSV(t.Context(), srv.Client(), srv.URL, "key", "token")
	if err != nil {
		t.Fatalf("loadInstrumentsCSV: %v", err)
	}
	if tokens["NSE:NIFTY 50"] != 256265 {
		t.Errorf("tokens[NSE:NIFTY 50] = %d, want 256265", tokens["NSE:NIFTY 50"])
	}
}

func TestLoadInstrumentsCSV_auth_error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := loadInstrumentsCSV(t.Context(), srv.Client(), srv.URL, "key", "badtoken")
	if !errors.Is(err, ErrAuthRequired) {
		t.Errorf("want ErrAuthRequired, got %v", err)
	}
}
