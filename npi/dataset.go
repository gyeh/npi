package npi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Dataset is the interface for querying NPPES provider data.
type Dataset interface {
	Len() int
	IsEmpty() bool
	GetByNPI(npi NPI) *NppesRecord
	GetByState(state string) []*NppesRecord
	GetByTaxonomy(taxonomyCode string) []*NppesRecord
	GetTaxonomyDescription(code string) *TaxonomyReference
	Query() *QueryBuilder
	Statistics() DatasetStatistics
}

var _ Dataset = (*NppesDataset)(nil)

// NppesDataset is the unified dataset containing all loaded data and indexes.
type NppesDataset struct {
	Providers            []NppesRecord
	TaxonomyMap          map[string]TaxonomyReference
	OtherNamesMap        map[NPI][]OtherNameRecord
	PracticeLocationsMap map[NPI][]PracticeLocationRecord
	EndpointsMap         map[NPI][]EndpointRecord

	npiIndex      map[NPI]int
	stateIndex    map[string][]int
	taxonomyIndex map[string][]int
}

// LoadOption configures LoadStandard behavior.
type LoadOption func(*loadConfig)

type loadConfig struct {
	logger func(format string, args ...interface{})
}

// WithLogger sets a logger callback for LoadStandard progress messages.
func WithLogger(fn func(string, ...interface{})) LoadOption {
	return func(c *loadConfig) {
		c.logger = fn
	}
}

// LoadStandard loads a standard dataset from a directory containing NPPES files.
func LoadStandard(dir string, opts ...LoadOption) (*NppesDataset, error) {
	cfg := &loadConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	logf := func(format string, args ...interface{}) {
		if cfg.logger != nil {
			cfg.logger(format, args...)
		}
	}

	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return nil, ErrCustom(fmt.Sprintf("'%s' is not a directory", dir))
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, ErrCustom(fmt.Sprintf("cannot read directory: %s", dir))
	}

	var mainDataPath, taxonomyPath, otherNamesPath, practiceLocationsPath, endpointsPath string

	for _, entry := range entries {
		name := entry.Name()
		lower := strings.ToLower(name)
		// Skip fileheader metadata files
		if strings.Contains(lower, "_fileheader") {
			continue
		}
		if strings.HasPrefix(name, "npidata_pfile_") && strings.HasSuffix(name, ".csv") {
			mainDataPath = filepath.Join(dir, name)
		} else if strings.HasPrefix(name, "nucc_taxonomy_") && strings.HasSuffix(name, ".csv") {
			taxonomyPath = filepath.Join(dir, name)
		} else if strings.HasPrefix(name, "othername_pfile_") && strings.HasSuffix(name, ".csv") {
			otherNamesPath = filepath.Join(dir, name)
		} else if strings.HasPrefix(name, "pl_pfile_") && strings.HasSuffix(name, ".csv") {
			practiceLocationsPath = filepath.Join(dir, name)
		} else if strings.HasPrefix(name, "endpoint_pfile_") && strings.HasSuffix(name, ".csv") {
			endpointsPath = filepath.Join(dir, name)
		}
	}

	if mainDataPath == "" {
		return nil, ErrCustom("no main NPPES data file (npidata_pfile_*.csv) found in directory")
	}

	reader := NewNppesReader()
	reader.SkipInvalidRecords = true
	reader.Logger = cfg.logger

	logf("Loading main provider data from: %s\n", mainDataPath)
	providers, err := reader.LoadMainData(mainDataPath)
	if err != nil {
		return nil, err
	}

	ds := &NppesDataset{
		Providers: providers,
	}

	if taxonomyPath != "" {
		logf("Loading taxonomy reference from: %s\n", taxonomyPath)
		taxonomies, err := reader.LoadTaxonomyData(taxonomyPath)
		if err != nil {
			return nil, err
		}
		ds.TaxonomyMap = createTaxonomyMap(taxonomies)
	}

	if otherNamesPath != "" {
		logf("Loading other names from: %s\n", otherNamesPath)
		otherNames, err := reader.LoadOtherNameData(otherNamesPath)
		if err != nil {
			return nil, err
		}
		ds.OtherNamesMap = createOtherNamesMap(otherNames)
	}

	if practiceLocationsPath != "" {
		logf("Loading practice locations from: %s\n", practiceLocationsPath)
		locations, err := reader.LoadPracticeLocationData(practiceLocationsPath)
		if err != nil {
			return nil, err
		}
		ds.PracticeLocationsMap = createPracticeLocationsMap(locations)
	}

	if endpointsPath != "" {
		logf("Loading endpoints from: %s\n", endpointsPath)
		endpoints, err := reader.LoadEndpointData(endpointsPath)
		if err != nil {
			return nil, err
		}
		ds.EndpointsMap = createEndpointsMap(endpoints)
	}

	logf("Building indexes...\n")
	ds.BuildIndexes()
	logf("Dataset loaded successfully!\n")

	return ds, nil
}

// Len returns the number of providers.
func (ds *NppesDataset) Len() int { return len(ds.Providers) }

// IsEmpty returns true if there are no providers.
func (ds *NppesDataset) IsEmpty() bool { return len(ds.Providers) == 0 }

// BuildIndexes builds NPI, state, and taxonomy indexes for fast lookups.
func (ds *NppesDataset) BuildIndexes() {
	npiIndex := make(map[NPI]int, len(ds.Providers))
	stateIndex := make(map[string][]int)
	taxonomyIndex := make(map[string][]int)

	for idx, provider := range ds.Providers {
		npiIndex[provider.NPI] = idx

		if provider.MailingAddress.State != nil {
			code := provider.MailingAddress.State.AsCode()
			stateIndex[code] = append(stateIndex[code], idx)
		}

		for _, taxonomy := range provider.TaxonomyCodes {
			taxonomyIndex[taxonomy.Code] = append(taxonomyIndex[taxonomy.Code], idx)
		}
	}

	ds.npiIndex = npiIndex
	ds.stateIndex = stateIndex
	ds.taxonomyIndex = taxonomyIndex
}

// GetByNPI returns a provider by NPI.
func (ds *NppesDataset) GetByNPI(npi NPI) *NppesRecord {
	if ds.npiIndex != nil {
		if idx, ok := ds.npiIndex[npi]; ok {
			return &ds.Providers[idx]
		}
		return nil
	}
	for i := range ds.Providers {
		if ds.Providers[i].NPI == npi {
			return &ds.Providers[i]
		}
	}
	return nil
}

// GetByState returns all providers in a given state.
func (ds *NppesDataset) GetByState(state string) []*NppesRecord {
	sc, ok := StateCodeFromCode(state)
	if !ok {
		return nil
	}
	if ds.stateIndex != nil {
		indices := ds.stateIndex[sc.AsCode()]
		result := make([]*NppesRecord, 0, len(indices))
		for _, idx := range indices {
			result = append(result, &ds.Providers[idx])
		}
		return result
	}
	var result []*NppesRecord
	for i := range ds.Providers {
		if ds.Providers[i].MailingAddress.State != nil && *ds.Providers[i].MailingAddress.State == sc {
			result = append(result, &ds.Providers[i])
		}
	}
	return result
}

// GetByTaxonomy returns all providers with a given taxonomy code.
func (ds *NppesDataset) GetByTaxonomy(taxonomyCode string) []*NppesRecord {
	if ds.taxonomyIndex != nil {
		indices := ds.taxonomyIndex[taxonomyCode]
		result := make([]*NppesRecord, 0, len(indices))
		for _, idx := range indices {
			result = append(result, &ds.Providers[idx])
		}
		return result
	}
	var result []*NppesRecord
	for i := range ds.Providers {
		for _, tc := range ds.Providers[i].TaxonomyCodes {
			if tc.Code == taxonomyCode {
				result = append(result, &ds.Providers[i])
				break
			}
		}
	}
	return result
}

// GetTaxonomyDescription returns the taxonomy reference for a code.
func (ds *NppesDataset) GetTaxonomyDescription(code string) *TaxonomyReference {
	if ds.TaxonomyMap == nil {
		return nil
	}
	ref, ok := ds.TaxonomyMap[code]
	if !ok {
		return nil
	}
	return &ref
}

// Query returns a new QueryBuilder for this dataset.
func (ds *NppesDataset) Query() *QueryBuilder {
	return &QueryBuilder{dataset: ds}
}

// Statistics returns dataset statistics.
func (ds *NppesDataset) Statistics() DatasetStatistics {
	return ComputeStatistics(ds)
}

// QueryBuilder provides a chainable query interface.
type QueryBuilder struct {
	dataset *NppesDataset
	filters []func(*NppesRecord) bool
}

// State adds a state filter.
func (qb *QueryBuilder) State(state string) *QueryBuilder {
	sc, ok := StateCodeFromCode(state)
	if !ok {
		qb.filters = append(qb.filters, func(r *NppesRecord) bool { return false })
		return qb
	}
	qb.filters = append(qb.filters, func(r *NppesRecord) bool {
		return r.MailingAddress.State != nil && *r.MailingAddress.State == sc
	})
	return qb
}

// Specialty adds a specialty (taxonomy display name) filter.
func (qb *QueryBuilder) Specialty(specialty string) *QueryBuilder {
	lower := strings.ToLower(specialty)
	qb.filters = append(qb.filters, func(r *NppesRecord) bool {
		for _, tc := range r.TaxonomyCodes {
			ref := qb.dataset.GetTaxonomyDescription(tc.Code)
			if ref != nil && ref.DisplayName != "" {
				if strings.Contains(strings.ToLower(ref.DisplayName), lower) {
					return true
				}
			}
		}
		return false
	})
	return qb
}

// PostalCode adds a postal code prefix filter matching against mailing or practice address.
func (qb *QueryBuilder) PostalCode(prefix string) *QueryBuilder {
	return qb.PostalCodes([]string{prefix})
}

// PostalCodes adds a filter matching any of the given postal code prefixes against mailing or practice address.
func (qb *QueryBuilder) PostalCodes(prefixes []string) *QueryBuilder {
	cleaned := make([]string, 0, len(prefixes))
	for _, p := range prefixes {
		if s := strings.TrimSpace(p); s != "" {
			cleaned = append(cleaned, s)
		}
	}
	if len(cleaned) == 0 {
		return qb
	}
	// Build a set for exact 5-digit lookups for performance
	exactSet := make(map[string]bool, len(cleaned))
	var prefixList []string
	for _, p := range cleaned {
		if len(p) == 5 {
			exactSet[p] = true
		} else {
			prefixList = append(prefixList, p)
		}
	}
	qb.filters = append(qb.filters, func(r *NppesRecord) bool {
		for _, pc := range []string{r.MailingAddress.PostalCode, r.PracticeAddress.PostalCode} {
			if pc == "" {
				continue
			}
			// Check exact set first (truncate to 5 chars for ZIP+4 codes like "100011234")
			zip5 := pc
			if len(zip5) > 5 {
				zip5 = zip5[:5]
			}
			if exactSet[zip5] {
				return true
			}
			// Then check prefix matches
			for _, pfx := range prefixList {
				if strings.HasPrefix(pc, pfx) {
					return true
				}
			}
		}
		return false
	})
	return qb
}

// ActiveOnly adds an active-only filter.
func (qb *QueryBuilder) ActiveOnly() *QueryBuilder {
	qb.filters = append(qb.filters, func(r *NppesRecord) bool {
		return r.IsActive()
	})
	return qb
}

// EntityTypeFilter adds an entity type filter.
func (qb *QueryBuilder) EntityTypeFilter(et EntityType) *QueryBuilder {
	qb.filters = append(qb.filters, func(r *NppesRecord) bool {
		return r.EntityType != nil && *r.EntityType == et
	})
	return qb
}

// Execute runs the query and returns all matching providers.
func (qb *QueryBuilder) Execute() []*NppesRecord {
	var results []*NppesRecord
	for i := range qb.dataset.Providers {
		p := &qb.dataset.Providers[i]
		match := true
		for _, f := range qb.filters {
			if !f(p) {
				match = false
				break
			}
		}
		if match {
			results = append(results, p)
		}
	}
	return results
}

// Count returns the number of matching providers.
func (qb *QueryBuilder) Count() int {
	return len(qb.Execute())
}

// Limit returns up to limit matching providers.
func (qb *QueryBuilder) Limit(limit int) []*NppesRecord {
	var results []*NppesRecord
	for i := range qb.dataset.Providers {
		p := &qb.dataset.Providers[i]
		match := true
		for _, f := range qb.filters {
			if !f(p) {
				match = false
				break
			}
		}
		if match {
			results = append(results, p)
			if len(results) >= limit {
				break
			}
		}
	}
	return results
}

// DatasetStatistics holds summary statistics for a dataset.
type DatasetStatistics struct {
	TotalProviders                 int
	IndividualProviders            int
	OrganizationProviders          int
	ActiveProviders                int
	InactiveProviders              int
	StatesRepresented              int
	UniqueTaxonomyCodes            int
	ProvidersWithOtherNames        int
	ProvidersWithPracticeLocations int
	ProvidersWithEndpoints         int
}

// ComputeStatistics calculates statistics from a dataset.
func ComputeStatistics(ds *NppesDataset) DatasetStatistics {
	stats := DatasetStatistics{
		TotalProviders: len(ds.Providers),
	}

	states := make(map[StateCode]bool)
	taxonomies := make(map[string]bool)

	for _, p := range ds.Providers {
		if p.EntityType != nil {
			switch *p.EntityType {
			case EntityTypeIndividual:
				stats.IndividualProviders++
			case EntityTypeOrganization:
				stats.OrganizationProviders++
			}
		}
		if p.IsActive() {
			stats.ActiveProviders++
		} else {
			stats.InactiveProviders++
		}
		if p.MailingAddress.State != nil {
			states[*p.MailingAddress.State] = true
		}
		for _, tc := range p.TaxonomyCodes {
			taxonomies[tc.Code] = true
		}
	}

	stats.StatesRepresented = len(states)
	stats.UniqueTaxonomyCodes = len(taxonomies)

	if ds.OtherNamesMap != nil {
		stats.ProvidersWithOtherNames = len(ds.OtherNamesMap)
	}
	if ds.PracticeLocationsMap != nil {
		stats.ProvidersWithPracticeLocations = len(ds.PracticeLocationsMap)
	}
	if ds.EndpointsMap != nil {
		stats.ProvidersWithEndpoints = len(ds.EndpointsMap)
	}

	return stats
}

// PrintSummary prints a formatted summary of the statistics.
func (s *DatasetStatistics) PrintSummary() {
	fmt.Println("=== NPPES Dataset Statistics ===")
	fmt.Printf("Total Providers: %d\n", s.TotalProviders)
	if s.TotalProviders > 0 {
		fmt.Printf("  Individual: %d (%.1f%%)\n",
			s.IndividualProviders,
			float64(s.IndividualProviders)/float64(s.TotalProviders)*100.0)
		fmt.Printf("  Organization: %d (%.1f%%)\n",
			s.OrganizationProviders,
			float64(s.OrganizationProviders)/float64(s.TotalProviders)*100.0)
		fmt.Printf("Active Providers: %d (%.1f%%)\n",
			s.ActiveProviders,
			float64(s.ActiveProviders)/float64(s.TotalProviders)*100.0)
	}
	fmt.Printf("States Represented: %d\n", s.StatesRepresented)
	fmt.Printf("Unique Taxonomy Codes: %d\n", s.UniqueTaxonomyCodes)
	if s.ProvidersWithOtherNames > 0 {
		fmt.Printf("Providers with Other Names: %d\n", s.ProvidersWithOtherNames)
	}
	if s.ProvidersWithPracticeLocations > 0 {
		fmt.Printf("Providers with Practice Locations: %d\n", s.ProvidersWithPracticeLocations)
	}
	if s.ProvidersWithEndpoints > 0 {
		fmt.Printf("Providers with Endpoints: %d\n", s.ProvidersWithEndpoints)
	}
}

// Helper functions to create lookup maps

func createTaxonomyMap(records []TaxonomyReference) map[string]TaxonomyReference {
	m := make(map[string]TaxonomyReference, len(records))
	for _, r := range records {
		m[r.Code] = r
	}
	return m
}

func createOtherNamesMap(records []OtherNameRecord) map[NPI][]OtherNameRecord {
	m := make(map[NPI][]OtherNameRecord)
	for _, r := range records {
		m[r.NPI] = append(m[r.NPI], r)
	}
	return m
}

func createPracticeLocationsMap(records []PracticeLocationRecord) map[NPI][]PracticeLocationRecord {
	m := make(map[NPI][]PracticeLocationRecord)
	for _, r := range records {
		m[r.NPI] = append(m[r.NPI], r)
	}
	return m
}

func createEndpointsMap(records []EndpointRecord) map[NPI][]EndpointRecord {
	m := make(map[NPI][]EndpointRecord)
	for _, r := range records {
		m[r.NPI] = append(m[r.NPI], r)
	}
	return m
}
