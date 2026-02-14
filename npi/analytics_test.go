package npi

import (
	"testing"
)

func makeTestAnalytics() *NppesAnalytics {
	ind := EntityTypeIndividual
	org := EntityTypeOrganization
	ca := StateCA
	ny := StateNY
	tx := StateTX

	providers := []NppesRecord{
		{
			NPI:              "1111111111",
			EntityType:       &ind,
			ProviderName:     ProviderName{First: "Alice", Last: "Johnson"},
			MailingAddress:   Address{State: &ca},
			TaxonomyCodes:    []TaxonomyCode{{Code: "207Q00000X", IsPrimary: true}},
			OtherIdentifiers: []OtherIdentifier{},
		},
		{
			NPI:              "2222222222",
			EntityType:       &ind,
			ProviderName:     ProviderName{First: "Bob", Last: "Williams"},
			MailingAddress:   Address{State: &ny},
			TaxonomyCodes:    []TaxonomyCode{{Code: "208600000X", IsPrimary: true}},
			OtherIdentifiers: []OtherIdentifier{},
		},
		{
			NPI:              "3333333333",
			EntityType:       &org,
			OrganizationName: OrganizationName{LegalBusinessName: "City Hospital"},
			MailingAddress:   Address{State: &ca},
			TaxonomyCodes:    []TaxonomyCode{{Code: "207Q00000X", IsPrimary: true}},
			OtherIdentifiers: []OtherIdentifier{},
		},
		{
			NPI:              "4444444444",
			EntityType:       &ind,
			ProviderName:     ProviderName{First: "Carol", Last: "Davis"},
			MailingAddress:   Address{State: &tx},
			TaxonomyCodes:    []TaxonomyCode{{Code: "207Q00000X", IsPrimary: true}},
			OtherIdentifiers: []OtherIdentifier{},
		},
	}

	return NewNppesAnalytics(providers)
}

func TestFindByNPI(t *testing.T) {
	a := makeTestAnalytics()
	rec := a.FindByNPI("1111111111")
	if rec == nil {
		t.Fatal("expected to find record")
	}
	if rec.ProviderName.First != "Alice" {
		t.Errorf("First = %q, want 'Alice'", rec.ProviderName.First)
	}

	rec = a.FindByNPI("9999999999")
	if rec != nil {
		t.Error("expected nil for non-existent NPI")
	}
}

func TestFindByName(t *testing.T) {
	a := makeTestAnalytics()
	results := a.FindByName("alice")
	if len(results) != 1 {
		t.Errorf("FindByName('alice') = %d results, want 1", len(results))
	}

	results = a.FindByName("hospital")
	if len(results) != 1 {
		t.Errorf("FindByName('hospital') = %d results, want 1", len(results))
	}

	results = a.FindByName("nonexistent")
	if len(results) != 0 {
		t.Errorf("FindByName('nonexistent') = %d results, want 0", len(results))
	}
}

func TestFindByState(t *testing.T) {
	a := makeTestAnalytics()
	results := a.FindByState("CA")
	if len(results) != 2 {
		t.Errorf("FindByState('CA') = %d results, want 2", len(results))
	}
	results = a.FindByState("TX")
	if len(results) != 1 {
		t.Errorf("FindByState('TX') = %d results, want 1", len(results))
	}
}

func TestFindByTaxonomyCode(t *testing.T) {
	a := makeTestAnalytics()
	results := a.FindByTaxonomyCode("207Q00000X")
	if len(results) != 3 {
		t.Errorf("FindByTaxonomyCode('207Q00000X') = %d results, want 3", len(results))
	}
	results = a.FindByTaxonomyCode("208600000X")
	if len(results) != 1 {
		t.Errorf("FindByTaxonomyCode('208600000X') = %d results, want 1", len(results))
	}
}

func TestProviderCountByState(t *testing.T) {
	a := makeTestAnalytics()
	counts := a.ProviderCountByState()
	if len(counts) != 3 {
		t.Fatalf("expected 3 states, got %d", len(counts))
	}
	// CA should be first (2 providers)
	if counts[0].Key != "CA" || counts[0].Count != 2 {
		t.Errorf("top state = %v, want CA:2", counts[0])
	}
}

func TestProviderCountByTaxonomy(t *testing.T) {
	a := makeTestAnalytics()
	counts := a.ProviderCountByTaxonomy()
	if len(counts) != 2 {
		t.Fatalf("expected 2 taxonomy codes, got %d", len(counts))
	}
	if counts[0].Key != "207Q00000X" || counts[0].Count != 3 {
		t.Errorf("top taxonomy = %v, want 207Q00000X:3", counts[0])
	}
}

func TestTopStatesByProviderCount(t *testing.T) {
	a := makeTestAnalytics()
	top := a.TopStatesByProviderCount(2)
	if len(top) != 2 {
		t.Errorf("expected 2 entries, got %d", len(top))
	}
	if top[0].Key != "CA" {
		t.Errorf("top state = %q, want 'CA'", top[0].Key)
	}
}

func TestTopTaxonomyCodesByProviderCount(t *testing.T) {
	a := makeTestAnalytics()
	top := a.TopTaxonomyCodesByProviderCount(1)
	if len(top) != 1 {
		t.Errorf("expected 1 entry, got %d", len(top))
	}
	if top[0].Key != "207Q00000X" {
		t.Errorf("top taxonomy = %q, want '207Q00000X'", top[0].Key)
	}
}

func TestDatasetStats(t *testing.T) {
	a := makeTestAnalytics()
	stats := a.ComputeDatasetStats()
	if stats.TotalProviders != 4 {
		t.Errorf("TotalProviders = %d, want 4", stats.TotalProviders)
	}
	if stats.IndividualProviders != 3 {
		t.Errorf("IndividualProviders = %d, want 3", stats.IndividualProviders)
	}
	if stats.OrganizationProviders != 1 {
		t.Errorf("OrganizationProviders = %d, want 1", stats.OrganizationProviders)
	}
	if stats.ActiveProviders != 4 {
		t.Errorf("ActiveProviders = %d, want 4", stats.ActiveProviders)
	}
	if stats.UniqueStates != 3 {
		t.Errorf("UniqueStates = %d, want 3", stats.UniqueStates)
	}
	if stats.UniqueTaxonomyCodes != 2 {
		t.Errorf("UniqueTaxonomyCodes = %d, want 2", stats.UniqueTaxonomyCodes)
	}
}
