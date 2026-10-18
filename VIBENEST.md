# Parse DMARC on VibeNest

[![Deploy to VibeNest](https://vibenest.net/deploy-to-vibenest.svg)](https://vibenest.net/deploy?repo=https://github.com/NikitaBabenko/parse-dmarc)

Independent deployment fork of [dmarcguardhq/parse-dmarc](https://github.com/dmarcguardhq/parse-dmarc), under the unchanged Apache-2.0 license. No upstream endorsement is implied.

Status on 2026-10-18: prepared for a guided pilot. Production build, HTTPS acceptance, Free-tier measurements and persistence after replacement are **not verified**. The button opens a repository preview; manual configuration below is required before deployment.

## Configure before starting

- Repository: this fork, branch `main`, build context `/`, root `Dockerfile`.
- Container port: `8080`; `SERVER_HOST=0.0.0.0`, `SERVER_PORT=8080`.
- Persistent writable volume: mount `/data`. Keep `DATABASE_PATH=/data/parse-dmarc.db`. The scratch image has no shell; do not use a shell startup command.
- Required secrets: `HTTP_AUTH_USER` (your chosen nonempty username, no colon) and `HTTP_AUTH_PASSWORD` (unique randomly generated password, at least 24 characters; recommended 32 random bytes encoded as hex). There is no default password. Configure these as runtime secrets, never Docker build arguments or committed files.
- Only access the public endpoint over HTTPS. The browser's Basic Auth prompt protects the dashboard, static files, API, unknown routes and `/metrics`. Anonymous or invalid requests return 401, including OPTIONS. Only GET/HEAD `/healthz` is public, with an empty 200 response. Set the platform health path to `/healthz` if HTTP health probes are enabled.
- `GOMEMLIMIT=180MiB`, `GOMAXPROCS=1` are conservative pilot settings, not proof that the app fits the plan. Target: 256 MB RAM, 0.5 platform CPU, 4 GB plan storage. Applied limits, CPU semantics, storage quota and memory peak still need runtime evidence.
- Leave MCP mode and `PARSE_DMARC_MCP_HTTP` disabled/unset. This fork's middleware protects the dashboard HTTP server, not a separately configured MCP listener.
- GitHub Actions on this fork are disabled to prevent inherited build/release workflows from consuming minutes.

## Synthetic acceptance without a mailbox

Set `MAILDIR_PATH=/data/maildir`, `VIBENEST_SYNTHETIC_FIXTURE=true`, `FETCH_INTERVAL=300`, `PARSE_DMARC_SERVE_ONLY=false`, and `PARSE_DMARC_METRICS=true`. Leave all IMAP credentials unset. Use the Dockerfile's default entrypoint and command.

The opt-in fixture creates Maildir directories and one fictional MIME email, then uses the normal Maildir parser and SQLite storage. It has domain `example.test`, source addresses `192.0.2.10` and `198.51.100.20`, 10 messages, 8 compliant messages (80%), and report ID `vibenest-synthetic-20261018`. It is not re-created if already present in `new/` or `cur/`; no private mailbox or DNS changes are needed.

Acceptance must check authenticated report list, statistics, report details, top sources and dashboard; anonymous rejection on all data routes; and preservation of the same report after one normal container replacement with `/data` retained. A blank 200 dashboard is insufficient. The constant `/healthz` response alone does not verify ingestion or database health.

## Your own reports later

For a new real installation, disable `VIBENEST_SYNTHETIC_FIXTURE` and choose a clean persistent database or knowingly retain the clearly labeled sample. Either supply your own Maildir with `new`, `cur`, `tmp` directories, or unset `MAILDIR_PATH` and configure `IMAP_HOST`, `IMAP_PORT=993`, `IMAP_USERNAME`, `IMAP_PASSWORD`, `IMAP_USE_TLS=true`, and optionally `IMAP_MAILBOX`. Never disable TLS certificate verification for a public mailbox. Use a dedicated reporting mailbox and provider-supported app credentials. The upstream defaults mark fetched mail as seen; review those mailbox effects before use.

Mail receivers need a DMARC `rua` record pointing to a mailbox you control. Configure this yourself for your own domain; this pilot performs no real-domain DNS edits. Keep the HTTP secret enabled before adding any real reports. Keep the frontend API on the same origin (the default); do not direct the dashboard to an unrelated API endpoint.

No Prometheus, Grafana, MCP or additional database sidecar is required. Back up the SQLite database and Maildir together; a volume is persistence, not a backup. Free-plan suitability for large real mailboxes is unknown.
