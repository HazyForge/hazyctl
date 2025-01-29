package azure

import (
	"fmt"
	azureUtils "github.com/hazyforge/hazyctl/pkg/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var AzureCmd = &cobra.Command{
	Use:   "azure",
	Short: "Azure Key Vault management utilities",
}


func init() {
	AzureCmd.AddCommand(newMigrateCmd())
	AzureCmd.AddCommand(newExportCmd())
	AzureCmd.PersistentFlags().StringP("subscription", "s", "", "Azure subscription ID")
	AzureCmd.MarkPersistentFlagRequired("subscription")
	viper.BindPFlag("azure.subscription", AzureCmd.PersistentFlags().Lookup("subscription"))
}

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate secrets between Key Vaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
 			sourceVaultName := viper.GetString("azure.migrate.source")
			destVaultName := viper.GetString("azure.migrate.destination")
			subscriptionID := viper.GetString("azure.subscription")
			fmt.Println("Using subscription ", subscriptionID)
			client, err := azureUtils.NewAzureClient(subscriptionID)
			if err != nil {
				return fmt.Errorf("failed to create Azure client: %w", err)
			}
			fmt.Println("Copying secrets from ", sourceVaultName, " to ", destVaultName)
			return client.MigrateSecrets(ctx, sourceVaultName, destVaultName)
		},
	}

	cmd.Flags().String("source", "", "Source Key Vault URL (e.g. https://src-vault.vault.azure.net)")
	cmd.Flags().String("destination", "", "Destination Key Vault URL (e.g. https://dst-vault.vault.azure.net)")
	cmd.MarkFlagRequired("source")
	cmd.MarkFlagRequired("destination")

	viper.BindPFlag("azure.migrate.source", cmd.Flags().Lookup("source"))
	viper.BindPFlag("azure.migrate.destination", cmd.Flags().Lookup("destination"))

	return cmd
}


func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export Key Vault secrets and certificates",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			outputPath := viper.GetString("output")
			fmt.Println(outputPath)
			subscriptionID, _ := cmd.Flags().GetString("subscription")
			vaultName, _ := cmd.Flags().GetString("name")
			client, err := azureUtils.NewAzureClient(subscriptionID)
			if err != nil {
				return fmt.Errorf("failed to create Azure client: %w", err)
			}
			secrets, err := client.ExportSecrets(ctx, vaultName)

			if err != nil {
				return fmt.Errorf("failed to export secrets: %w", err)
			}
			for _, secret := range secrets {
				fmt.Printf("Name: %s, Value: %s\n", secret.Name, secret.Value)
			}
			
			if outputPath != "" {
				azureUtils.WriteToJSONFile(secrets, outputPath)
			}		
			return nil
		},
	}

	cmd.Flags().StringP("name", "n", "", "Name of the vault")
	cmd.Flags().StringP("output", "o", "secrets.json", "Output file path")
	cmd.MarkFlagRequired("name")
	viper.BindPFlag("azure.export.name", AzureCmd.PersistentFlags().Lookup("name"))
	viper.BindPFlag("azure.export.output", AzureCmd.PersistentFlags().Lookup("output"))

	return cmd
}
