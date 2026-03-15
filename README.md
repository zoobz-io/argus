# Argus

Multi-tenant document ingestion platform. Connects to cloud storage providers, watches for changes, and builds a searchable knowledge base with versioning, content extraction, AI summaries, and vector embeddings.

## What It Does

1. **Watch** — Connects to cloud storage providers via [flux](https://github.com/zoobzio/flux) and monitors files the tenant has registered for observation
2. **Extract** — Pulls document content using format-specific extractors; delegates OCR to a Tesseract sidecar over gRPC for scanned/handwritten documents
3. **Version** — Tracks every revision of every document, maintaining a complete history
4. **Enrich** — Generates AI summaries via [zyn](https://github.com/zoobzio/zyn) and vector embeddings via [vex](https://github.com/zoobzio/vex) for each document version
5. **Index** — Stores extracted content, summaries, and embeddings in OpenSearch, providing full-text and semantic search across all ingested documents for a given tenant

## Storage Providers

Each provider implements a common interface and is developed independently:

- Google Drive
- OneDrive / SharePoint
- Dropbox
- Amazon S3
- Google Cloud Storage
- Azure Blob Storage

## Supported Document Types

| Category | Formats |
|----------|---------|
| Documents | PDF, DOCX, DOC, ODT, RTF, TXT, Markdown |
| Spreadsheets | XLSX, XLS, CSV, ODS |
| Presentations | PPTX, PPT, ODP |
| Images (OCR) | PNG, JPEG, TIFF, BMP, WebP |
| Scanned Documents | PDF (image-only), multi-page TIFF |

## Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Provider   │────▶│   Ingestion  │────▶│  Enrichment  │
│   Watchers   │     │   Pipeline   │     │   Pipeline   │
│  (flux)      │     │              │     │  (zyn + vex) │
└──────────────┘     └──────┬───────┘     └──────┬───────┘
                            │                    │
                     ┌──────▼───────┐     ┌──────▼───────┐
                     │   Tesseract  │     │  OpenSearch   │
                     │   Sidecar    │     │   Cluster     │
                     │   (gRPC)     │     │              │
                     └──────────────┘     └──────────────┘
```

- **Provider Watchers** — flux capacitors monitor registered files/folders across cloud storage providers
- **Ingestion Pipeline** — Extracts content, normalises formats, manages document versions; delegates OCR to a Tesseract sidecar via gRPC
- **Enrichment Pipeline** — Generates AI summaries (zyn) and vector embeddings (vex) per document version
- **OpenSearch** — Full-text and semantic search index per tenant

The system is designed for horizontal scalability from day one — pipelines are queue-driven and stateless, allowing independent scaling of ingestion, OCR, and enrichment workloads.

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go |
| Framework | [sum](https://github.com/zoobzio/sum) |
| Configuration | [flux](https://github.com/zoobzio/flux) |
| LLM Orchestration | [zyn](https://github.com/zoobzio/zyn) |
| Embeddings | [vex](https://github.com/zoobzio/vex) |
| OCR | Tesseract (gRPC sidecar) |
| Search & Storage | OpenSearch |
| Database | PostgreSQL |
| Object Storage | MinIO (dev) / S3-compatible (prod) |
| Observability | OpenTelemetry |

## Development

```bash
make dev        # Start local infrastructure
make run        # Run the application
make test       # Run tests
make check      # Run tests + lint
```

## License

MIT
