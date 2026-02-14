package npi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestMainCSV(t *testing.T, dir string) string {
	t.Helper()
	cols := MainSchemaColumnNames()
	header := strings.Join(cols, ",")

	// Build a row with 330 fields
	row := make([]string, len(cols))
	row[0] = "1234567890"   // NPI
	row[1] = "1"            // Entity Type Code
	row[4] = ""             // Org Name
	row[5] = "DOE"          // Last Name
	row[6] = "JOHN"         // First Name
	row[7] = "A"            // Middle Name
	row[20] = "123 MAIN ST" // Mailing Line 1
	row[22] = "ANYTOWN"     // City
	row[23] = "CA"          // State
	row[24] = "90001"       // Postal Code
	row[25] = "US"          // Country
	row[36] = "01/01/2020"  // Enumeration Date
	row[37] = "06/15/2023"  // Last Update Date
	row[41] = "M"           // Sex Code
	row[47] = "207Q00000X"  // Taxonomy Code 1
	row[50] = "Y"           // Primary Switch 1

	path := filepath.Join(dir, "npidata_pfile_test.csv")
	content := header + "\n" + strings.Join(row, ",") + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadMainDataBasic(t *testing.T) {
	dir := t.TempDir()
	csvPath := createTestMainCSV(t, dir)

	reader := NewNppesReader()
	records, err := reader.LoadMainData(csvPath)
	if err != nil {
		t.Fatalf("LoadMainData error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	rec := records[0]
	if rec.NPI != "1234567890" {
		t.Errorf("NPI = %q, want '1234567890'", rec.NPI)
	}
	if rec.EntityType == nil || *rec.EntityType != EntityTypeIndividual {
		t.Errorf("EntityType = %v, want Individual", rec.EntityType)
	}
	if rec.ProviderName.Last != "DOE" {
		t.Errorf("Last = %q, want 'DOE'", rec.ProviderName.Last)
	}
	if rec.ProviderName.First != "JOHN" {
		t.Errorf("First = %q, want 'JOHN'", rec.ProviderName.First)
	}
	if rec.MailingAddress.State == nil || *rec.MailingAddress.State != StateCA {
		t.Errorf("State = %v, want CA", rec.MailingAddress.State)
	}
	if rec.ProviderGender == nil || *rec.ProviderGender != SexMale {
		t.Errorf("Gender = %v, want M", rec.ProviderGender)
	}
	if len(rec.TaxonomyCodes) != 1 || rec.TaxonomyCodes[0].Code != "207Q00000X" {
		t.Errorf("TaxonomyCodes = %v", rec.TaxonomyCodes)
	}
	if !rec.TaxonomyCodes[0].IsPrimary {
		t.Error("expected first taxonomy to be primary")
	}
	if !rec.IsActive() {
		t.Error("expected record to be active")
	}
	if rec.EnumerationDate == nil {
		t.Error("expected enumeration date")
	}
}

func TestLoadMainDataFileNotFound(t *testing.T) {
	reader := NewNppesReader()
	_, err := reader.LoadMainData("/nonexistent/file.csv")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadTaxonomyData(t *testing.T) {
	dir := t.TempDir()
	content := "Code,Grouping,Classification,Specialization,Definition,Notes,Display Name,Section\n" +
		"207Q00000X,Allopathic & Osteopathic Physicians,Family Medicine,,A doc,,Family Medicine,Individual\n"
	path := filepath.Join(dir, "nucc_taxonomy_test.csv")
	os.WriteFile(path, []byte(content), 0644)

	reader := NewNppesReader()
	records, err := reader.LoadTaxonomyData(path)
	if err != nil {
		t.Fatalf("LoadTaxonomyData error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Code != "207Q00000X" {
		t.Errorf("Code = %q", records[0].Code)
	}
	if records[0].DisplayName != "Family Medicine" {
		t.Errorf("DisplayName = %q", records[0].DisplayName)
	}
}

func TestLoadOtherNameData(t *testing.T) {
	dir := t.TempDir()
	content := "NPI,Provider Other Organization Name,Provider Other Organization Name Type Code\n" +
		"1234567890,ACME CLINIC,3\n"
	path := filepath.Join(dir, "othername_pfile_test.csv")
	os.WriteFile(path, []byte(content), 0644)

	reader := NewNppesReader()
	records, err := reader.LoadOtherNameData(path)
	if err != nil {
		t.Fatalf("LoadOtherNameData error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].NPI != "1234567890" {
		t.Errorf("NPI = %q", records[0].NPI)
	}
	if records[0].ProviderOtherOrganizationName != "ACME CLINIC" {
		t.Errorf("Name = %q", records[0].ProviderOtherOrganizationName)
	}
}

func TestLoadPracticeLocationData(t *testing.T) {
	dir := t.TempDir()
	cols := PracticeLocationSchemaColumnNames()
	header := strings.Join(cols, ",")
	row := make([]string, len(cols))
	row[0] = "1234567890"
	row[1] = "456 OAK AVE"
	row[3] = "PORTLAND"
	row[4] = "OR"
	row[5] = "97201"
	row[6] = "US"
	content := header + "\n" + strings.Join(row, ",") + "\n"
	path := filepath.Join(dir, "pl_pfile_test.csv")
	os.WriteFile(path, []byte(content), 0644)

	reader := NewNppesReader()
	records, err := reader.LoadPracticeLocationData(path)
	if err != nil {
		t.Fatalf("LoadPracticeLocationData error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Address.City != "PORTLAND" {
		t.Errorf("City = %q", records[0].Address.City)
	}
}

func TestLoadEndpointData(t *testing.T) {
	dir := t.TempDir()
	cols := EndpointSchemaColumnNames()
	header := strings.Join(cols, ",")
	row := make([]string, len(cols))
	row[0] = "1234567890"
	row[1] = "DIRECT"
	row[2] = "Direct Messaging"
	row[3] = "provider@example.com"
	row[4] = "Y"
	content := header + "\n" + strings.Join(row, ",") + "\n"
	path := filepath.Join(dir, "endpoint_pfile_test.csv")
	os.WriteFile(path, []byte(content), 0644)

	reader := NewNppesReader()
	records, err := reader.LoadEndpointData(path)
	if err != nil {
		t.Fatalf("LoadEndpointData error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].EndpointType != "DIRECT" {
		t.Errorf("EndpointType = %q", records[0].EndpointType)
	}
	if records[0].Affiliation == nil || !*records[0].Affiliation {
		t.Error("expected affiliation to be true")
	}
}

func TestLoadMainDataSkipInvalid(t *testing.T) {
	dir := t.TempDir()
	cols := MainSchemaColumnNames()
	header := strings.Join(cols, ",")

	// Valid row
	row1 := make([]string, len(cols))
	row1[0] = "1234567890"
	row1[1] = "1"
	row1[5] = "DOE"
	row1[6] = "JOHN"

	// Invalid row (bad NPI)
	row2 := make([]string, len(cols))
	row2[0] = "BAD"
	row2[1] = "1"

	// Another valid row
	row3 := make([]string, len(cols))
	row3[0] = "9876543210"
	row3[1] = "2"
	row3[4] = "ACME CORP"

	content := header + "\n" +
		strings.Join(row1, ",") + "\n" +
		strings.Join(row2, ",") + "\n" +
		strings.Join(row3, ",") + "\n"

	path := filepath.Join(dir, "npidata_pfile_test.csv")
	os.WriteFile(path, []byte(content), 0644)

	reader := NewNppesReader()
	reader.SkipInvalidRecords = true
	records, err := reader.LoadMainData(path)
	if err != nil {
		t.Fatalf("LoadMainData error: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("expected 2 valid records, got %d", len(records))
	}
}

func TestParseDateValid(t *testing.T) {
	d, err := parseDate("01/15/2020")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil date")
	}
	if d.Year() != 2020 || d.Month() != 1 || d.Day() != 15 {
		t.Errorf("date = %v, want 2020-01-15", d)
	}
}

func TestParseDateEmpty(t *testing.T) {
	d, err := parseDate("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != nil {
		t.Error("expected nil for empty date string")
	}
}

func TestParseDateInvalid(t *testing.T) {
	_, err := parseDate("not-a-date")
	if err == nil {
		t.Error("expected error for invalid date")
	}
}
