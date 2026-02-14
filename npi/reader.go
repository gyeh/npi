package npi

import (
	"encoding/csv"
	"io"
	"os"
	"strings"
	"time"
)

const dateFormat = "01/02/2006" // Go's MM/DD/YYYY

// Loader is the interface for loading NPPES CSV data files.
type Loader interface {
	LoadMainData(path string) ([]NppesRecord, error)
	LoadTaxonomyData(path string) ([]TaxonomyReference, error)
	LoadOtherNameData(path string) ([]OtherNameRecord, error)
	LoadPracticeLocationData(path string) ([]PracticeLocationRecord, error)
	LoadEndpointData(path string) ([]EndpointRecord, error)
}

var _ Loader = (*NppesReader)(nil)

// NppesReader reads and parses NPPES CSV files.
type NppesReader struct {
	ValidateHeaders    bool
	SkipInvalidRecords bool
	Logger             func(format string, args ...interface{})
}

// NewNppesReader creates a reader with default settings.
func NewNppesReader() *NppesReader {
	return &NppesReader{
		ValidateHeaders:    true,
		SkipInvalidRecords: false,
	}
}

func (r *NppesReader) logf(format string, args ...interface{}) {
	if r.Logger != nil {
		r.Logger(format, args...)
	}
}

func getField(record []string, index int) string {
	if index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func getOptionalField(record []string, index int) string {
	v := getField(record, index)
	if v == "" {
		return ""
	}
	return v
}

func parseDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		return nil, ErrDateParse(s, "MM/DD/YYYY")
	}
	return &t, nil
}

func parseStateCode(s string) *StateCode {
	if s == "" {
		return nil
	}
	sc, ok := StateCodeFromCode(s)
	if !ok {
		return nil
	}
	return &sc
}

func parseCountryCode(s string) *CountryCode {
	if s == "" {
		return nil
	}
	cc := CountryCodeFromCode(s)
	return &cc
}

func parseNamePrefix(s string) *NamePrefixCode {
	if s == "" {
		return nil
	}
	p, ok := NamePrefixCodeFromCode(s)
	if !ok {
		return nil
	}
	return &p
}

func parseNameSuffix(s string) *NameSuffixCode {
	if s == "" {
		return nil
	}
	p, ok := NameSuffixCodeFromCode(s)
	if !ok {
		return nil
	}
	return &p
}

// LoadMainData loads the main NPPES provider CSV file.
func (r *NppesReader) LoadMainData(path string) ([]NppesRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, ErrFileNotFound(path)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	// Read and validate headers
	headerRow, err := reader.Read()
	if err != nil {
		return nil, ErrCsvParse("failed to read CSV header", 1)
	}

	if r.ValidateHeaders {
		expected := MainSchemaColumnNames()
		if err := ValidateHeaders(expected, headerRow); err != nil {
			return nil, err
		}
	}

	// Estimate capacity
	fi, _ := f.Stat()
	estimatedRecords := 1000
	if fi != nil && fi.Size() > 0 {
		estimatedRecords = int(fi.Size() / 2000)
	}

	records := make([]NppesRecord, 0, estimatedRecords)
	lineNum := 1
	invalidCount := 0
	startTime := time.Now()

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		lineNum++
		if err != nil {
			if r.SkipInvalidRecords {
				invalidCount++
				continue
			}
			return nil, ErrCsvParse(err.Error(), lineNum)
		}

		rec, err := r.parseMainRecord(row, lineNum)
		if err != nil {
			if r.SkipInvalidRecords {
				invalidCount++
				if invalidCount <= 10 {
					r.logf("Warning: Skipping invalid record %d: %v\n", lineNum, err)
				}
				continue
			}
			return nil, err
		}
		records = append(records, rec)
	}

	elapsed := time.Since(startTime)
	rate := float64(len(records)) / elapsed.Seconds()
	r.logf("Successfully loaded %d NPPES provider records in %.2fs (%.0f records/sec)\n",
		len(records), elapsed.Seconds(), rate)
	if invalidCount > 0 {
		r.logf("Skipped %d invalid records\n", invalidCount)
	}

	return records, nil
}

func (r *NppesReader) parseMainRecord(row []string, lineNum int) (NppesRecord, error) {
	npiStr := getField(row, 0)
	if npiStr == "" {
		return NppesRecord{}, ErrDataValidation("Missing required field: NPI", "NPI")
	}
	npi, err := NewNPI(npiStr)
	if err != nil {
		return NppesRecord{}, err
	}

	var entityType *EntityType
	if code := getOptionalField(row, 1); code != "" {
		et, err := EntityTypeFromCode(code)
		if err == nil {
			entityType = &et
		}
	}

	var replacementNPI *NPI
	if s := getOptionalField(row, 2); s != "" {
		n, err := NewNPI(s)
		if err == nil {
			replacementNPI = &n
		}
	}

	providerName := ProviderName{
		Prefix:     parseNamePrefix(getOptionalField(row, 9)),
		First:      getOptionalField(row, 7),
		Middle:     getOptionalField(row, 8),
		Last:       getOptionalField(row, 6),
		Suffix:     parseNameSuffix(getOptionalField(row, 10)),
		Credential: getOptionalField(row, 11),
	}

	providerOtherName := ProviderName{
		Prefix:     parseNamePrefix(getOptionalField(row, 17)),
		First:      getOptionalField(row, 15),
		Middle:     getOptionalField(row, 16),
		Last:       getOptionalField(row, 14),
		Suffix:     parseNameSuffix(getOptionalField(row, 18)),
		Credential: getOptionalField(row, 19),
	}

	var orgOtherNameType *OtherProviderNameTypeCode
	if s := getOptionalField(row, 13); s != "" {
		t, ok := OtherProviderNameTypeCodeFromCode(s)
		if ok {
			orgOtherNameType = &t
		}
	}

	orgName := OrganizationName{
		LegalBusinessName: getOptionalField(row, 4),
		OtherName:         getOptionalField(row, 12),
		OtherNameType:     orgOtherNameType,
	}

	mailingAddress := Address{
		Line1:      getOptionalField(row, 20),
		Line2:      getOptionalField(row, 21),
		City:       getOptionalField(row, 22),
		PostalCode: getOptionalField(row, 24),
		Telephone:  getOptionalField(row, 26),
		Fax:        getOptionalField(row, 27),
		State:      parseStateCode(getOptionalField(row, 23)),
		Country:    parseCountryCode(getOptionalField(row, 25)),
	}

	practiceAddress := Address{
		Line1:      getOptionalField(row, 28),
		Line2:      getOptionalField(row, 29),
		City:       getOptionalField(row, 30),
		PostalCode: getOptionalField(row, 32),
		Telephone:  getOptionalField(row, 34),
		Fax:        getOptionalField(row, 35),
		State:      parseStateCode(getOptionalField(row, 31)),
		Country:    parseCountryCode(getOptionalField(row, 33)),
	}

	enumerationDate, err := parseDate(getOptionalField(row, 36))
	if err != nil {
		return NppesRecord{}, err
	}
	lastUpdateDate, err := parseDate(getOptionalField(row, 37))
	if err != nil {
		return NppesRecord{}, err
	}
	deactivationDate, err := parseDate(getOptionalField(row, 39))
	if err != nil {
		return NppesRecord{}, err
	}
	reactivationDate, err := parseDate(getOptionalField(row, 40))
	if err != nil {
		return NppesRecord{}, err
	}

	// Taxonomy codes (columns 47-106)
	var taxonomyCodes []TaxonomyCode
	for i := 0; i < MaxTaxonomyCodes; i++ {
		baseIdx := 47 + (i * 4)
		code := getOptionalField(row, baseIdx)
		if code == "" {
			continue
		}
		var groupTC *GroupTaxonomyCode
		if s := getOptionalField(row, 307+i); s != "" {
			g, ok := GroupTaxonomyCodeFromCode(s)
			if ok {
				groupTC = &g
			}
		}
		var primarySwitch *PrimaryTaxonomySwitch
		if s := getOptionalField(row, baseIdx+3); s != "" {
			p, ok := PrimaryTaxonomySwitchFromCode(s)
			if ok {
				primarySwitch = &p
			}
		}
		tc := TaxonomyCode{
			Code:              code,
			LicenseNumber:     getOptionalField(row, baseIdx+1),
			LicenseState:      getOptionalField(row, baseIdx+2),
			IsPrimary:         getOptionalField(row, baseIdx+3) == "Y",
			TaxonomyGroup:     getOptionalField(row, 307+i),
			GroupTaxonomyCode: groupTC,
			PrimarySwitch:     primarySwitch,
		}
		taxonomyCodes = append(taxonomyCodes, tc)
	}
	if taxonomyCodes == nil {
		taxonomyCodes = []TaxonomyCode{}
	}

	// Other identifiers (columns 107-306)
	var otherIdentifiers []OtherIdentifier
	for i := 0; i < MaxOtherIdentifiers; i++ {
		baseIdx := 107 + (i * 4)
		identifier := getOptionalField(row, baseIdx)
		if identifier == "" {
			continue
		}
		var issuer *OtherProviderIdentifierIssuerCode
		if s := getOptionalField(row, baseIdx+3); s != "" {
			ic, ok := OtherProviderIdentifierIssuerCodeFromCode(s)
			if ok {
				issuer = &ic
			}
		}
		oi := OtherIdentifier{
			Identifier: identifier,
			TypeCode:   getOptionalField(row, baseIdx+1),
			State:      parseStateCode(getOptionalField(row, baseIdx+2)),
			Issuer:     issuer,
		}
		otherIdentifiers = append(otherIdentifiers, oi)
	}
	if otherIdentifiers == nil {
		otherIdentifiers = []OtherIdentifier{}
	}

	// Authorized official
	var authOfficial *AuthorizedOfficial
	if entityType != nil && *entityType == EntityTypeOrganization {
		authOfficial = &AuthorizedOfficial{
			Prefix:     parseNamePrefix(getOptionalField(row, 308)),
			FirstName:  getOptionalField(row, 43),
			MiddleName: getOptionalField(row, 44),
			LastName:   getOptionalField(row, 42),
			Suffix:     parseNameSuffix(getOptionalField(row, 309)),
			Credential: getOptionalField(row, 310),
			Title:      getOptionalField(row, 45),
			Telephone:  getOptionalField(row, 46),
		}
	}

	// Organization flags
	var soleProprietor *SoleProprietorCode
	if s := getOptionalField(row, 307); s != "" {
		sp, ok := SoleProprietorCodeFromCode(s)
		if ok {
			soleProprietor = &sp
		}
	}
	var orgSubpart *SubpartCode
	if s := getOptionalField(row, 308); s != "" {
		sc, ok := SubpartCodeFromCode(s)
		if ok {
			orgSubpart = &sc
		}
	}

	certificationDate, err := parseDate(getOptionalField(row, 329))
	if err != nil {
		return NppesRecord{}, err
	}

	// Deactivation reason and gender codes
	var deactivationReason *DeactivationReasonCode
	if s := getOptionalField(row, 38); s != "" {
		dr, ok := DeactivationReasonCodeFromCode(s)
		if ok {
			deactivationReason = &dr
		}
	}
	var providerGender *SexCode
	if s := getOptionalField(row, 41); s != "" {
		sc, ok := SexCodeFromCode(s)
		if ok {
			providerGender = &sc
		}
	}

	// Provider other name type code
	var providerOtherNameType *OtherProviderNameTypeCode
	if s := getOptionalField(row, 20); s != "" {
		t, ok := OtherProviderNameTypeCodeFromCode(s)
		if ok {
			providerOtherNameType = &t
		}
	}

	return NppesRecord{
		NPI:                   npi,
		EntityType:            entityType,
		ReplacementNPI:        replacementNPI,
		EIN:                   getOptionalField(row, 3),
		ProviderName:          providerName,
		ProviderOtherName:     providerOtherName,
		ProviderOtherNameType: providerOtherNameType,
		OrganizationName:      orgName,
		MailingAddress:        mailingAddress,
		PracticeAddress:       practiceAddress,
		EnumerationDate:       enumerationDate,
		LastUpdateDate:        lastUpdateDate,
		DeactivationDate:      deactivationDate,
		ReactivationDate:      reactivationDate,
		CertificationDate:     certificationDate,
		DeactivationReason:    deactivationReason,
		ProviderGender:        providerGender,
		AuthorizedOfficial:    authOfficial,
		TaxonomyCodes:         taxonomyCodes,
		OtherIdentifiers:      otherIdentifiers,
		SoleProprietor:        soleProprietor,
		OrganizationSubpart:   orgSubpart,
		ParentOrganizationLBN: getOptionalField(row, 309),
		ParentOrganizationTIN: getOptionalField(row, 310),
	}, nil
}

// LoadTaxonomyData loads taxonomy reference data from a CSV file.
func (r *NppesReader) LoadTaxonomyData(path string) ([]TaxonomyReference, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, ErrFileNotFound(path)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	headerRow, err := reader.Read()
	if err != nil {
		return nil, ErrCsvParse("failed to read CSV header", 1)
	}
	if r.ValidateHeaders {
		if err := ValidateHeaders(TaxonomySchemaColumnNames(), headerRow); err != nil {
			return nil, err
		}
	}

	var records []TaxonomyReference
	startTime := time.Now()
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if r.SkipInvalidRecords {
				continue
			}
			return nil, ErrCsvParse(err.Error(), len(records)+2)
		}
		records = append(records, TaxonomyReference{
			Code:           getOptionalField(row, 0),
			Grouping:       getOptionalField(row, 1),
			Classification: getOptionalField(row, 2),
			Specialization: getOptionalField(row, 3),
			Definition:     getOptionalField(row, 4),
			Notes:          getOptionalField(row, 5),
			DisplayName:    getOptionalField(row, 6),
			Section:        getOptionalField(row, 7),
		})
	}

	elapsed := time.Since(startTime)
	r.logf("Successfully loaded %d taxonomy reference records in %.2fs\n", len(records), elapsed.Seconds())
	return records, nil
}

// LoadOtherNameData loads other name reference data from a CSV file.
func (r *NppesReader) LoadOtherNameData(path string) ([]OtherNameRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, ErrFileNotFound(path)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	headerRow, err := reader.Read()
	if err != nil {
		return nil, ErrCsvParse("failed to read CSV header", 1)
	}
	if r.ValidateHeaders {
		if err := ValidateHeaders(OtherNameSchemaColumnNames(), headerRow); err != nil {
			return nil, err
		}
	}

	var records []OtherNameRecord
	startTime := time.Now()
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if r.SkipInvalidRecords {
				continue
			}
			return nil, ErrCsvParse(err.Error(), len(records)+2)
		}

		npiStr := getOptionalField(row, 0)
		if npiStr == "" {
			if r.SkipInvalidRecords {
				continue
			}
			return nil, ErrDataValidation("Missing NPI in other name record", "NPI")
		}
		npi, err := NewNPI(npiStr)
		if err != nil {
			if r.SkipInvalidRecords {
				continue
			}
			return nil, err
		}

		records = append(records, OtherNameRecord{
			NPI:                                   npi,
			ProviderOtherOrganizationName:         getOptionalField(row, 1),
			ProviderOtherOrganizationNameTypeCode: getOptionalField(row, 2),
		})
	}

	elapsed := time.Since(startTime)
	r.logf("Successfully loaded %d other name records in %.2fs\n", len(records), elapsed.Seconds())
	return records, nil
}

// LoadPracticeLocationData loads practice location reference data from a CSV file.
func (r *NppesReader) LoadPracticeLocationData(path string) ([]PracticeLocationRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, ErrFileNotFound(path)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	headerRow, err := reader.Read()
	if err != nil {
		return nil, ErrCsvParse("failed to read CSV header", 1)
	}
	if r.ValidateHeaders {
		if err := ValidateHeaders(PracticeLocationSchemaColumnNames(), headerRow); err != nil {
			return nil, err
		}
	}

	var records []PracticeLocationRecord
	startTime := time.Now()
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if r.SkipInvalidRecords {
				continue
			}
			return nil, ErrCsvParse(err.Error(), len(records)+2)
		}

		npiStr := getOptionalField(row, 0)
		if npiStr == "" {
			if r.SkipInvalidRecords {
				continue
			}
			return nil, ErrDataValidation("Missing NPI in practice location record", "NPI")
		}
		npi, err := NewNPI(npiStr)
		if err != nil {
			if r.SkipInvalidRecords {
				continue
			}
			return nil, err
		}

		records = append(records, PracticeLocationRecord{
			NPI: npi,
			Address: Address{
				Line1:      getOptionalField(row, 1),
				Line2:      getOptionalField(row, 2),
				City:       getOptionalField(row, 3),
				PostalCode: getOptionalField(row, 5),
				Telephone:  getOptionalField(row, 7),
				Fax:        getOptionalField(row, 9),
				State:      parseStateCode(getOptionalField(row, 4)),
				Country:    parseCountryCode(getOptionalField(row, 6)),
			},
			TelephoneExtension: getOptionalField(row, 8),
		})
	}

	elapsed := time.Since(startTime)
	r.logf("Successfully loaded %d practice location records in %.2fs\n", len(records), elapsed.Seconds())
	return records, nil
}

// LoadEndpointData loads endpoint reference data from a CSV file.
func (r *NppesReader) LoadEndpointData(path string) ([]EndpointRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, ErrFileNotFound(path)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	headerRow, err := reader.Read()
	if err != nil {
		return nil, ErrCsvParse("failed to read CSV header", 1)
	}
	if r.ValidateHeaders {
		if err := ValidateHeaders(EndpointSchemaColumnNames(), headerRow); err != nil {
			return nil, err
		}
	}

	var records []EndpointRecord
	startTime := time.Now()
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if r.SkipInvalidRecords {
				continue
			}
			return nil, ErrCsvParse(err.Error(), len(records)+2)
		}

		npiStr := getOptionalField(row, 0)
		if npiStr == "" {
			if r.SkipInvalidRecords {
				continue
			}
			return nil, ErrDataValidation("Missing NPI in endpoint record", "NPI")
		}
		npi, err := NewNPI(npiStr)
		if err != nil {
			if r.SkipInvalidRecords {
				continue
			}
			return nil, err
		}

		var affiliation *bool
		if s := getOptionalField(row, 4); s != "" {
			b := s == "Y"
			affiliation = &b
		}

		var affiliationAddress *Address
		if getOptionalField(row, 13) != "" || getOptionalField(row, 14) != "" {
			affiliationAddress = &Address{
				Line1:      getOptionalField(row, 13),
				Line2:      getOptionalField(row, 14),
				City:       getOptionalField(row, 15),
				PostalCode: getOptionalField(row, 18),
				State:      parseStateCode(getOptionalField(row, 16)),
				Country:    parseCountryCode(getOptionalField(row, 17)),
			}
		}

		records = append(records, EndpointRecord{
			NPI:                          npi,
			EndpointType:                 getOptionalField(row, 1),
			EndpointTypeDescription:      getOptionalField(row, 2),
			Endpoint:                     getOptionalField(row, 3),
			Affiliation:                  affiliation,
			EndpointDescription:          getOptionalField(row, 5),
			AffiliationLegalBusinessName: getOptionalField(row, 6),
			UseCode:                      getOptionalField(row, 7),
			UseDescription:               getOptionalField(row, 8),
			OtherUseDescription:          getOptionalField(row, 9),
			ContentType:                  getOptionalField(row, 10),
			ContentDescription:           getOptionalField(row, 11),
			OtherContentDescription:      getOptionalField(row, 12),
			AffiliationAddress:           affiliationAddress,
		})
	}

	elapsed := time.Since(startTime)
	r.logf("Successfully loaded %d endpoint records in %.2fs\n", len(records), elapsed.Seconds())
	return records, nil
}
