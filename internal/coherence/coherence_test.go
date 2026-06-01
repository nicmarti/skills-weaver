package coherence

import "testing"

func TestScoreFor(t *testing.T) {
	cases := []struct {
		name     string
		findings []Finding
		want     int
	}{
		{"empty", nil, 100},
		{"one info", []Finding{{Severity: SeverityInfo}}, 99},
		{"one warning", []Finding{{Severity: SeverityWarning}}, 95},
		{"one error", []Finding{{Severity: SeverityError}}, 85},
		{"mixed", []Finding{{Severity: SeverityError}, {Severity: SeverityWarning}, {Severity: SeverityInfo}}, 79},
		{"floored at zero", []Finding{
			{Severity: SeverityError}, {Severity: SeverityError}, {Severity: SeverityError},
			{Severity: SeverityError}, {Severity: SeverityError}, {Severity: SeverityError},
			{Severity: SeverityError},
		}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := scoreFor(c.findings); got != c.want {
				t.Errorf("scoreFor() = %d, want %d", got, c.want)
			}
		})
	}
}

func TestLayerReportCounts(t *testing.T) {
	lr := newLayerReport(LayerIntegrity, []Finding{
		{Severity: SeverityError},
		{Severity: SeverityWarning},
		{Severity: SeverityWarning},
		{Severity: SeverityInfo},
	})
	e, w, i := lr.Counts()
	if e != 1 || w != 2 || i != 1 {
		t.Errorf("Counts() = (%d,%d,%d), want (1,2,1)", e, w, i)
	}
	if !lr.HasErrors() {
		t.Error("HasErrors() = false, want true")
	}
}

func TestNewLayerReportNeverNilFindings(t *testing.T) {
	lr := newLayerReport(LayerIntegrity, nil)
	if lr.Findings == nil {
		t.Error("Findings should be a non-nil empty slice for clean JSON output")
	}
	if lr.Score != 100 {
		t.Errorf("empty layer score = %d, want 100", lr.Score)
	}
}
