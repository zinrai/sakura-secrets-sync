# sakura-secrets-sync

A command-line tool to sync a directory of sops-encrypted files to [SAKURA Cloud Secret Manager](https://cloud.sakura.ad.jp/products/secrets-manager/), one secret per file.

## Features

- Mirrors the directory to the vault: creates missing secrets, updates changed ones, deletes secrets that no file corresponds to
- Decrypts each file with an external command and writes the value only when it differs from the stored value
- Derives the secret name from the file path, so the directory structure is the routing
- Reports every decision on stderr: `[create]`, `[update]`, `[unchanged]`, `[delete]`
- `-dry-run` reports the same decisions without writing to the store
- Refuses to run against an empty directory, because that would mean deleting every secret in the vault

## Requirements

- SAKURA Cloud account with Secret Manager access
- Valid API credentials (static API keys or a service principal)
- A decrypt command such as [sops](https://github.com/getsops/sops) in PATH

## Configuration

Set the Vault resource ID:

```bash
$ export SAKURA_SECRETS_ID="your-vault-resource-id"
```

API credentials are resolved by [sacloud-sdk-go](https://github.com/sacloud/sacloud-sdk-go). Set either static API keys:

```bash
$ export SAKURA_ACCESS_TOKEN="your-access-token"
$ export SAKURA_ACCESS_TOKEN_SECRET="your-access-token-secret"
```

or service principal credentials:

```bash
$ export SAKURA_SERVICE_PRINCIPAL_ID="your-service-principal-id"
$ export SAKURA_SERVICE_PRINCIPAL_KEY_ID="your-key-id"
$ export SAKURA_PRIVATE_KEY_PATH="/path/to/private-key.pem"
```

## Usage

Sync a directory to the vault:

```bash
$ sakura-secrets-sync -dir staging
walked 3 file(s) under staging
[unchanged] staging/config/database.yaml -> config/database.yaml
[update] staging/config/production.env -> config/production.env
[create] staging/ssh_keys/id_ed25519 -> ssh_keys/id_ed25519
[delete] old/retired.yaml
```

See what a sync would do without changing the store:

```bash
$ sakura-secrets-sync -dir staging -dry-run
```

### Options

- `-dir` (required): Directory to sync
- `-dry-run`: Report decisions without writing to the store
- `-zone` (optional, default: `is1a`): Zone name
- `-decrypt-cmd` (optional, default: `sops -d`): Command that decrypts a file to stdout. The file path is appended as the last argument
- `-version`: Print version

## How It Works

For each file under `-dir`:

1. The decrypt command is executed and its stdout becomes the secret value
2. Trailing newlines are stripped, because the store removes them on write and keeping them would make every comparison differ
3. If the secret does not exist it is created
4. If it exists, the stored value is fetched and the secret is written only when the values differ

Secrets in the vault that no file corresponds to are then deleted, so the vault mirrors the directory.

The secret name is the file path relative to `-dir`, as-is. `staging/config/database.yaml` becomes `config/database.yaml`. Secret Manager accepts `/` in names (verified against the API, the manual does not document name constraints), so the repository layout carries over unmangled.

## Exit Codes

- `0`: The vault mirrors the directory
- `1`: Error (decrypt failure, API failure, empty directory, etc.)

## License

This project is licensed under the [MIT License](./LICENSE).
