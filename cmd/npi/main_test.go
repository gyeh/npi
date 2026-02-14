package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"nppes_golang/npi"
)

// TestMain builds the CLI binary once before all tests in this file.
var cliBinary string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "npcli-test-*")
	if err != nil {
		panic(err)
	}
	cliBinary = filepath.Join(tmp, "npcli")

	build := exec.Command("go", "build", "-o", cliBinary, ".")
	build.Dir = filepath.Join(".")
	build.Env = append(os.Environ(), "GOWORK=off")
	out, err := build.CombinedOutput()
	if err != nil {
		panic("failed to build CLI: " + err.Error() + "\n" + string(out))
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

// makeTestDataDir creates a temp directory with valid NPPES CSV files for CLI tests.
func makeTestDataDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Main provider CSV
	cols := npi.MainSchemaColumnNames()
	header := strings.Join(cols, ",")

	row1 := make([]string, len(cols))
	row1[0] = "1234567890"  // NPI
	row1[1] = "1"           // Entity Type (Individual)
	row1[5] = "DOE"         // Last name (schema index 5)
	row1[6] = "JOHN"        // First name (schema index 6)
	row1[23] = "CA"         // Mailing state
	row1[24] = "90210"      // Mailing postal code
	row1[47] = "207Q00000X" // Taxonomy code 1

	row2 := make([]string, len(cols))
	row2[0] = "2345678901"
	row2[1] = "1"
	row2[5] = "SMITH"
	row2[6] = "JANE"
	row2[23] = "NY"
	row2[24] = "10001"
	row2[47] = "208600000X"

	row3 := make([]string, len(cols))
	row3[0] = "3456789012"
	row3[1] = "2" // Organization
	row3[4] = "ACME HEALTH CORP"
	row3[23] = "CA"
	row3[24] = "94105"
	row3[47] = "207Q00000X"

	content := header + "\n" +
		strings.Join(row1, ",") + "\n" +
		strings.Join(row2, ",") + "\n" +
		strings.Join(row3, ",") + "\n"
	os.WriteFile(filepath.Join(dir, "npidata_pfile_20240101.csv"), []byte(content), 0644)

	// Taxonomy reference CSV
	taxContent := "Code,Grouping,Classification,Specialization,Definition,Notes,Display Name,Section\n" +
		"207Q00000X,Allopathic,Family Medicine,,,,Family Medicine,\n" +
		"208600000X,Allopathic,Surgery,,,,Surgery,\n"
	os.WriteFile(filepath.Join(dir, "nucc_taxonomy_240.csv"), []byte(taxContent), 0644)

	return dir
}

func runCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(cliBinary, args...)
	cmd.Env = append(os.Environ(), "GOWORK=off")
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestCLINoArgs(t *testing.T) {
	stdout, _, err := runCLI(t)
	if err != nil {
		t.Fatalf("expected no error for root command, got: %v", err)
	}
	if !strings.Contains(stdout, "NPPES Data CLI") {
		t.Errorf("expected help text containing 'NPPES Data CLI', got:\n%s", stdout)
	}
}

func TestCLIHelpFlag(t *testing.T) {
	stdout, _, err := runCLI(t, "--help")
	if err != nil {
		t.Fatalf("--help failed: %v", err)
	}
	for _, sub := range []string{"stats", "query", "export", "download"} {
		if !strings.Contains(stdout, sub) {
			t.Errorf("help output missing subcommand %q", sub)
		}
	}
}

func TestCLIStats(t *testing.T) {
	dataDir := makeTestDataDir(t)
	stdout, _, err := runCLI(t, "stats", "--data-dir", dataDir)
	if err != nil {
		t.Fatalf("stats failed: %v", err)
	}
	if !strings.Contains(stdout, "3") {
		t.Errorf("expected total provider count of 3 in output:\n%s", stdout)
	}
}

func TestCLIStatsMissingFlag(t *testing.T) {
	_, stderr, err := runCLI(t, "stats")
	if err == nil {
		t.Fatal("expected error when --data-dir is missing")
	}
	if !strings.Contains(stderr, "data-dir") {
		t.Errorf("expected error about missing data-dir flag, got:\n%s", stderr)
	}
}

func TestCLIQueryByState(t *testing.T) {
	dataDir := makeTestDataDir(t)
	stdout, _, err := runCLI(t, "query", "--data-dir", dataDir, "--state", "CA")
	if err != nil {
		t.Fatalf("query --state CA failed: %v", err)
	}
	if !strings.Contains(stdout, "1234567890") {
		t.Error("expected NPI 1234567890 in CA results")
	}
	if !strings.Contains(stdout, "3456789012") {
		t.Error("expected NPI 3456789012 in CA results")
	}
	if strings.Contains(stdout, "2345678901") {
		t.Error("NY provider should not appear in CA query")
	}
	if !strings.Contains(stdout, "Total matches: 2") {
		t.Errorf("expected 'Total matches: 2' in output:\n%s", stdout)
	}
}

func TestCLIQueryByNPI(t *testing.T) {
	dataDir := makeTestDataDir(t)
	stdout, _, err := runCLI(t, "query", "--data-dir", dataDir, "--npi", "2345678901")
	if err != nil {
		t.Fatalf("query --npi failed: %v", err)
	}
	if !strings.Contains(stdout, "2345678901") {
		t.Error("expected NPI 2345678901 in output")
	}
	if !strings.Contains(stdout, "SMITH") || !strings.Contains(stdout, "JANE") {
		t.Errorf("expected provider name in output:\n%s", stdout)
	}
}

func TestCLIQueryByPostalCode(t *testing.T) {
	dataDir := makeTestDataDir(t)
	stdout, _, err := runCLI(t, "query", "--data-dir", dataDir, "--postal-code", "902")
	if err != nil {
		t.Fatalf("query --postal-code failed: %v", err)
	}
	if !strings.Contains(stdout, "1234567890") {
		t.Error("expected NPI 1234567890 for postal code prefix 902")
	}
	if !strings.Contains(stdout, "Total matches: 1") {
		t.Errorf("expected 'Total matches: 1' in output:\n%s", stdout)
	}
}

func TestCLIQueryLimit(t *testing.T) {
	dataDir := makeTestDataDir(t)
	stdout, _, err := runCLI(t, "query", "--data-dir", dataDir, "--limit", "1")
	if err != nil {
		t.Fatalf("query --limit failed: %v", err)
	}
	// Should print only 1 provider line plus the total line
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	providerLines := 0
	for _, line := range lines {
		if strings.Contains(line, "|") {
			providerLines++
		}
	}
	if providerLines != 1 {
		t.Errorf("expected 1 provider line with --limit 1, got %d:\n%s", providerLines, stdout)
	}
}

func TestCLIExportJSON(t *testing.T) {
	dataDir := makeTestDataDir(t)
	outFile := filepath.Join(t.TempDir(), "out.json")

	stdout, _, err := runCLI(t, "export", "--data-dir", dataDir, "--output", outFile, "--format", "json", "--state", "CA")
	if err != nil {
		t.Fatalf("export json failed: %v", err)
	}
	if !strings.Contains(stdout, "Exported to") {
		t.Errorf("expected 'Exported to' confirmation, got:\n%s", stdout)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	var records []json.RawMessage
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("exported JSON is not a valid array: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("expected 2 CA providers in export, got %d", len(records))
	}
}

func TestCLIExportCSV(t *testing.T) {
	dataDir := makeTestDataDir(t)
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "out.csv")

	_, _, err := runCLI(t, "export", "--data-dir", dataDir, "--output", outFile, "--format", "csv")
	if err != nil {
		t.Fatalf("export csv failed: %v", err)
	}

	// CSV exporter writes to <base>_providers.csv
	providersFile := filepath.Join(tmpDir, "out_providers.csv")
	data, err := os.ReadFile(providersFile)
	if err != nil {
		t.Fatalf("failed to read exported CSV: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	// 3 data rows (no header in providers CSV)
	if len(lines) != 3 {
		t.Errorf("expected 3 records in CSV, got %d", len(lines))
	}
}

func TestCLIExportSQL(t *testing.T) {
	dataDir := makeTestDataDir(t)
	outFile := filepath.Join(t.TempDir(), "out.sql")

	_, _, err := runCLI(t, "export", "--data-dir", dataDir, "--output", outFile, "--format", "sql", "--state", "NY")
	if err != nil {
		t.Fatalf("export sql failed: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read exported SQL: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "CREATE TABLE") {
		t.Error("expected CREATE TABLE in SQL export")
	}
	if !strings.Contains(content, "INSERT") {
		t.Error("expected INSERT statements in SQL export")
	}
	if !strings.Contains(content, "2345678901") {
		t.Error("expected NY provider NPI in SQL export")
	}
}

func TestCLIExportMissingFlags(t *testing.T) {
	_, stderr, err := runCLI(t, "export", "--data-dir", "/tmp")
	if err == nil {
		t.Fatal("expected error when --output is missing")
	}
	if !strings.Contains(stderr, "output") {
		t.Errorf("expected error about missing output flag, got:\n%s", stderr)
	}
}

func TestCLIDownloadMissingFlag(t *testing.T) {
	_, stderr, err := runCLI(t, "download")
	if err == nil {
		t.Fatal("expected error when --out-dir is missing")
	}
	if !strings.Contains(stderr, "out-dir") {
		t.Errorf("expected error about missing out-dir flag, got:\n%s", stderr)
	}
}
