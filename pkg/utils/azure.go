package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
)

func VaultNameToURL(vaultName string) string {
	return fmt.Sprintf("https://%s.vault.azure.net", vaultName)
}

func CreatVaultClient(vaultURL string, credential *azidentity.DefaultAzureCredential) (*azsecrets.Client, error) {
	return azsecrets.NewClient(vaultURL, credential, nil)
}

type AzureClient struct {
	SubscriptionID string
	Credential     *azidentity.DefaultAzureCredential
}

func NewAzureClient(subscriptionID string) (*AzureClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	return &AzureClient{
		SubscriptionID: subscriptionID,
		Credential:     cred,
	}, nil
}

func (c *AzureClient) CreateSecretsClient(vaultName string) (*azsecrets.Client, error) {
	vaultURL := fmt.Sprintf("https://%s.vault.azure.net", vaultName)
	return azsecrets.NewClient(vaultURL, c.Credential, nil)
}

type ExportSecret struct {
	Name         string
	Value        string
	ContentType  *string
	Attributes   *azsecrets.SecretAttributes
	Tags         map[string]*string
	ID           string
	Version      string
	VaultURL     string
}

func (c *AzureClient) ExportSecrets(ctx context.Context, vaultName string) ([]ExportSecret, error) {
	var exportSecrets []ExportSecret
	secretsClient, err := c.CreateSecretsClient(vaultName)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret client: %w", err)
	}
	pager := secretsClient.NewListSecretsPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get secrets page: %w", err)
		}

		for _, secretItem := range page.Value {
			secret, err := secretsClient.GetSecret(ctx, secretItem.ID.Name(), "", nil)
			if err != nil {
				return nil, fmt.Errorf("failed to get secret value for %s: %w", secretItem.ID.Name(), err)
			}

			exportSecrets = append(exportSecrets, ExportSecret{
				Name:        secretItem.ID.Name(),
				Value:       *secret.Value,
				ContentType: secretItem.ContentType,
				Attributes:  secretItem.Attributes,
				Tags:        secretItem.Tags,
				ID:          secretItem.ID.Name(),
				Version:     secretItem.ID.Version(),
			})
		}
	}

	return exportSecrets, nil
}

func WriteToJSONFile(data interface{}, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode data to JSON: %w", err)
	}

	return nil
}

func (c *AzureClient) MigrateSecrets(
	ctx context.Context,
	sourceVaultName string,
	destVaultName string,
) error {

	sourceClient, err := c.CreateSecretsClient(sourceVaultName)
	if err != nil {
		return fmt.Errorf("failed to create source secret client: %w", err)
	}
	destClient, err := c.CreateSecretsClient(destVaultName)
	if err != nil {
		return fmt.Errorf("failed to create destination secret client: %w", err)
	}
	pager := sourceClient.NewListSecretsPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to get secrets page: %w", err)
		}

		for _, secret := range page.Value {
			// Get full secret with value
			resp, err := sourceClient.GetSecret(ctx, secret.ID.Name(), "", nil)
			if err != nil {
				return fmt.Errorf("failed to get secret %s: %w", secret.ID.Name(), err)
			}
			// Create in destination vault
			_, err = destClient.SetSecret(ctx, secret.ID.Name(), azsecrets.SetSecretParameters{
				Value:       resp.Value,
				ContentType: secret.ContentType,
				SecretAttributes: &azsecrets.SecretAttributes{
					Enabled:   secret.Attributes.Enabled,
					Expires:   secret.Attributes.Expires,
					NotBefore: secret.Attributes.NotBefore,
				},
				Tags: secret.Tags,
			}, nil)
			if err != nil {
				return fmt.Errorf("failed to set secret %s: %w", secret.ID.Name(), err)
			}

			fmt.Printf("Successfully migrated secret: %s\n", secret.ID.Name())
		}
	}
	return nil
}

