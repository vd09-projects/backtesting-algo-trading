package output

// LoadCurveCSV is in the output package alongside writeCurveCSV — symmetric read/write
// boundary for the timestamp,equity_value format.
//
// **Decision (2026-04.1.0) — architecture: experimental**
// scope: internal/output, correlation
// tags: LoadCurveCSV, csv-reader, TASK-0027
//
// CSV reader lives in internal/output alongside writeCurveCSV. Alternative was a
// separate internal/curveio package — overkill for one function at this stage.
// The read/write pair is collocated so schema changes touch one file.

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// LoadCurveCSV reads an equity curve CSV file written by writeCurveCSV and returns
// the parsed equity points in file order. The expected format is:
//
//	timestamp,equity_value
//	2018-01-02T09:15:00Z,100000.00
//
// Timestamps must be RFC 3339 UTC. Returns an error for any malformed row.
// An empty file (header only) returns a nil slice and a nil error.
func LoadCurveCSV(path string) ([]model.EquityPoint, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("output: load curve CSV %q: %w", path, err)
	}
	defer f.Close() //nolint:errcheck // read-only open; Close error is not actionable

	var pts []model.EquityPoint
	scanner := bufio.NewScanner(f)

	// Skip the header row.
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("output: load curve CSV %q: read header: %w", path, err)
		}
		return nil, nil // empty file
	}

	lineNum := 1
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, ",", 2)
		if len(fields) != 2 {
			return nil, fmt.Errorf("output: load curve CSV %q: line %d: expected 2 fields, got %d", path, lineNum, len(fields))
		}

		ts, err := time.Parse(time.RFC3339, strings.TrimSpace(fields[0]))
		if err != nil {
			return nil, fmt.Errorf("output: load curve CSV %q: line %d: parse timestamp: %w", path, lineNum, err)
		}

		val, err := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("output: load curve CSV %q: line %d: parse value: %w", path, lineNum, err)
		}

		pts = append(pts, model.EquityPoint{Timestamp: ts.UTC(), Value: val})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("output: load curve CSV %q: scan: %w", path, err)
	}

	return pts, nil
}
