package sweep2d

import (
	"bytes"
	"fmt"
	"os"
)

// WriteCSV writes the Sharpe ratio matrix from report to a CSV file at path.
// The file is created or truncated. Rows index Param1Values; columns index Param2Values.
//
// Format (readable by pandas with index_col=0 for heatmap consumption):
//
//	param1\param2,v2_0,v2_1,...
//	v1_0,sharpe_00,sharpe_01,...
//	v1_1,sharpe_10,sharpe_11,...
func WriteCSV(path string, report Report2D) error { //nolint:gocritic // Report2D is a result value; callers hold it by value and passing by pointer adds noise at call sites
	var buf bytes.Buffer

	// Header: corner label then param2 values.
	corner := report.Param1Name + `\` + report.Param2Name
	buf.WriteString(corner)
	for _, v2 := range report.Param2Values {
		fmt.Fprintf(&buf, ",%g", v2)
	}
	buf.WriteByte('\n')

	// One row per param1 value.
	for i, v1 := range report.Param1Values {
		fmt.Fprintf(&buf, "%g", v1)
		for j := range report.Param2Values {
			fmt.Fprintf(&buf, ",%.6f", report.Grid[i][j].SharpeRatio)
		}
		buf.WriteByte('\n')
	}

	// Metadata footer as comments so pandas ignores them.
	fmt.Fprintf(&buf, "# variants=%d  peak_sharpe=%.6f  dsr_corrected=%.6f\n",
		report.VariantCount, report.PeakSharpe, report.DSRCorrectedPeakSharpe)

	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("sweep2d: write CSV %q: %w", path, err)
	}
	return nil
}
