# Parse DMARC on VibeNest

[![Deploy to VibeNest](https://vibenest.net/deploy-to-vibenest.svg)](https://vibenest.net/deploy?repo=https://github.com/NikitaBabenko/parse-dmarc)

Independent deployment fork of [dmarcguardhq/parse-dmarc](https://github.com/dmarcguardhq/parse-dmarc), under the unchanged Apache-2.0 license. No upstream endorsement is implied.

Historical result on 2026-10-18: **the small synthetic Maildir/API scenario passed on VibeNest Free**, at code revision `6bb48714f1bb34202b466d03a0a8a62b010b3392`. A production build, authenticated HTTPS report retrieval and preservation after one normal container replacement were verified. The button opens a repository preview; manual configuration below is required before deployment. This is not unattended one-click hosting.

The report contained 10 fictional messages, 8 compliant (80%). Anonymous and invalid credentials returned 401 for the dashboard HTML, API and metrics; authenticated HTML, JS/CSS and report endpoints returned 200. Twenty sequential statistics requests succeeded. The rendered dashboard was not visually accepted: the available browser tool could not complete Basic Auth. Real mailbox ingestion, larger workloads and long-term stability remain unverified.

Runtime evidence: actual memory limit 256 MB; platform sampled memory 2.64–4.24 MB across the observed initial/replacement window, while the application reported process RSS about 18.5–19.4 MiB at two checks (these are different measurements, not interchangeable peaks). No unexpected restarts were observed. Initial dispatch through build to successful deployment took about 6 min 38 s. Replacement used a new container, retained the identical SQLite report and found no new Maildir messages, demonstrating persistence without reseeding.

## Button flow prepared on 2026-10-22 - acceptance pending

The root `compose.yml` describes one application on port 8080, a named `report-data` volume at `/data`, and a fictional Maildir fixture. The upstream local Dovecot stack is unchanged in `compose.dev.yml`. The original Dockerfile is unchanged at `deploy/vibenest/Dockerfile`, built with context `.`. A root Dockerfile would cause the current VibeNest detector to collapse a one-service Compose recipe and lose its configuration.

In inspected platform revision `3ebf31e92d075e5f489eb14e5d51b9f3e2683dd5`, Compose setup reads `.env.example`, generates a unique 192-bit `HTTP_AUTH_PASSWORD` when absent, persists it as a sensitive project variable, and supplies runtime interpolation. It preserves an existing owner value. This is a source-based expectation, **not a runtime result for this new recipe**. No secret value is committed. Missing/invalid credentials still fail closed: no shared password, anonymous data access, or public first-visitor account claim.

### First deployment and first login

1. Click the badge and verify this repository and `main`. The current public preview uses a simpler detector and is expected to say "Auto-detected build (Nixpacks)"; it does not show the final Compose recipe. It must remain a solo-project flow, not a group wizard. After sign-in, verify that the actual deployment resolver selects `compose.yml`, service `parse-dmarc`, port 8080; stop on a Dovecot stack or a retained forced Dockerfile recipe.
2. Register or sign in through VibeNest. After authentication, `/dashboard/launch` starts deployment automatically. VibeNest login does not authenticate the separate Parse DMARC origin.
3. Open the resulting project's Settings > Environment and reveal `HTTP_AUTH_PASSWORD` through the normal owner-only Reveal action. Username: `pilot-owner`, unless you changed `HTTP_AUTH_USER`. **Manual credential retrieval is required.** If generation/persistence did not occur, configure a unique password through Settings before any deliberate retry; never disable authentication.
4. Open the HTTPS app and enter those credentials in its Basic Auth prompt. Check the synthetic `example.test` report: 10 messages, 8 compliant (80%). A green deployment badge or blank page is insufficient.
5. Confirm data survives a normal container replacement before relying on persistence. Declaring a named volume is not proof it was materialized, and persistence is not a backup.

The current platform Compose route refuses GitHub App direct source mode. Use only a supported explicit public/legacy source choice exposed by the product; never silently downgrade a persisted grant. This source choice remains part of pending acceptance.

### Existing projects and real reports

VibeNest reuses an existing project for the same owner/repository. Old forced build packs, mounts, credentials and data can mask fresh-install defects. Do not blindly convert an existing project: the Compose named volume may differ from its old standalone volume. Preserve data and plan that migration explicitly. A successful redeploy of a preconfigured project is not fresh-user acceptance.

The default recipe intentionally demonstrates fictional Maildir data and requires no real mailbox or DNS changes. Real use is a later owner configuration: disable the fixture, choose a clean database or retain the labeled sample deliberately, supply Maildir or your own IMAP host/credentials, and configure DMARC `rua` for your domain. These are application inputs, not prerequisites for the synthetic build/demo. The Compose environment is explicit: editing the deployment configuration is required to switch to real IMAP; do not assume conflicting dashboard values override YAML literals.

No complete badge > authentication > deployment > working browser dashboard run has been accepted for this new recipe. The measurements below describe only the historical manually configured Dockerfile deployment.

## Historical manual configuration

- Historical manual recipe: branch `main`, context `/`, Dockerfile now at `deploy/vibenest/Dockerfile`. The new button recipe uses Compose; do not force the old Dockerfile build pack.
- Container port: `8080`; `SERVER_HOST=0.0.0.0`, `SERVER_PORT=8080`.
- Persistent writable volume: mount `/data`. Keep `DATABASE_PATH=/data/parse-dmarc.db`. The scratch image has no shell; do not use a shell startup command.
- Required secrets: `HTTP_AUTH_USER` (your chosen nonempty username, no colon) and `HTTP_AUTH_PASSWORD` (unique randomly generated password, at least 24 characters; recommended 32 random bytes encoded as hex). There is no default password. Configure these as runtime secrets, never Docker build arguments or committed files.
- Only access the public endpoint over HTTPS. The browser's Basic Auth prompt protects the dashboard, static files, API, unknown routes and `/metrics`. Anonymous or invalid requests return 401, including OPTIONS. Only GET/HEAD `/healthz` is public, with an empty 200 response. Set the platform health path to `/healthz` if HTTP health probes are enabled.
- `GOMEMLIMIT=180MiB`, `GOMAXPROCS=1` were used for the small pilot. Free catalog: 256 MB RAM, 0.5 platform CPU, 4 GB plan storage. The 256 MB memory limit was observed. The selected host used a CPU quota scale of 0.5, so 0.5 platform CPU corresponds to an expected 0.25 host CPU; the actual CPU cgroup quota was not independently inspected. A hard 4 GB storage quota and instantaneous memory peak were not verified.
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
