# Security policy

## Supported versions

Security fixes are applied to the latest release on the default branch.

## Reporting a vulnerability

Please report security issues through [GitHub private vulnerability reporting](https://github.com/0xSidBanerjee/wortwerk/security/advisories/new) if that feature is enabled, or by opening a minimal public issue asking for a private channel (do not include exploit details in public issues).

Do not commit API keys, `.env` files, or LLM responses containing sensitive data.

## Scope

The Wortwerk app is static: no accounts, no server-side user data, and no analytics. Ingestion tools run locally and use environment variables for LLM credentials.
