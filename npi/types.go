package npi

import (
	"strings"
	"time"
)

const (
	MaxTaxonomyCodes    = 15
	MaxOtherIdentifiers = 50
)

// NPI represents a 10-digit National Provider Identifier.
type NPI string

// ValidateNPI checks if the string is a valid NPI (exactly 10 ASCII digits).
func ValidateNPI(npi string) error {
	if len(npi) != 10 {
		return ErrInvalidNPI(npi)
	}
	for _, c := range npi {
		if c < '0' || c > '9' {
			return ErrInvalidNPI(npi)
		}
	}
	return nil
}

// NewNPI creates a validated NPI.
func NewNPI(npi string) (NPI, error) {
	if err := ValidateNPI(npi); err != nil {
		return "", err
	}
	return NPI(npi), nil
}

// String returns the NPI as a string.
func (n NPI) String() string { return string(n) }

// EntityType represents the provider entity type.
type EntityType string

const (
	EntityTypeIndividual   EntityType = "Individual"
	EntityTypeOrganization EntityType = "Organization"
)

// EntityTypeFromCode parses an entity type from its code.
func EntityTypeFromCode(code string) (EntityType, error) {
	switch code {
	case "1":
		return EntityTypeIndividual, nil
	case "2":
		return EntityTypeOrganization, nil
	default:
		return "", ErrInvalidEntityType(code)
	}
}

// ToCode returns the numeric code for the entity type.
func (e EntityType) ToCode() string {
	switch e {
	case EntityTypeIndividual:
		return "1"
	case EntityTypeOrganization:
		return "2"
	default:
		return ""
	}
}

// StateCode represents a US state or territory code.
type StateCode string

const (
	StateAK StateCode = "AK"
	StateAL StateCode = "AL"
	StateAR StateCode = "AR"
	StateAS StateCode = "AS"
	StateAZ StateCode = "AZ"
	StateCA StateCode = "CA"
	StateCO StateCode = "CO"
	StateCT StateCode = "CT"
	StateDC StateCode = "DC"
	StateDE StateCode = "DE"
	StateFL StateCode = "FL"
	StateFM StateCode = "FM"
	StateGA StateCode = "GA"
	StateGU StateCode = "GU"
	StateHI StateCode = "HI"
	StateIA StateCode = "IA"
	StateID StateCode = "ID"
	StateIL StateCode = "IL"
	StateIN StateCode = "IN"
	StateKS StateCode = "KS"
	StateKY StateCode = "KY"
	StateLA StateCode = "LA"
	StateMA StateCode = "MA"
	StateMD StateCode = "MD"
	StateME StateCode = "ME"
	StateMH StateCode = "MH"
	StateMI StateCode = "MI"
	StateMN StateCode = "MN"
	StateMO StateCode = "MO"
	StateMP StateCode = "MP"
	StateMS StateCode = "MS"
	StateMT StateCode = "MT"
	StateNC StateCode = "NC"
	StateND StateCode = "ND"
	StateNE StateCode = "NE"
	StateNH StateCode = "NH"
	StateNJ StateCode = "NJ"
	StateNM StateCode = "NM"
	StateNV StateCode = "NV"
	StateNY StateCode = "NY"
	StateOH StateCode = "OH"
	StateOK StateCode = "OK"
	StateOR StateCode = "OR"
	StatePA StateCode = "PA"
	StatePR StateCode = "PR"
	StatePW StateCode = "PW"
	StateRI StateCode = "RI"
	StateSC StateCode = "SC"
	StateSD StateCode = "SD"
	StateTN StateCode = "TN"
	StateTX StateCode = "TX"
	StateUT StateCode = "UT"
	StateVA StateCode = "VA"
	StateVI StateCode = "VI"
	StateVT StateCode = "VT"
	StateWA StateCode = "WA"
	StateWI StateCode = "WI"
	StateWV StateCode = "WV"
	StateWY StateCode = "WY"
	StateZZ StateCode = "ZZ"
)

var validStateCodes = map[string]StateCode{
	"AK": StateAK, "AL": StateAL, "AR": StateAR, "AS": StateAS, "AZ": StateAZ,
	"CA": StateCA, "CO": StateCO, "CT": StateCT, "DC": StateDC, "DE": StateDE,
	"FL": StateFL, "FM": StateFM, "GA": StateGA, "GU": StateGU, "HI": StateHI,
	"IA": StateIA, "ID": StateID, "IL": StateIL, "IN": StateIN, "KS": StateKS,
	"KY": StateKY, "LA": StateLA, "MA": StateMA, "MD": StateMD, "ME": StateME,
	"MH": StateMH, "MI": StateMI, "MN": StateMN, "MO": StateMO, "MP": StateMP,
	"MS": StateMS, "MT": StateMT, "NC": StateNC, "ND": StateND, "NE": StateNE,
	"NH": StateNH, "NJ": StateNJ, "NM": StateNM, "NV": StateNV, "NY": StateNY,
	"OH": StateOH, "OK": StateOK, "OR": StateOR, "PA": StatePA, "PR": StatePR,
	"PW": StatePW, "RI": StateRI, "SC": StateSC, "SD": StateSD, "TN": StateTN,
	"TX": StateTX, "UT": StateUT, "VA": StateVA, "VI": StateVI, "VT": StateVT,
	"WA": StateWA, "WI": StateWI, "WV": StateWV, "WY": StateWY, "ZZ": StateZZ,
}

// StateCodeFromCode parses a state code (case-insensitive).
func StateCodeFromCode(code string) (StateCode, bool) {
	sc, ok := validStateCodes[strings.ToUpper(code)]
	return sc, ok
}

// AsCode returns the uppercase 2-letter state code.
func (s StateCode) AsCode() string { return string(s) }

// SexCode represents the provider sex code.
type SexCode string

const (
	SexMale        SexCode = "M"
	SexFemale      SexCode = "F"
	SexUndisclosed SexCode = "U"
)

// SexCodeFromCode parses a sex code.
func SexCodeFromCode(code string) (SexCode, bool) {
	switch code {
	case "M":
		return SexMale, true
	case "F":
		return SexFemale, true
	case "U", "X":
		return SexUndisclosed, true
	default:
		return "", false
	}
}

// AsCode returns the canonical sex code (prefers "U" for Undisclosed).
func (s SexCode) AsCode() string { return string(s) }

// DeactivationReasonCode represents the reason for NPI deactivation.
type DeactivationReasonCode string

const (
	DeactivationDeath       DeactivationReasonCode = "Death"
	DeactivationDisbandment DeactivationReasonCode = "Disbandment"
	DeactivationFraud       DeactivationReasonCode = "Fraud"
	DeactivationOther       DeactivationReasonCode = "Other"
	DeactivationUndisclosed DeactivationReasonCode = "Undisclosed"
)

// DeactivationReasonCodeFromCode parses a deactivation reason code (case-insensitive).
func DeactivationReasonCodeFromCode(code string) (DeactivationReasonCode, bool) {
	switch strings.ToLower(code) {
	case "death":
		return DeactivationDeath, true
	case "disbandment":
		return DeactivationDisbandment, true
	case "fraud":
		return DeactivationFraud, true
	case "other":
		return DeactivationOther, true
	case "u", "undisclosed", "x":
		return DeactivationUndisclosed, true
	default:
		return "", false
	}
}

// AsCode returns the canonical deactivation reason code.
func (d DeactivationReasonCode) AsCode() string { return string(d) }

// OtherProviderNameTypeCode represents the type of other provider name.
type OtherProviderNameTypeCode string

const (
	NameTypeFormerName              OtherProviderNameTypeCode = "1"
	NameTypeProfessionalName        OtherProviderNameTypeCode = "2"
	NameTypeDoingBusinessAs         OtherProviderNameTypeCode = "3"
	NameTypeFormerLegalBusinessName OtherProviderNameTypeCode = "4"
	NameTypeOtherName               OtherProviderNameTypeCode = "5"
)

// OtherProviderNameTypeCodeFromCode parses a name type code.
func OtherProviderNameTypeCodeFromCode(code string) (OtherProviderNameTypeCode, bool) {
	switch code {
	case "1":
		return NameTypeFormerName, true
	case "2":
		return NameTypeProfessionalName, true
	case "3":
		return NameTypeDoingBusinessAs, true
	case "4":
		return NameTypeFormerLegalBusinessName, true
	case "5":
		return NameTypeOtherName, true
	default:
		return "", false
	}
}

// AsCode returns the numeric code.
func (o OtherProviderNameTypeCode) AsCode() string { return string(o) }

// NamePrefixCode represents a name prefix.
type NamePrefixCode string

const (
	PrefixMs   NamePrefixCode = "Ms."
	PrefixMr   NamePrefixCode = "Mr."
	PrefixMiss NamePrefixCode = "Miss"
	PrefixMrs  NamePrefixCode = "Mrs."
	PrefixDr   NamePrefixCode = "Dr."
	PrefixProf NamePrefixCode = "Prof."
)

// NamePrefixCodeFromCode parses a name prefix code.
func NamePrefixCodeFromCode(code string) (NamePrefixCode, bool) {
	switch code {
	case "Ms.":
		return PrefixMs, true
	case "Mr.":
		return PrefixMr, true
	case "Miss":
		return PrefixMiss, true
	case "Mrs.":
		return PrefixMrs, true
	case "Dr.":
		return PrefixDr, true
	case "Prof.":
		return PrefixProf, true
	default:
		return "", false
	}
}

// AsCode returns the prefix string.
func (n NamePrefixCode) AsCode() string { return string(n) }

// NameSuffixCode represents a name suffix.
type NameSuffixCode string

const (
	SuffixJr   NameSuffixCode = "Jr."
	SuffixSr   NameSuffixCode = "Sr."
	SuffixI    NameSuffixCode = "I"
	SuffixII   NameSuffixCode = "II"
	SuffixIII  NameSuffixCode = "III"
	SuffixIV   NameSuffixCode = "IV"
	SuffixV    NameSuffixCode = "V"
	SuffixVI   NameSuffixCode = "VI"
	SuffixVII  NameSuffixCode = "VII"
	SuffixVIII NameSuffixCode = "VIII"
	SuffixIX   NameSuffixCode = "IX"
	SuffixX    NameSuffixCode = "X"
)

// NameSuffixCodeFromCode parses a name suffix code.
func NameSuffixCodeFromCode(code string) (NameSuffixCode, bool) {
	switch code {
	case "Jr.":
		return SuffixJr, true
	case "Sr.":
		return SuffixSr, true
	case "I":
		return SuffixI, true
	case "II":
		return SuffixII, true
	case "III":
		return SuffixIII, true
	case "IV":
		return SuffixIV, true
	case "V":
		return SuffixV, true
	case "VI":
		return SuffixVI, true
	case "VII":
		return SuffixVII, true
	case "VIII":
		return SuffixVIII, true
	case "IX":
		return SuffixIX, true
	case "X":
		return SuffixX, true
	default:
		return "", false
	}
}

// AsCode returns the suffix string.
func (n NameSuffixCode) AsCode() string { return string(n) }

// SoleProprietorCode represents the sole proprietor flag.
type SoleProprietorCode string

const (
	SoleProprietorNotAnswered SoleProprietorCode = "X"
	SoleProprietorYes         SoleProprietorCode = "Y"
	SoleProprietorNo          SoleProprietorCode = "N"
)

// SoleProprietorCodeFromCode parses a sole proprietor code.
func SoleProprietorCodeFromCode(code string) (SoleProprietorCode, bool) {
	switch code {
	case "X":
		return SoleProprietorNotAnswered, true
	case "Y":
		return SoleProprietorYes, true
	case "N":
		return SoleProprietorNo, true
	default:
		return "", false
	}
}

// AsCode returns the code string.
func (s SoleProprietorCode) AsCode() string { return string(s) }

// SubpartCode represents the organization subpart flag.
type SubpartCode string

const (
	SubpartNotAnswered SubpartCode = "X"
	SubpartYes         SubpartCode = "Y"
	SubpartNo          SubpartCode = "N"
)

// SubpartCodeFromCode parses a subpart code.
func SubpartCodeFromCode(code string) (SubpartCode, bool) {
	switch code {
	case "X":
		return SubpartNotAnswered, true
	case "Y":
		return SubpartYes, true
	case "N":
		return SubpartNo, true
	default:
		return "", false
	}
}

// AsCode returns the code string.
func (s SubpartCode) AsCode() string { return string(s) }

// PrimaryTaxonomySwitch represents the primary taxonomy switch.
type PrimaryTaxonomySwitch string

const (
	PrimaryTaxonomyNotAnswered PrimaryTaxonomySwitch = "X"
	PrimaryTaxonomyYes         PrimaryTaxonomySwitch = "Y"
	PrimaryTaxonomyNo          PrimaryTaxonomySwitch = "N"
)

// PrimaryTaxonomySwitchFromCode parses a primary taxonomy switch code.
func PrimaryTaxonomySwitchFromCode(code string) (PrimaryTaxonomySwitch, bool) {
	switch code {
	case "X":
		return PrimaryTaxonomyNotAnswered, true
	case "Y":
		return PrimaryTaxonomyYes, true
	case "N":
		return PrimaryTaxonomyNo, true
	default:
		return "", false
	}
}

// AsCode returns the code string.
func (p PrimaryTaxonomySwitch) AsCode() string { return string(p) }

// GroupTaxonomyCode represents a group taxonomy code.
type GroupTaxonomyCode string

const (
	GroupTaxonomyMultiSpecialty  GroupTaxonomyCode = "193200000X"
	GroupTaxonomySingleSpecialty GroupTaxonomyCode = "193400000X"
)

// GroupTaxonomyCodeFromCode parses a group taxonomy code.
func GroupTaxonomyCodeFromCode(code string) (GroupTaxonomyCode, bool) {
	switch code {
	case "193200000X":
		return GroupTaxonomyMultiSpecialty, true
	case "193400000X":
		return GroupTaxonomySingleSpecialty, true
	default:
		return "", false
	}
}

// AsCode returns the code string.
func (g GroupTaxonomyCode) AsCode() string { return string(g) }

// OtherProviderIdentifierIssuerCode represents an identifier issuer code.
type OtherProviderIdentifierIssuerCode string

const (
	IssuerOther    OtherProviderIdentifierIssuerCode = "01"
	IssuerMedicaid OtherProviderIdentifierIssuerCode = "05"
)

// OtherProviderIdentifierIssuerCodeFromCode parses an issuer code.
func OtherProviderIdentifierIssuerCodeFromCode(code string) (OtherProviderIdentifierIssuerCode, bool) {
	switch code {
	case "01":
		return IssuerOther, true
	case "05":
		return IssuerMedicaid, true
	default:
		return "", false
	}
}

// AsCode returns the code string.
func (o OtherProviderIdentifierIssuerCode) AsCode() string { return string(o) }

// CountryCode represents an ISO 3166-1 alpha-2 country code.
type CountryCode string

// CountryCodeFromCode creates a CountryCode (uppercase normalization).
func CountryCodeFromCode(code string) CountryCode {
	return CountryCode(strings.ToUpper(code))
}

// AsCode returns the country code string.
func (c CountryCode) AsCode() string { return string(c) }

// Address represents a mailing or practice location address.
type Address struct {
	Line1      string       `json:"line_1,omitempty"`
	Line2      string       `json:"line_2,omitempty"`
	City       string       `json:"city,omitempty"`
	PostalCode string       `json:"postal_code,omitempty"`
	Telephone  string       `json:"telephone,omitempty"`
	Fax        string       `json:"fax,omitempty"`
	State      *StateCode   `json:"state,omitempty"`
	Country    *CountryCode `json:"country,omitempty"`
}

// IsEmpty returns true if the address has no data.
func (a *Address) IsEmpty() bool {
	return a.Line1 == "" && a.Line2 == "" && a.City == "" && a.State == nil && a.PostalCode == ""
}

// FormatSingleLine formats the address as a single-line string.
func (a *Address) FormatSingleLine() string {
	var parts []string
	if a.Line1 != "" {
		parts = append(parts, a.Line1)
	}
	if a.City != "" {
		parts = append(parts, a.City)
	}
	if a.State != nil {
		parts = append(parts, a.State.AsCode())
	}
	if a.PostalCode != "" {
		parts = append(parts, a.PostalCode)
	}
	return strings.Join(parts, ", ")
}

// ProviderName represents an individual provider's name.
type ProviderName struct {
	Prefix     *NamePrefixCode `json:"prefix,omitempty"`
	First      string          `json:"first,omitempty"`
	Middle     string          `json:"middle,omitempty"`
	Last       string          `json:"last,omitempty"`
	Suffix     *NameSuffixCode `json:"suffix,omitempty"`
	Credential string          `json:"credential,omitempty"`
}

// FullName formats the complete name.
func (p *ProviderName) FullName() string {
	var parts []string
	if p.Prefix != nil {
		parts = append(parts, p.Prefix.AsCode())
	}
	if p.First != "" {
		parts = append(parts, p.First)
	}
	if p.Middle != "" {
		parts = append(parts, p.Middle)
	}
	if p.Last != "" {
		parts = append(parts, p.Last)
	}
	if p.Suffix != nil {
		parts = append(parts, p.Suffix.AsCode())
	}
	if p.Credential != "" {
		parts = append(parts, "("+p.Credential+")")
	}
	return strings.Join(parts, " ")
}

// OrganizationName represents an organization's name.
type OrganizationName struct {
	LegalBusinessName string                     `json:"legal_business_name,omitempty"`
	OtherName         string                     `json:"other_name,omitempty"`
	OtherNameType     *OtherProviderNameTypeCode `json:"other_name_type,omitempty"`
}

// AuthorizedOfficial represents the authorized official for an organization.
type AuthorizedOfficial struct {
	Prefix     *NamePrefixCode `json:"prefix,omitempty"`
	FirstName  string          `json:"first_name,omitempty"`
	MiddleName string          `json:"middle_name,omitempty"`
	LastName   string          `json:"last_name,omitempty"`
	Suffix     *NameSuffixCode `json:"suffix,omitempty"`
	Credential string          `json:"credential,omitempty"`
	Title      string          `json:"title,omitempty"`
	Telephone  string          `json:"telephone,omitempty"`
}

// FullName formats the authorized official's complete name.
func (a *AuthorizedOfficial) FullName() string {
	var parts []string
	if a.Prefix != nil {
		parts = append(parts, a.Prefix.AsCode())
	}
	if a.FirstName != "" {
		parts = append(parts, a.FirstName)
	}
	if a.MiddleName != "" {
		parts = append(parts, a.MiddleName)
	}
	if a.LastName != "" {
		parts = append(parts, a.LastName)
	}
	if a.Suffix != nil {
		parts = append(parts, a.Suffix.AsCode())
	}
	if a.Credential != "" {
		parts = append(parts, "("+a.Credential+")")
	}
	return strings.Join(parts, " ")
}

// TaxonomyCode represents a healthcare provider taxonomy code.
type TaxonomyCode struct {
	Code              string                 `json:"code"`
	LicenseNumber     string                 `json:"license_number,omitempty"`
	LicenseState      string                 `json:"license_state,omitempty"`
	IsPrimary         bool                   `json:"is_primary"`
	TaxonomyGroup     string                 `json:"taxonomy_group,omitempty"`
	GroupTaxonomyCode *GroupTaxonomyCode     `json:"group_taxonomy_code,omitempty"`
	PrimarySwitch     *PrimaryTaxonomySwitch `json:"primary_switch,omitempty"`
}

// OtherIdentifier represents an other provider identifier.
type OtherIdentifier struct {
	Identifier string                             `json:"identifier"`
	TypeCode   string                             `json:"type_code,omitempty"`
	Issuer     *OtherProviderIdentifierIssuerCode `json:"issuer,omitempty"`
	State      *StateCode                         `json:"state,omitempty"`
}

// NppesRecord represents the main NPPES provider record.
type NppesRecord struct {
	NPI                   NPI                        `json:"npi"`
	EntityType            *EntityType                `json:"entity_type,omitempty"`
	ReplacementNPI        *NPI                       `json:"replacement_npi,omitempty"`
	EIN                   string                     `json:"ein,omitempty"`
	ProviderName          ProviderName               `json:"provider_name"`
	ProviderOtherName     ProviderName               `json:"provider_other_name"`
	ProviderOtherNameType *OtherProviderNameTypeCode `json:"provider_other_name_type,omitempty"`
	OrganizationName      OrganizationName           `json:"organization_name"`
	MailingAddress        Address                    `json:"mailing_address"`
	PracticeAddress       Address                    `json:"practice_address"`
	EnumerationDate       *time.Time                 `json:"enumeration_date,omitempty"`
	LastUpdateDate        *time.Time                 `json:"last_update_date,omitempty"`
	DeactivationDate      *time.Time                 `json:"deactivation_date,omitempty"`
	ReactivationDate      *time.Time                 `json:"reactivation_date,omitempty"`
	CertificationDate     *time.Time                 `json:"certification_date,omitempty"`
	DeactivationReason    *DeactivationReasonCode    `json:"deactivation_reason,omitempty"`
	ProviderGender        *SexCode                   `json:"provider_gender,omitempty"`
	AuthorizedOfficial    *AuthorizedOfficial        `json:"authorized_official,omitempty"`
	TaxonomyCodes         []TaxonomyCode             `json:"taxonomy_codes"`
	OtherIdentifiers      []OtherIdentifier          `json:"other_identifiers"`
	SoleProprietor        *SoleProprietorCode        `json:"sole_proprietor,omitempty"`
	OrganizationSubpart   *SubpartCode               `json:"organization_subpart,omitempty"`
	ParentOrganizationLBN string                     `json:"parent_organization_lbn,omitempty"`
	ParentOrganizationTIN string                     `json:"parent_organization_tin,omitempty"`
}

// PrimaryTaxonomy returns the primary taxonomy code, or nil if none.
func (r *NppesRecord) PrimaryTaxonomy() *TaxonomyCode {
	for i := range r.TaxonomyCodes {
		if r.TaxonomyCodes[i].IsPrimary {
			return &r.TaxonomyCodes[i]
		}
	}
	return nil
}

// IsActive returns true if the provider is not deactivated.
func (r *NppesRecord) IsActive() bool {
	return r.DeactivationDate == nil
}

// DisplayName returns the provider's primary name based on entity type.
func (r *NppesRecord) DisplayName() string {
	if r.EntityType == nil {
		return "Unknown"
	}
	switch *r.EntityType {
	case EntityTypeIndividual:
		name := strings.TrimSpace(r.ProviderName.First + " " + r.ProviderName.Last)
		if name == "" {
			return "Unknown"
		}
		return name
	case EntityTypeOrganization:
		if r.OrganizationName.LegalBusinessName != "" {
			return r.OrganizationName.LegalBusinessName
		}
		return "Unknown Organization"
	default:
		return "Unknown"
	}
}

// FullDisplayName returns the full formatted name (includes credentials/titles).
func (r *NppesRecord) FullDisplayName() string {
	if r.EntityType == nil {
		return "Unknown"
	}
	switch *r.EntityType {
	case EntityTypeIndividual:
		return r.ProviderName.FullName()
	case EntityTypeOrganization:
		if r.OrganizationName.LegalBusinessName != "" {
			return r.OrganizationName.LegalBusinessName
		}
		return "Unknown Organization"
	default:
		return "Unknown"
	}
}

// OptionDisplay returns the string representation of an entity type, or empty string if nil.
func OptionDisplay(et *EntityType) string {
	if et == nil {
		return ""
	}
	return string(*et)
}

// TaxonomyReference represents a healthcare taxonomy reference entry.
type TaxonomyReference struct {
	Code           string `json:"code"`
	Grouping       string `json:"grouping,omitempty"`
	Classification string `json:"classification,omitempty"`
	Specialization string `json:"specialization,omitempty"`
	Definition     string `json:"definition,omitempty"`
	Notes          string `json:"notes,omitempty"`
	DisplayName    string `json:"display_name,omitempty"`
	Section        string `json:"section,omitempty"`
}

// OtherNameRecord represents an other name reference record.
type OtherNameRecord struct {
	NPI                                   NPI    `json:"npi"`
	ProviderOtherOrganizationName         string `json:"provider_other_organization_name"`
	ProviderOtherOrganizationNameTypeCode string `json:"provider_other_organization_name_type_code,omitempty"`
}

// PracticeLocationRecord represents a practice location reference record.
type PracticeLocationRecord struct {
	NPI                NPI     `json:"npi"`
	Address            Address `json:"address"`
	TelephoneExtension string  `json:"telephone_extension,omitempty"`
}

// EndpointRecord represents an endpoint reference record.
type EndpointRecord struct {
	NPI                          NPI      `json:"npi"`
	EndpointType                 string   `json:"endpoint_type,omitempty"`
	EndpointTypeDescription      string   `json:"endpoint_type_description,omitempty"`
	Endpoint                     string   `json:"endpoint,omitempty"`
	Affiliation                  *bool    `json:"affiliation,omitempty"`
	EndpointDescription          string   `json:"endpoint_description,omitempty"`
	AffiliationLegalBusinessName string   `json:"affiliation_legal_business_name,omitempty"`
	UseCode                      string   `json:"use_code,omitempty"`
	UseDescription               string   `json:"use_description,omitempty"`
	OtherUseDescription          string   `json:"other_use_description,omitempty"`
	ContentType                  string   `json:"content_type,omitempty"`
	ContentDescription           string   `json:"content_description,omitempty"`
	OtherContentDescription      string   `json:"other_content_description,omitempty"`
	AffiliationAddress           *Address `json:"affiliation_address,omitempty"`
}
