package main

import (
	"context"
	"fmt"
	"os"

	secretmanager "github.com/sacloud/sacloud-sdk-go/api/secretmanager"
	v1 "github.com/sacloud/sacloud-sdk-go/api/secretmanager/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
)

func NewClient(zone string) (*v1.Client, error) {
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

	return client, nil
}

func ListNames(client *v1.Client, vaultID string) (map[string]bool, error) {
	res, err := client.ListVaultSecrets(context.Background(), v1.ListVaultSecretsParams{
		VaultResourceID: vaultID,
	})
	if err != nil {
		return nil, err
	}
	// Not paged with Count and From: their defaults and semantics are
	// undocumented, so a partial list is refused rather than guessed past.
	if len(res.Secrets) != res.Total {
		return nil, fmt.Errorf("listed %d of %d secrets, refusing to sync", len(res.Secrets), res.Total)
	}

	names := make(map[string]bool, len(res.Secrets))
	for _, s := range res.Secrets {
		names[s.Name] = true
	}
	return names, nil
}

func PutSecret(op secretmanager.SecretAPI, name, value string) error {
	_, err := op.Create(context.Background(), secretmanager.CreateSecretParams{
		Name:  name,
		Value: value,
	})
	return err
}

func DeleteSecret(op secretmanager.SecretAPI, name string) error {
	return op.Delete(context.Background(), name)
}

func GetSecret(op secretmanager.SecretAPI, name string) (string, error) {
	res, err := op.Unveil(context.Background(), secretmanager.UnveilParams{
		Name: name,
	})
	if err != nil {
		return "", err
	}
	return res.Value, nil
}
