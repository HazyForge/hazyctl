package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/spf13/cobra"
)

var AzureCmd = &cobra.Command{
	Use:   "azure",
	Short: "Azure Key Vault management utilities",
}

type AzureClient struct {
	Subscription   string
	VaultsClient   *armkeyvault.VaultsClient
	SecretsClient  *azsecrets.Client
	Credential     *azidentity.DefaultAzureCredential
}

func init() {
	AzureCmd.AddCommand(newMigrateCmd())
	AzureCmd.AddCommand(newExportCmd())
	AzureCmd.PersistentFlags().StringP("subscription", "s", "", "Azure subscription ID")
	AzureCmd.MarkPersistentFlagRequired("subscription")
}

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate secrets between Key Vaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			subscription := cmd.Flag("subscription").Value.String()
			sourceVault, _ := cmd.Flags().GetString("source")
			destVault, _ := cmd.Flags().GetString("destination")

			client, err := NewAzureClient(subscription, "", "") // Initialize without data plane clients
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}
			sourceVaultURL := fmt.Sprintf("https://%s.vault.azure.net", sourceVault)
			// Initialize data plane clients for source and destination
			sourceSecretClient, err := azsecrets.NewClient(sourceVaultURL, client.Credential, nil)
			if err != nil {
				return fmt.Errorf("failed to create source secret client: %w", err)
			}
			
			destVaultURL := fmt.Sprintf("https://%s.vault.azure.net", destVault)
			destSecretClient, err := azsecrets.NewClient(destVaultURL, client.Credential, nil)
			if err != nil {
				return fmt.Errorf("failed to create destination secret client: %w", err)
			}

			return client.MigrateSecrets(ctx, sourceSecretClient, destSecretClient)
		},
	}

	cmd.Flags().String("source", "", "Source Key Vault URL (e.g. https://src-vault.vault.azure.net)")
	cmd.Flags().String("destination", "", "Destination Key Vault URL (e.g. https://dst-vault.vault.azure.net)")
	cmd.MarkFlagRequired("source")
	cmd.MarkFlagRequired("destination")

	return cmd
}

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export Key Vault secrets and certificates",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			subscription := cmd.Flag("subscription").Value.String()
			fmt.Println("subscription", subscription)
			name, _ := cmd.Flags().GetString("name")
			outputPath, _ := cmd.Flags().GetString("output")

			// Create data plane client directly
			cred, err := azidentity.NewDefaultAzureCredential(nil)
			if err != nil {
				return fmt.Errorf("failed to create credential: %w", err)
			}
			url := fmt.Sprintf("https://%s.vault.azure.net", name)
			secretClient, err := azsecrets.NewClient(url, cred, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret client: %w", err)
			}

			return ExportSecrets(ctx, secretClient, outputPath)
		},
	}

	cmd.Flags().StringP("name", "n", "", "Name of the vault")
	cmd.Flags().StringP("output", "o", "secrets.json", "Output file path")
	cmd.MarkFlagRequired("name")

	return cmd
}

func NewAzureClient(subscriptionID, vaultURL, certURL string) (*AzureClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	vaultsClient, err := armkeyvault.NewVaultsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create vaults client: %w", err)
	}

	var secretsClient *azsecrets.Client

	if vaultURL != "" {
		secretsClient, err = azsecrets.NewClient(vaultURL, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create secrets client: %w", err)
		}
	}

	if certURL != "" {
		if err != nil {
			return nil, fmt.Errorf("failed to create certificates client: %w", err)
		}
	}

	return &AzureClient{
		Subscription:  subscriptionID,
		VaultsClient:  vaultsClient,
		SecretsClient: secretsClient,
		Credential:    cred,
	}, nil
}

func (c *AzureClient) MigrateSecrets(
	ctx context.Context,
	sourceClient *azsecrets.Client,
	destClient *azsecrets.Client,
) error {

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

func ExportSecrets(ctx context.Context, client *azsecrets.Client, outputPath string) error {
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

	var exportSecrets []ExportSecret

	pager := client.NewListSecretsPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to get secrets page: %w", err)
		}

		for _, secretItem := range page.Value {
			// Get the actual secret value
			secret, err := client.GetSecret(ctx, secretItem.ID.Name(), "", nil)
			if err != nil {
				return fmt.Errorf("failed to get secret value for %s: %w", secretItem.ID.Name(), err)
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

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(exportSecrets); err != nil {
		return fmt.Errorf("failed to encode secrets: %w", err)
	}

	fmt.Printf("Exported %d secrets to %s\n", len(exportSecrets), outputPath)
	return nil
}