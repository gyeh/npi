package npi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeTestDataset() *NppesDataset {
	ind := EntityTypeIndividual
	org := EntityTypeOrganization
	ca := StateCA
	ny := StateNY

	providers := []NppesRecord{
		{
			NPI:              "1234567890",
			EntityType:       &ind,
			ProviderName:     ProviderName{First: "John", Last: "Doe"},
			MailingAddress:   Address{State: &ca, City: "Los Angeles", PostalCode: "90210"},
			PracticeAddress:  Address{State: &ca, City: "Beverly Hills", PostalCode: "90211"},
			TaxonomyCodes:    []TaxonomyCode{{Code: "207Q00000X", IsPrimary: true}},
			OtherIdentifiers: []OtherIdentifier{},
		},
		{
			NPI:              "2345678901",
			EntityType:       &ind,
			ProviderName:     ProviderName{First: "Jane", Last: "Smith"},
			MailingAddress:   Address{State: &ny, City: "New York", PostalCode: "10001"},
			TaxonomyCodes:    []TaxonomyCode{{Code: "208600000X", IsPrimary: true}},
			OtherIdentifiers: []OtherIdentifier{},
		},
		{
			NPI:              "3456789012",
			EntityType:       &org,
			OrganizationName: OrganizationName{LegalBusinessName: "Acme Health"},
			MailingAddress:   Address{State: &ca, City: "San Francisco", PostalCode: "94105"},
			PracticeAddress:  Address{State: &ca, City: "San Francisco", PostalCode: "94107"},
			TaxonomyCodes:    []TaxonomyCode{{Code: "207Q00000X", IsPrimary: true}},
			OtherIdentifiers: []OtherIdentifier{},
		},
	}

	ds := &NppesDataset{
		Providers: providers,
		TaxonomyMap: map[string]TaxonomyReference{
			"207Q00000X": {Code: "207Q00000X", DisplayName: "Family Medicine"},
			"208600000X": {Code: "208600000X", DisplayName: "Surgery"},
		},
	}
	ds.BuildIndexes()
	return ds
}

func TestDatasetLen(t *testing.T) {
	ds := makeTestDataset()
	if ds.Len() != 3 {
		t.Errorf("Len() = %d, want 3", ds.Len())
	}
}

func TestDatasetIsEmpty(t *testing.T) {
	ds := makeTestDataset()
	if ds.IsEmpty() {
		t.Error("expected non-empty dataset")
	}
	empty := &NppesDataset{}
	if !empty.IsEmpty() {
		t.Error("expected empty dataset")
	}
}

func TestGetByNPI(t *testing.T) {
	ds := makeTestDataset()
	rec := ds.GetByNPI("1234567890")
	if rec == nil {
		t.Fatal("expected to find record")
	}
	if rec.ProviderName.Last != "Doe" {
		t.Errorf("Last = %q, want 'Doe'", rec.ProviderName.Last)
	}

	rec = ds.GetByNPI("9999999999")
	if rec != nil {
		t.Error("expected nil for non-existent NPI")
	}
}

func TestGetByState(t *testing.T) {
	ds := makeTestDataset()
	results := ds.GetByState("CA")
	if len(results) != 2 {
		t.Errorf("GetByState('CA') = %d records, want 2", len(results))
	}
	results = ds.GetByState("NY")
	if len(results) != 1 {
		t.Errorf("GetByState('NY') = %d records, want 1", len(results))
	}
	results = ds.GetByState("TX")
	if len(results) != 0 {
		t.Errorf("GetByState('TX') = %d records, want 0", len(results))
	}
}

func TestGetByTaxonomy(t *testing.T) {
	ds := makeTestDataset()
	results := ds.GetByTaxonomy("207Q00000X")
	if len(results) != 2 {
		t.Errorf("GetByTaxonomy('207Q00000X') = %d records, want 2", len(results))
	}
	results = ds.GetByTaxonomy("208600000X")
	if len(results) != 1 {
		t.Errorf("GetByTaxonomy('208600000X') = %d records, want 1", len(results))
	}
}

func TestGetTaxonomyDescription(t *testing.T) {
	ds := makeTestDataset()
	ref := ds.GetTaxonomyDescription("207Q00000X")
	if ref == nil {
		t.Fatal("expected taxonomy reference")
	}
	if ref.DisplayName != "Family Medicine" {
		t.Errorf("DisplayName = %q, want 'Family Medicine'", ref.DisplayName)
	}

	ref = ds.GetTaxonomyDescription("NOTEXIST")
	if ref != nil {
		t.Error("expected nil for non-existent code")
	}
}

func TestQueryBuilderState(t *testing.T) {
	ds := makeTestDataset()
	results := ds.Query().State("CA").Execute()
	if len(results) != 2 {
		t.Errorf("Query().State('CA').Execute() = %d results, want 2", len(results))
	}
}

func TestQueryBuilderSpecialty(t *testing.T) {
	ds := makeTestDataset()
	results := ds.Query().Specialty("Family Medicine").Execute()
	if len(results) != 2 {
		t.Errorf("Query().Specialty('Family Medicine').Execute() = %d results, want 2", len(results))
	}
	results = ds.Query().Specialty("Surgery").Execute()
	if len(results) != 1 {
		t.Errorf("Query().Specialty('Surgery').Execute() = %d results, want 1", len(results))
	}
}

func TestQueryBuilderActiveOnly(t *testing.T) {
	ds := makeTestDataset()
	results := ds.Query().ActiveOnly().Execute()
	if len(results) != 3 {
		t.Errorf("Query().ActiveOnly().Execute() = %d results, want 3", len(results))
	}
}

func TestQueryBuilderPostalCode(t *testing.T) {
	ds := makeTestDataset()

	// Exact match on mailing address
	results := ds.Query().PostalCode("90210").Execute()
	if len(results) != 1 {
		t.Errorf("Query().PostalCode('90210').Execute() = %d results, want 1", len(results))
	}

	// Prefix match - "902" matches mailing 90210 and practice 90211
	results = ds.Query().PostalCode("902").Execute()
	if len(results) != 1 {
		t.Errorf("Query().PostalCode('902').Execute() = %d results, want 1", len(results))
	}

	// Prefix match across multiple providers - "9" matches CA providers
	results = ds.Query().PostalCode("9").Execute()
	if len(results) != 2 {
		t.Errorf("Query().PostalCode('9').Execute() = %d results, want 2", len(results))
	}

	// Match on practice address only (94107 is only on practice address of provider 3)
	results = ds.Query().PostalCode("94107").Execute()
	if len(results) != 1 {
		t.Errorf("Query().PostalCode('94107').Execute() = %d results, want 1", len(results))
	}

	// No match
	results = ds.Query().PostalCode("99999").Execute()
	if len(results) != 0 {
		t.Errorf("Query().PostalCode('99999').Execute() = %d results, want 0", len(results))
	}
}

func TestQueryBuilderPostalCodeWithState(t *testing.T) {
	ds := makeTestDataset()
	results := ds.Query().State("CA").PostalCode("902").Execute()
	if len(results) != 1 {
		t.Errorf("Query().State('CA').PostalCode('902').Execute() = %d results, want 1", len(results))
	}
	if results[0].ProviderName.Last != "Doe" {
		t.Errorf("expected Doe, got %s", results[0].ProviderName.Last)
	}
}

func TestQueryBuilderChained(t *testing.T) {
	ds := makeTestDataset()
	results := ds.Query().State("CA").Specialty("Family Medicine").Execute()
	if len(results) != 2 {
		t.Errorf("chained query = %d results, want 2", len(results))
	}
}

func TestQueryBuilderCount(t *testing.T) {
	ds := makeTestDataset()
	count := ds.Query().State("CA").Count()
	if count != 2 {
		t.Errorf("Count() = %d, want 2", count)
	}
}

func TestQueryBuilderLimit(t *testing.T) {
	ds := makeTestDataset()
	results := ds.Query().Limit(1)
	if len(results) != 1 {
		t.Errorf("Limit(1) = %d results, want 1", len(results))
	}
}

func TestDatasetStatistics(t *testing.T) {
	ds := makeTestDataset()
	stats := ds.Statistics()
	if stats.TotalProviders != 3 {
		t.Errorf("TotalProviders = %d, want 3", stats.TotalProviders)
	}
	if stats.IndividualProviders != 2 {
		t.Errorf("IndividualProviders = %d, want 2", stats.IndividualProviders)
	}
	if stats.OrganizationProviders != 1 {
		t.Errorf("OrganizationProviders = %d, want 1", stats.OrganizationProviders)
	}
	if stats.ActiveProviders != 3 {
		t.Errorf("ActiveProviders = %d, want 3", stats.ActiveProviders)
	}
	if stats.StatesRepresented != 2 {
		t.Errorf("StatesRepresented = %d, want 2", stats.StatesRepresented)
	}
	if stats.UniqueTaxonomyCodes != 2 {
		t.Errorf("UniqueTaxonomyCodes = %d, want 2", stats.UniqueTaxonomyCodes)
	}
}

func TestLoadStandard(t *testing.T) {
	dir := t.TempDir()

	// Create main CSV
	cols := MainSchemaColumnNames()
	header := strings.Join(cols, ",")
	row := make([]string, len(cols))
	row[0] = "1234567890"
	row[1] = "1"
	row[5] = "DOE"
	row[6] = "JOHN"
	row[23] = "CA"
	content := header + "\n" + strings.Join(row, ",") + "\n"
	os.WriteFile(filepath.Join(dir, "npidata_pfile_20240101.csv"), []byte(content), 0644)

	// Create taxonomy CSV
	taxContent := "Code,Grouping,Classification,Specialization,Definition,Notes,Display Name,Section\n" +
		"207Q00000X,,,,,,,\n"
	os.WriteFile(filepath.Join(dir, "nucc_taxonomy_240.csv"), []byte(taxContent), 0644)

	ds, err := LoadStandard(dir)
	if err != nil {
		t.Fatalf("LoadStandard error: %v", err)
	}
	if ds.Len() != 1 {
		t.Errorf("Len() = %d, want 1", ds.Len())
	}
	if ds.TaxonomyMap == nil {
		t.Error("expected taxonomy map")
	}
}

func TestLoadStandardNotDirectory(t *testing.T) {
	_, err := LoadStandard("/nonexistent/directory")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

func TestDatasetInterface(t *testing.T) {
	// Compile-time check that NppesDataset implements Dataset
	var _ Dataset = (*NppesDataset)(nil)
}
