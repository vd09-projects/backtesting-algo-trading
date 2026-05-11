package main

import (
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// ---------------------------------------------------------------------------
// TestParseAndValidateFlags
// ---------------------------------------------------------------------------

func TestParseAndValidateFlags_ValidInput(t *testing.T) {
	t.Parallel()
	f := &flags{
		fromStr: "2018-01-01",
		toStr:   "2024-12-31",
		tfStr:   "daily",
	}
	from, to, tf, err := parseAndValidateFlags(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !from.Equal(time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("from: got %s", from)
	}
	if !to.Equal(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("to: got %s", to)
	}
	if tf != model.TimeframeDaily {
		t.Errorf("tf: got %s, want daily", tf)
	}
}

func TestParseAndValidateFlags_MissingFrom(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags(&flags{fromStr: "", toStr: "2024-12-31", tfStr: "daily"})
	if err == nil {
		t.Fatal("expected error for missing --from")
	}
	if !strings.Contains(err.Error(), "--from") {
		t.Errorf("error should mention --from, got: %v", err)
	}
}

func TestParseAndValidateFlags_MissingTo(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags(&flags{fromStr: "2018-01-01", toStr: "", tfStr: "daily"})
	if err == nil {
		t.Fatal("expected error for missing --to")
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("error should mention --to, got: %v", err)
	}
}

func TestParseAndValidateFlags_InvalidFromDate(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags(&flags{fromStr: "not-a-date", toStr: "2024-12-31", tfStr: "daily"})
	if err == nil {
		t.Fatal("expected error for invalid --from date")
	}
	if !strings.Contains(err.Error(), "--from") {
		t.Errorf("error should mention --from, got: %v", err)
	}
}

func TestParseAndValidateFlags_InvalidToDate(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags(&flags{fromStr: "2018-01-01", toStr: "bad", tfStr: "daily"})
	if err == nil {
		t.Fatal("expected error for invalid --to date")
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("error should mention --to, got: %v", err)
	}
}

func TestParseAndValidateFlags_ToNotAfterFrom(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags(&flags{fromStr: "2024-01-01", toStr: "2018-01-01", tfStr: "daily"})
	if err == nil {
		t.Fatal("expected error when --to is not after --from")
	}
	if !strings.Contains(err.Error(), "--to must be strictly after") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParseAndValidateFlags_ToEqualFrom(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags(&flags{fromStr: "2024-01-01", toStr: "2024-01-01", tfStr: "daily"})
	if err == nil {
		t.Fatal("expected error when --to equals --from")
	}
}

func TestParseAndValidateFlags_InvalidTimeframe(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags(&flags{fromStr: "2018-01-01", toStr: "2024-12-31", tfStr: "monthly"})
	if err == nil {
		t.Fatal("expected error for invalid --timeframe")
	}
	if !strings.Contains(err.Error(), "--timeframe") {
		t.Errorf("error should mention --timeframe, got: %v", err)
	}
}

func TestParseAndValidateFlags_AllValidTimeframes(t *testing.T) {
	t.Parallel()
	tfs := []string{"1min", "5min", "15min", "daily", "weekly"}
	for _, tfStr := range tfs {
		t.Run(tfStr, func(t *testing.T) {
			t.Parallel()
			_, _, tf, err := parseAndValidateFlags(&flags{fromStr: "2018-01-01", toStr: "2024-12-31", tfStr: tfStr})
			if err != nil {
				t.Fatalf("unexpected error for timeframe %q: %v", tfStr, err)
			}
			if string(tf) != tfStr {
				t.Errorf("tf: got %s, want %s", tf, tfStr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestParseSizingConfig
// ---------------------------------------------------------------------------

func TestParseSizingConfig_Fixed(t *testing.T) {
	t.Parallel()
	sm, err := parseSizingConfig("fixed", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm != model.SizingFixed {
		t.Errorf("got %v, want SizingFixed", sm)
	}
}

func TestParseSizingConfig_VolTargetValid(t *testing.T) {
	t.Parallel()
	sm, err := parseSizingConfig("vol-target", 0.10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm != model.SizingVolatilityTarget {
		t.Errorf("got %v, want SizingVolatilityTarget", sm)
	}
}

func TestParseSizingConfig_VolTargetZeroVolTarget(t *testing.T) {
	t.Parallel()
	_, err := parseSizingConfig("vol-target", 0)
	if err == nil {
		t.Fatal("expected error for vol-target with vol-target=0")
	}
	if !strings.Contains(err.Error(), "--vol-target") {
		t.Errorf("error should mention --vol-target, got: %v", err)
	}
}

func TestParseSizingConfig_UnknownModel(t *testing.T) {
	t.Parallel()
	_, err := parseSizingConfig("unknown", 0)
	if err == nil {
		t.Fatal("expected error for unknown sizing model")
	}
	if !strings.Contains(err.Error(), "--sizing-model") {
		t.Errorf("error should mention --sizing-model, got: %v", err)
	}
}
