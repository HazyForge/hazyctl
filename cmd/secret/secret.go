package secret

import (
	"github.com/hazyforge/hazyctl/cmd/secret/azure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SecretProvider interface {
	Migrate(sourceVaultURL, destinationVaultURL string) error
}

var SecretCmd = &cobra.Command{
	Use:   "secret",
	Short: "secrets utilities",
	Long: `secrets utilities
    hazyctl secret [provider] [command]

	example:
	1. migrate secrets from one vault to another on azure key vault
		hazyctl secret azure migrate --source vault1 --destination vault2 -s 1234567890
	2. export secrets to a local file
		hazyctl secret azure export --vault vault1 --output secrets.json
	`,
}

func init() {
	SecretCmd.PersistentFlags().StringP("provider", "p", "", "the provider to use")
	viper.BindPFlag("secret.provider", SecretCmd.PersistentFlags().Lookup("provider"))
	SecretCmd.AddCommand(azure.AzureCmd)
}


