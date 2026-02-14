package npi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func makeExportProviders() []*NppesRecord {
	ind := EntityTypeIndividual
	org := EntityTypeOrganization
	ca := StateCA
	ny := StateNY
	enumDate := time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC)

	return []*NppesRecord{
		{
			NPI:              "1234567890",
			EntityType:       &ind,
			ProviderName:     ProviderName{First: "John", Last: "Doe"},
			MailingAddress:   Address{State: &ca, City: "LA", Line1: "123 Main"},
			EnumerationDate:  &enumDate,
			TaxonomyCodes:    []TaxonomyCode{{Code: "207Q00000X", IsPrimary: true}},
			OtherIdentifiers: []OtherIdentifier{},
		},
		{
			NPI:              "2345678901",
			EntityType:       &org,
			OrganizationName: OrganizationName{LegalBusinessName: "Acme Health"},
			MailingAddress:   Address{State: &ny, City: "NYC"},
			TaxonomyCodes:    []TaxonomyCode{{Code: "208600000X", IsPrimary: true}},
			OtherIdentifiers: []OtherIdentifier{},
		},
	}
}

func TestExportJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "export.json")
	providers := makeExportProviders()

	err := ExportJSON(providers, path)
	if err != nil {
		t.Fatalf("ExportJSON error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}
}

func TestExportCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "export.csv")
	providers := makeExportProviders()

	err := ExportCSV(providers, path)
	if err != nil {
		t.Fatalf("ExportCSV error: %v", err)
	}

	// Check providers file
	providersPath := filepath.Join(dir, "export_providers.csv")
	data, err := os.ReadFile(providersPath)
	if err != nil {
		t.Fatalf("failed to read providers file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines in providers CSV, got %d", len(lines))
	}

	// Check taxonomies file
	taxPath := filepath.Join(dir, "export_taxonomies.csv")
	data, err = os.ReadFile(taxPath)
	if err != nil {
		t.Fatalf("failed to read taxonomies file: %v", err)
	}
	lines = strings.Split(strings.TrimSpace(string(data)), "\n")
	// 1 header + 2 records
	if len(lines) != 3 {
		t.Errorf("expected 3 lines in taxonomies CSV, got %d", len(lines))
	}
}

func TestExportSQL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "export.sql")
	providers := makeExportProviders()

	err := ExportSQL(providers, path)
	if err != nil {
		t.Fatalf("ExportSQL error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read SQL file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "CREATE TABLE") {
		t.Error("expected CREATE TABLE statement")
	}
	if !strings.Contains(content, "INSERT INTO") {
		t.Error("expected INSERT INTO statement")
	}
	if !strings.Contains(content, "1234567890") {
		t.Error("expected NPI in output")
	}
}

func TestExportSubset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subset.json")

	ds := makeTestDataset()
	err := ExportSubset(ds, path, func(p *NppesRecord) bool {
		return p.MailingAddress.State != nil && *p.MailingAddress.State == StateCA
	}, ExportFormatJSON)
	if err != nil {
		t.Fatalf("ExportSubset error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("expected 2 CA records, got %d", len(records))
	}
}

func TestNewExporter(t *testing.T) {
	exp, err := NewExporter(ExportFormatJSON)
	if err != nil {
		t.Fatalf("NewExporter(JSON) error: %v", err)
	}
	if _, ok := exp.(*JSONExporter); !ok {
		t.Error("expected *JSONExporter")
	}

	exp, err = NewExporter(ExportFormatCSV)
	if err != nil {
		t.Fatalf("NewExporter(CSV) error: %v", err)
	}
	if _, ok := exp.(*CSVExporter); !ok {
		t.Error("expected *CSVExporter")
	}

	exp, err = NewExporter(ExportFormatSQL)
	if err != nil {
		t.Fatalf("NewExporter(SQL) error: %v", err)
	}
	if _, ok := exp.(*SQLExporter); !ok {
		t.Error("expected *SQLExporter")
	}

	_, err = NewExporter("invalid")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}
