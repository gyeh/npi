# NPI

An importable Go library and CLI for downloading, querying, analyzing, and exporting NPI healthcare provider data from NPPES (National Plan and Provider Enumeration System).

## Use Cases

- **Import as a library** into Go services that need provider lookup, filtering, or export
- **Download NPPES data** from CMS without navigating their website manually
- **Explore provider data** interactively - look up providers by NPI, state, or specialty
- **Generate dataset statistics** - total providers, entity type breakdown, state distribution, top specialties
- **Export filtered subsets** to JSON, normalized CSV, or PostgreSQL SQL for downstream pipelines
- **Integrate into shell scripts** and CI jobs that need provider data in a specific format

## Library Usage

```go
import "npi/npi"
```

### Loading a dataset

```go
// Silent (no progress output) — suitable for library consumers
ds, err := npi.LoadStandard("./data")

// With progress logging
ds, err := npi.LoadStandard("./data", nppes.WithLogger(func(format string, args ...interface{}) {
    fmt.Printf(format, args...)
}))
```

### Querying providers

```go
// By NPI
provider := ds.GetByNPI("1234567890")

// By state
caProviders := ds.GetByState("CA")

// Chainable query builder
results := ds.Query().
    State("CA").
    PostalCode("90210").
    ActiveOnly().
    Execute()

// With specialty filter and limit
results := ds.Query().
    Specialty("Cardiology").
    Limit(100)
```

### Analytics

```go
analytics := npi.NewNppesAnalytics(ds.Providers)

byName := analytics.FindByName("smith")
topStates := analytics.TopStatesByProviderCount(10)
stats := analytics.ComputeDatasetStats()
```

### Exporting

```go
// Direct export functions
npi.ExportJSON(providers, "output.json")
npi.ExportCSV(providers, "output.csv")
npi.ExportSQL(providers, "output.sql")

// Via the Exporter interface
exporter, _ := npi.NewExporter(npi.ExportFormatJSON)
exporter.Export(providers, "output.json")

// Filtered subset export
npi.ExportSubset(ds, "ca_providers.json", func(p *nppes.NppesRecord) bool {
    return p.MailingAddress.State != nil && p.MailingAddress.State.AsCode() == "CA"
}, npi.ExportFormatJSON)
```

### Downloading NPI data

```go
downloader := npi.NewNppesDownloaderWithConfig(nppes.DownloadConfig{
    OutputDir: "./data",
    KeepFiles: true,
})
files, err := downloader.DownloadLatestNPPES()
```

### Interfaces

The library defines interfaces for all major components, making it easy to mock in tests:

- `npi.Dataset` — query and lookup operations on loaded data
- `npi.Loader` — CSV file parsing (`NppesReader`)
- `npi.Exporter` — data export (`JSONExporter`, `CSVExporter`, `SQLExporter`)
- `npi.Downloader` — file download and extraction (`NppesDownloader`)
- `npi.Analyzer` — analytics and aggregation (`NppesAnalytics`)

## CLI Installation

```sh
cd npi
go build -o npi ./cmd/npi
```

This produces a single `npi` binary. Move it to a directory on your `$PATH` if desired.

## Usage

### Download the latest NPPES data

```sh
npi download --out-dir ./data
```

Downloads the current month's full NPPES data dissemination ZIP (~1GB compressed) from CMS, extracts it, and categorizes the files (main data, taxonomy reference, endpoints, etc.).

### Show dataset statistics

```sh
npi stats --data-dir ./data
```

Output includes total provider count, individual vs. organization breakdown, active providers, unique states, and top taxonomy codes.

### Query providers

```sh
# By state
npi query --data-dir ./data --state CA

# By specialty (substring match on taxonomy display name)
npi query --data-dir ./data --specialty Cardiology

# By postal code prefix (matches mailing or practice address)
npi query --data-dir ./data --postal-code 90210
npi query --data-dir ./data --postal-code 902    # matches all 902xx zip codes

# Combined filters
npi query --data-dir ./data --state NY --specialty "Internal Medicine" --active --limit 50
npi query --data-dir ./data --state CA --postal-code 902 --active

# By NPI
npi query --data-dir ./data --npi 1234567890
```

Output format: `NPI | Name | EntityType | State`

### Export data

```sh
# JSON (pretty-printed array)
npi export --data-dir ./data --output providers.json --format json --state CA

# Normalized CSV (produces two files: *_providers.csv and *_taxonomies.csv)
npi export --data-dir ./data --output providers.csv --format csv --specialty Cardiology

# Filter by postal code prefix (matches mailing or practice address)
npi export --data-dir ./data --output local.json --format json --postal-code 902

# PostgreSQL SQL (CREATE TABLE + batched INSERTs)
npi export --data-dir ./data --output providers.sql --format sql --state NY --active
```

## Data Files

The CLI expects a directory containing NPPES files with standard CMS naming:

| File pattern | Description |
|---|---|
| `npidata_pfile_*.csv` | Main provider data (~8M records, 330 columns) |
| `nucc_taxonomy_*.csv` | NUCC taxonomy reference codes |
| `othername_pfile_*.csv` | Additional organization names |
| `pl_pfile_*.csv` | Non-primary practice locations |
| `endpoint_pfile_*.csv` | Healthcare endpoints |

Use `npi download` to obtain these files, or download manually from [CMS NPPES](https://download.cms.gov/nppes/NPI_Files.html).

## Performance

- The main NPI file is ~10GB with ~8M records
- Parsing the full dataset requires 16GB+ RAM
- Indexes are built in-memory for NPI, state, and taxonomy lookups after loading

## License

MIT
