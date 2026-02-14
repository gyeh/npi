package npi

import (
	"testing"
)

func TestValidateNPI(t *testing.T) {
	tests := []struct {
		npi   string
		valid bool
	}{
		{"1234567890", true},
		{"0000000000", true},
		{"9999999999", true},
		{"123456789", false},   // too short
		{"12345678901", false}, // too long
		{"123456789a", false},  // contains letter
		{"", false},
		{"abcdefghij", false},
	}
	for _, tt := range tests {
		err := ValidateNPI(tt.npi)
		if (err == nil) != tt.valid {
			t.Errorf("ValidateNPI(%q) valid=%v, want %v", tt.npi, err == nil, tt.valid)
		}
	}
}

func TestNewNPI(t *testing.T) {
	npi, err := NewNPI("1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if npi.String() != "1234567890" {
		t.Errorf("got %q, want %q", npi.String(), "1234567890")
	}

	_, err = NewNPI("bad")
	if err == nil {
		t.Error("expected error for invalid NPI")
	}
}

func TestEntityTypeFromCode(t *testing.T) {
	et, err := EntityTypeFromCode("1")
	if err != nil || et != EntityTypeIndividual {
		t.Errorf("EntityTypeFromCode('1') = %v, %v; want Individual, nil", et, err)
	}

	et, err = EntityTypeFromCode("2")
	if err != nil || et != EntityTypeOrganization {
		t.Errorf("EntityTypeFromCode('2') = %v, %v; want Organization, nil", et, err)
	}

	_, err = EntityTypeFromCode("3")
	if err == nil {
		t.Error("expected error for invalid entity type code")
	}
}

func TestEntityTypeToCode(t *testing.T) {
	if EntityTypeIndividual.ToCode() != "1" {
		t.Errorf("Individual.ToCode() = %q, want '1'", EntityTypeIndividual.ToCode())
	}
	if EntityTypeOrganization.ToCode() != "2" {
		t.Errorf("Organization.ToCode() = %q, want '2'", EntityTypeOrganization.ToCode())
	}
}

func TestStateCodeFromCode(t *testing.T) {
	tests := []struct {
		code string
		ok   bool
		want StateCode
	}{
		{"CA", true, StateCA},
		{"ca", true, StateCA},
		{"NY", true, StateNY},
		{"ZZ", true, StateZZ},
		{"XX", false, ""},
		{"", false, ""},
	}
	for _, tt := range tests {
		sc, ok := StateCodeFromCode(tt.code)
		if ok != tt.ok {
			t.Errorf("StateCodeFromCode(%q) ok=%v, want %v", tt.code, ok, tt.ok)
		}
		if ok && sc != tt.want {
			t.Errorf("StateCodeFromCode(%q) = %v, want %v", tt.code, sc, tt.want)
		}
	}
}

func TestSexCodeFromCode(t *testing.T) {
	tests := []struct {
		code string
		ok   bool
		want SexCode
	}{
		{"M", true, SexMale},
		{"F", true, SexFemale},
		{"U", true, SexUndisclosed},
		{"X", true, SexUndisclosed},
		{"Z", false, ""},
	}
	for _, tt := range tests {
		sc, ok := SexCodeFromCode(tt.code)
		if ok != tt.ok || (ok && sc != tt.want) {
			t.Errorf("SexCodeFromCode(%q) = %v, %v; want %v, %v", tt.code, sc, ok, tt.want, tt.ok)
		}
	}
}

func TestDeactivationReasonCodeFromCode(t *testing.T) {
	tests := []struct {
		code string
		ok   bool
		want DeactivationReasonCode
	}{
		{"Death", true, DeactivationDeath},
		{"death", true, DeactivationDeath},
		{"DEATH", true, DeactivationDeath},
		{"Disbandment", true, DeactivationDisbandment},
		{"Fraud", true, DeactivationFraud},
		{"Other", true, DeactivationOther},
		{"u", true, DeactivationUndisclosed},
		{"Undisclosed", true, DeactivationUndisclosed},
		{"x", true, DeactivationUndisclosed},
		{"invalid", false, ""},
	}
	for _, tt := range tests {
		dr, ok := DeactivationReasonCodeFromCode(tt.code)
		if ok != tt.ok || (ok && dr != tt.want) {
			t.Errorf("DeactivationReasonCodeFromCode(%q) = %v, %v; want %v, %v", tt.code, dr, ok, tt.want, tt.ok)
		}
	}
}

func TestOtherProviderNameTypeCodeFromCode(t *testing.T) {
	tests := []struct {
		code string
		ok   bool
		want OtherProviderNameTypeCode
	}{
		{"1", true, NameTypeFormerName},
		{"2", true, NameTypeProfessionalName},
		{"3", true, NameTypeDoingBusinessAs},
		{"4", true, NameTypeFormerLegalBusinessName},
		{"5", true, NameTypeOtherName},
		{"6", false, ""},
	}
	for _, tt := range tests {
		o, ok := OtherProviderNameTypeCodeFromCode(tt.code)
		if ok != tt.ok || (ok && o != tt.want) {
			t.Errorf("OtherProviderNameTypeCodeFromCode(%q) = %v, %v; want %v, %v", tt.code, o, ok, tt.want, tt.ok)
		}
	}
}

func TestNamePrefixCodeFromCode(t *testing.T) {
	tests := []struct {
		code string
		ok   bool
		want NamePrefixCode
	}{
		{"Ms.", true, PrefixMs},
		{"Mr.", true, PrefixMr},
		{"Miss", true, PrefixMiss},
		{"Mrs.", true, PrefixMrs},
		{"Dr.", true, PrefixDr},
		{"Prof.", true, PrefixProf},
		{"Esq.", false, ""},
	}
	for _, tt := range tests {
		p, ok := NamePrefixCodeFromCode(tt.code)
		if ok != tt.ok || (ok && p != tt.want) {
			t.Errorf("NamePrefixCodeFromCode(%q) = %v, %v; want %v, %v", tt.code, p, ok, tt.want, tt.ok)
		}
	}
}

func TestNameSuffixCodeFromCode(t *testing.T) {
	tests := []struct {
		code string
		ok   bool
		want NameSuffixCode
	}{
		{"Jr.", true, SuffixJr},
		{"Sr.", true, SuffixSr},
		{"I", true, SuffixI},
		{"II", true, SuffixII},
		{"III", true, SuffixIII},
		{"IV", true, SuffixIV},
		{"V", true, SuffixV},
		{"VIII", true, SuffixVIII},
		{"X", true, SuffixX},
		{"PhD", false, ""},
	}
	for _, tt := range tests {
		s, ok := NameSuffixCodeFromCode(tt.code)
		if ok != tt.ok || (ok && s != tt.want) {
			t.Errorf("NameSuffixCodeFromCode(%q) = %v, %v; want %v, %v", tt.code, s, ok, tt.want, tt.ok)
		}
	}
}

func TestSoleProprietorCodeFromCode(t *testing.T) {
	tests := []struct {
		code string
		ok   bool
		want SoleProprietorCode
	}{
		{"X", true, SoleProprietorNotAnswered},
		{"Y", true, SoleProprietorYes},
		{"N", true, SoleProprietorNo},
		{"Z", false, ""},
	}
	for _, tt := range tests {
		s, ok := SoleProprietorCodeFromCode(tt.code)
		if ok != tt.ok || (ok && s != tt.want) {
			t.Errorf("SoleProprietorCodeFromCode(%q) = %v, %v; want %v, %v", tt.code, s, ok, tt.want, tt.ok)
		}
	}
}

func TestSubpartCodeFromCode(t *testing.T) {
	s, ok := SubpartCodeFromCode("Y")
	if !ok || s != SubpartYes {
		t.Errorf("SubpartCodeFromCode('Y') = %v, %v", s, ok)
	}
}

func TestPrimaryTaxonomySwitchFromCode(t *testing.T) {
	p, ok := PrimaryTaxonomySwitchFromCode("Y")
	if !ok || p != PrimaryTaxonomyYes {
		t.Errorf("PrimaryTaxonomySwitchFromCode('Y') = %v, %v", p, ok)
	}
	p, ok = PrimaryTaxonomySwitchFromCode("N")
	if !ok || p != PrimaryTaxonomyNo {
		t.Errorf("PrimaryTaxonomySwitchFromCode('N') = %v, %v", p, ok)
	}
}

func TestGroupTaxonomyCodeFromCode(t *testing.T) {
	g, ok := GroupTaxonomyCodeFromCode("193200000X")
	if !ok || g != GroupTaxonomyMultiSpecialty {
		t.Errorf("GroupTaxonomyCodeFromCode('193200000X') = %v, %v", g, ok)
	}
	g, ok = GroupTaxonomyCodeFromCode("193400000X")
	if !ok || g != GroupTaxonomySingleSpecialty {
		t.Errorf("GroupTaxonomyCodeFromCode('193400000X') = %v, %v", g, ok)
	}
	_, ok = GroupTaxonomyCodeFromCode("invalid")
	if ok {
		t.Error("expected false for invalid code")
	}
}

func TestOtherProviderIdentifierIssuerCodeFromCode(t *testing.T) {
	o, ok := OtherProviderIdentifierIssuerCodeFromCode("01")
	if !ok || o != IssuerOther {
		t.Errorf("got %v, %v", o, ok)
	}
	o, ok = OtherProviderIdentifierIssuerCodeFromCode("05")
	if !ok || o != IssuerMedicaid {
		t.Errorf("got %v, %v", o, ok)
	}
}

func TestCountryCode(t *testing.T) {
	cc := CountryCodeFromCode("us")
	if cc.AsCode() != "US" {
		t.Errorf("CountryCodeFromCode('us') = %q, want 'US'", cc.AsCode())
	}
}

func TestProviderNameFullName(t *testing.T) {
	dr := PrefixDr
	jr := SuffixJr
	name := ProviderName{
		Prefix:     &dr,
		First:      "John",
		Middle:     "A",
		Last:       "Smith",
		Suffix:     &jr,
		Credential: "MD",
	}
	got := name.FullName()
	want := "Dr. John A Smith Jr. (MD)"
	if got != want {
		t.Errorf("FullName() = %q, want %q", got, want)
	}
}

func TestAddressFormatSingleLine(t *testing.T) {
	ca := StateCA
	addr := Address{
		Line1:      "123 Main St",
		City:       "Los Angeles",
		State:      &ca,
		PostalCode: "90001",
	}
	got := addr.FormatSingleLine()
	want := "123 Main St, Los Angeles, CA, 90001"
	if got != want {
		t.Errorf("FormatSingleLine() = %q, want %q", got, want)
	}
}

func TestAddressIsEmpty(t *testing.T) {
	empty := Address{}
	if !empty.IsEmpty() {
		t.Error("empty address should be empty")
	}
	nonEmpty := Address{Line1: "123 Main St"}
	if nonEmpty.IsEmpty() {
		t.Error("non-empty address should not be empty")
	}
}

func TestNppesRecordDisplayName(t *testing.T) {
	ind := EntityTypeIndividual
	org := EntityTypeOrganization

	// Individual
	rec := NppesRecord{
		NPI:          "1234567890",
		EntityType:   &ind,
		ProviderName: ProviderName{First: "Jane", Last: "Doe"},
	}
	if got := rec.DisplayName(); got != "Jane Doe" {
		t.Errorf("DisplayName() = %q, want 'Jane Doe'", got)
	}

	// Organization
	rec2 := NppesRecord{
		NPI:              "1234567890",
		EntityType:       &org,
		OrganizationName: OrganizationName{LegalBusinessName: "Acme Health"},
	}
	if got := rec2.DisplayName(); got != "Acme Health" {
		t.Errorf("DisplayName() = %q, want 'Acme Health'", got)
	}

	// Unknown
	rec3 := NppesRecord{NPI: "1234567890"}
	if got := rec3.DisplayName(); got != "Unknown" {
		t.Errorf("DisplayName() = %q, want 'Unknown'", got)
	}
}

func TestNppesRecordIsActive(t *testing.T) {
	rec := NppesRecord{NPI: "1234567890"}
	if !rec.IsActive() {
		t.Error("record without deactivation date should be active")
	}
}

func TestNppesRecordPrimaryTaxonomy(t *testing.T) {
	rec := NppesRecord{
		NPI: "1234567890",
		TaxonomyCodes: []TaxonomyCode{
			{Code: "111111", IsPrimary: false},
			{Code: "222222", IsPrimary: true},
			{Code: "333333", IsPrimary: false},
		},
	}
	pt := rec.PrimaryTaxonomy()
	if pt == nil || pt.Code != "222222" {
		t.Errorf("PrimaryTaxonomy() = %v, want code '222222'", pt)
	}

	rec2 := NppesRecord{NPI: "1234567890", TaxonomyCodes: []TaxonomyCode{}}
	if rec2.PrimaryTaxonomy() != nil {
		t.Error("expected nil for no taxonomy codes")
	}
}

func TestOptionDisplay(t *testing.T) {
	ind := EntityTypeIndividual
	if got := OptionDisplay(&ind); got != "Individual" {
		t.Errorf("OptionDisplay(Individual) = %q", got)
	}
	if got := OptionDisplay(nil); got != "" {
		t.Errorf("OptionDisplay(nil) = %q", got)
	}
}
