package npi

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLatestNPPESUrl(t *testing.T) {
	url := LatestNPPESUrl()
	if !strings.HasPrefix(url, "https://download.cms.gov/nppes/NPPES_Data_Dissemination_") {
		t.Errorf("unexpected URL: %s", url)
	}
	if !strings.HasSuffix(url, "_V2.zip") {
		t.Errorf("URL should end with _V2.zip: %s", url)
	}
}

func TestDefaultDownloadConfig(t *testing.T) {
	config := DefaultDownloadConfig()
	if config.TimeoutSeconds != 300 {
		t.Errorf("TimeoutSeconds = %d, want 300", config.TimeoutSeconds)
	}
	if config.MaxFileSize != 20*1024*1024*1024 {
		t.Errorf("MaxFileSize = %d", config.MaxFileSize)
	}
	if config.UserAgent == "" {
		t.Error("expected non-empty user agent")
	}
}

func TestExtractedFilesSummary(t *testing.T) {
	ef := &ExtractedFiles{}
	if got := ef.Summary(); got != "No recognized NPPES files found" {
		t.Errorf("Summary() = %q", got)
	}

	ef.MainDataFile = "/tmp/npidata.csv"
	ef.TaxonomyFile = "/tmp/taxonomy.csv"
	if got := ef.Summary(); !strings.Contains(got, "Main Data") || !strings.Contains(got, "Taxonomy") {
		t.Errorf("Summary() = %q", got)
	}
}

func TestExtractedFilesHasMainData(t *testing.T) {
	ef := &ExtractedFiles{}
	if ef.HasMainData() {
		t.Error("expected false when no main data file")
	}
	ef.MainDataFile = "/tmp/npidata.csv"
	if !ef.HasMainData() {
		t.Error("expected true when main data file is set")
	}
}

func TestExtractZip(t *testing.T) {
	dir := t.TempDir()

	// Create a test zip file
	zipPath := filepath.Join(dir, "test.zip")
	zf, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(zf)

	// Add a fake main data file
	f, _ := w.Create("npidata_pfile_20240101.csv")
	f.Write([]byte("NPI,Entity Type Code\n1234567890,1\n"))

	// Add a fake taxonomy file
	f2, _ := w.Create("nucc_taxonomy_240.csv")
	f2.Write([]byte("Code,Grouping\n207Q00000X,Test\n"))

	// Add an endpoint file
	f3, _ := w.Create("endpoint_pfile_20240101.csv")
	f3.Write([]byte("NPI,Endpoint Type\n1234567890,DIRECT\n"))

	w.Close()
	zf.Close()

	// Extract
	extractDir := filepath.Join(dir, "extracted")
	os.MkdirAll(extractDir, 0755)

	downloader := &NppesDownloader{
		Config: DownloadConfig{OutputDir: extractDir},
	}
	extracted, err := downloader.ExtractZip(zipPath)
	if err != nil {
		t.Fatalf("ExtractZip error: %v", err)
	}

	if !extracted.HasMainData() {
		t.Error("expected main data file to be found")
	}
	if extracted.TaxonomyFile == "" {
		t.Error("expected taxonomy file to be found")
	}
	if extracted.EndpointsFile == "" {
		t.Error("expected endpoints file to be found")
	}
	if len(extracted.Files) != 3 {
		t.Errorf("expected 3 files, got %d", len(extracted.Files))
	}
}

func TestFileCategorization(t *testing.T) {
	tests := []struct {
		filename string
		field    string
	}{
		{"npidata_pfile_20240101-20240107.csv", "main"},
		{"nucc_taxonomy_240.csv", "taxonomy"},
		{"othername_pfile_20240101-20240107.csv", "other_names"},
		{"pl_pfile_20240101-20240107.csv", "practice_locations"},
		{"endpoint_pfile_20240101-20240107.csv", "endpoints"},
		{"npidata_pfile_20240101-20240107_fileheader.csv", ""},
	}

	for _, tt := range tests {
		lower := strings.ToLower(tt.filename)
		var matched string
		if strings.Contains(lower, "npidata_pfile") && strings.HasSuffix(lower, ".csv") && !strings.Contains(lower, "_fileheader") {
			matched = "main"
		} else if strings.Contains(lower, "nucc_taxonomy") && strings.HasSuffix(lower, ".csv") && !strings.Contains(lower, "_fileheader") {
			matched = "taxonomy"
		} else if strings.Contains(lower, "othername_pfile") && strings.HasSuffix(lower, ".csv") && !strings.Contains(lower, "_fileheader") {
			matched = "other_names"
		} else if strings.Contains(lower, "pl_pfile") && strings.HasSuffix(lower, ".csv") && !strings.Contains(lower, "_fileheader") {
			matched = "practice_locations"
		} else if strings.Contains(lower, "endpoint_pfile") && strings.HasSuffix(lower, ".csv") && !strings.Contains(lower, "_fileheader") {
			matched = "endpoints"
		}

		if matched != tt.field {
			t.Errorf("file %q categorized as %q, want %q", tt.filename, matched, tt.field)
		}
	}
}
