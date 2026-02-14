package npi

import (
	"testing"
)

func TestMainSchemaColumnCount(t *testing.T) {
	cols := MainSchemaColumnNames()
	if len(cols) != MainSchemaColumnCount {
		t.Errorf("MainSchemaColumnNames() has %d columns, want %d", len(cols), MainSchemaColumnCount)
	}
}

func TestMainSchemaKeyColumns(t *testing.T) {
	cols := MainSchemaColumnNames()
	if cols[0] != "NPI" {
		t.Errorf("first column = %q, want 'NPI'", cols[0])
	}
	if cols[1] != "Entity Type Code" {
		t.Errorf("second column = %q, want 'Entity Type Code'", cols[1])
	}
	if cols[329] != "Certification Date" {
		t.Errorf("last column = %q, want 'Certification Date'", cols[329])
	}
}

func TestMainSchemaTaxonomyColumns(t *testing.T) {
	cols := MainSchemaColumnNames()
	// First taxonomy code at index 47
	if cols[47] != "Healthcare Provider Taxonomy Code_1" {
		t.Errorf("column 47 = %q, want 'Healthcare Provider Taxonomy Code_1'", cols[47])
	}
	// Last taxonomy column at index 106
	if cols[106] != "Healthcare Provider Primary Taxonomy Switch_15" {
		t.Errorf("column 106 = %q, want 'Healthcare Provider Primary Taxonomy Switch_15'", cols[106])
	}
}

func TestMainSchemaOtherIdentifierColumns(t *testing.T) {
	cols := MainSchemaColumnNames()
	// First other identifier at index 107
	if cols[107] != "Other Provider Identifier_1" {
		t.Errorf("column 107 = %q, want 'Other Provider Identifier_1'", cols[107])
	}
	// Last other identifier at index 306
	if cols[306] != "Other Provider Identifier Issuer_50" {
		t.Errorf("column 306 = %q, want 'Other Provider Identifier Issuer_50'", cols[306])
	}
}

func TestOtherNameSchemaColumnCount(t *testing.T) {
	cols := OtherNameSchemaColumnNames()
	if len(cols) != OtherNameSchemaColumnCount {
		t.Errorf("OtherNameSchemaColumnNames() has %d columns, want %d", len(cols), OtherNameSchemaColumnCount)
	}
}

func TestPracticeLocationSchemaColumnCount(t *testing.T) {
	cols := PracticeLocationSchemaColumnNames()
	if len(cols) != PracticeLocationSchemaColumnCount {
		t.Errorf("PracticeLocationSchemaColumnNames() has %d columns, want %d", len(cols), PracticeLocationSchemaColumnCount)
	}
}

func TestEndpointSchemaColumnCount(t *testing.T) {
	cols := EndpointSchemaColumnNames()
	if len(cols) != EndpointSchemaColumnCount {
		t.Errorf("EndpointSchemaColumnNames() has %d columns, want %d", len(cols), EndpointSchemaColumnCount)
	}
}

func TestTaxonomySchemaColumnCount(t *testing.T) {
	cols := TaxonomySchemaColumnNames()
	if len(cols) != TaxonomySchemaColumnCount {
		t.Errorf("TaxonomySchemaColumnNames() has %d columns, want %d", len(cols), TaxonomySchemaColumnCount)
	}
}

func TestValidateHeadersMatch(t *testing.T) {
	expected := []string{"A", "B", "C"}
	actual := []string{"A", "B", "C"}
	if err := ValidateHeaders(expected, actual); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateHeadersCountMismatch(t *testing.T) {
	expected := []string{"A", "B", "C"}
	actual := []string{"A", "B"}
	err := ValidateHeaders(expected, actual)
	if err == nil {
		t.Error("expected error for column count mismatch")
	}
}

func TestValidateHeadersNameMismatch(t *testing.T) {
	expected := []string{"A", "B", "C"}
	actual := []string{"A", "X", "C"}
	err := ValidateHeaders(expected, actual)
	if err == nil {
		t.Error("expected error for column name mismatch")
	}
}
