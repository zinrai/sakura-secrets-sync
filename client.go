package main

import (
	"context"
	"fmt"
	"os"

	secretmanager "github.com/sacloud/sacloud-sdk-go/api/secretmanager"
	v1 "github.com/sacloud/sacloud-sdk-go/api/secretmanager/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
)

// NewSecretOp creates a Secret Manager API client bound to the specified Vault.
// Credentials are resolved by saclient-go from environment variables,
// supporting both static API keys and service principals.
func NewSecretOp(zone, vaultID string) (secretmanager.SecretAPI, error) {
	endpoint := fmt.Sprintf("SAKURA_ENDPOINTS_SECRETMANAGER=https://secure.sakura.ad.jp/cloud/zone/%s/api/cloud/1.1", zone)

	var sc saclient.Client
	if err := sc.SetEnviron(append(os.Environ(), endpoint)); err != nil {
		return nil, fmt.Errorf("failed to configure saclient: %w", err)
	}
	if err := sc.Populate(); err != nil {
		return nil, fmt.Errorf("failed to configure saclient: %w", err)
	}

	client, err := secretmanager.NewClient(&sc)
	if err != nil {
		return nil, fmt.Errorf("failed to create Secret Manager client: %w", err)
	}

	return secretmanager.NewSecretOp(client, vaultID), nil
}

// ListNames returns the names of all secrets in the Vault
func ListNames(op secretmanager.SecretAPI) (map[string]bool, error) {
	secrets, err := op.List(context.Background())
	if err != nil {
		return nil, err
	}

	names := make(map[string]bool, len(secrets))
	for _, s := range secrets {
		names[s.Name] = true
	}
	return names, nil
}

// PutSecret registers a secret value to the Vault
func PutSecret(op secretmanager.SecretAPI, name, value string) error {
	_, err := op.Create(context.Background(), v1.CreateSecret{
		Name:  name,
		Value: value,
	})
	return err
}

// GetSecret retrieves the latest value of a secret using the unveil API
func GetSecret(op secretmanager.SecretAPI, name string) (string, error) {
	res, err := op.Unveil(context.Background(), v1.Unveil{
		Name: name,
	})
	if err != nil {
		return "", err
	}
	return res.Value, nil
}
