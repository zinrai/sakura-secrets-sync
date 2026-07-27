# sakura-secrets-sync

A command-line tool to sync sops-encrypted files to [SAKURA Cloud Secret Manager](https://cloud.sakura.ad.jp/products/secrets-manager/), one secret per file.

## Features

- Decrypts each file with an external command and writes the value to Secret Manager only when it differs from the stored value
- Derives the secret name from the file path, so the directory structure is the routing
- Reports every decision on stderr: `[create]`, `[update]`, `[unchanged]`, `[orphan]`
- Never deletes secrets. Orphans are only reported
- Candidate files come from an explicit list file, so what a CI run consumed stays inspectable

## Requirements

- SAKURA Cloud account with Secret Manager access
- Valid API credentials (static API keys or a service principal)
- A decrypt command such as [sops](https://github.com/getsops/sops) in PATH

## Configuration

Set the Vault resource ID:

```bash
$ export SAKURA_SECRETS_ID="your-vault-resource-id"
```

API credentials are resolved by [saclient-go](https://github.com/sacloud/saclient-go). Set either static API keys:

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

Sync the files listed in a candidates file (for example the output of a git change detector):

```bash
$ sakura-secrets-sync -dir staging -files changed.txt
```

Sync every file under the directory and report orphans:

```bash
$ sakura-secrets-sync -dir staging -all
```

### Options

- `-dir` (required): Environment directory to sync
- `-files`: File listing candidate paths, one per line (exclusive with `-all`)
- `-all`: Sync every file under `-dir` and report orphans (exclusive with `-files`)
- `-zone` (optional, default: `is1a`): Zone name
- `-decrypt-cmd` (optional, default: `sops -d`): Command that decrypts a file to stdout. The file path is appended as the last argument
- `-version`: Print version

## How It Works

For each candidate file:

1. The decrypt command is executed and its stdout becomes the secret value
2. Trailing newlines are stripped, because the store removes them on write and keeping them would make every comparison differ
3. If the secret does not exist it is created
4. If it exists, the stored value is fetched and the secret is written only when the values differ

The secret name is the file path relative to `-dir`, as-is. `staging/config/database.yaml` becomes `config/database.yaml`. Secret Manager accepts `/` in names (verified against the API, the manual does not document name constraints), so the repository layout carries over unmangled.

With `-all`, secrets that have no corresponding file are reported as `[orphan]` and left untouched. Deleting them is a manual operation.

Candidate paths must be under `-dir`. An empty candidates file is a successful no-op.

## Exit Codes

- `0`: All candidates synced (orphans do not affect the exit code)
- `1`: Error (decrypt failure, API failure, invalid candidate, etc.)

## License

This project is licensed under the [MIT License](./LICENSE).
