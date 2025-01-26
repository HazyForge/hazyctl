package secret

import (
	"hazyctl/cmd/secret/azure"
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
	`,
}

func init() {
	SecretCmd.PersistentFlags().StringP("provider", "p", "", "the provider to use")
	viper.BindPFlag("secret.provider", SecretCmd.PersistentFlags().Lookup("provider"))
	SecretCmd.AddCommand(azure.AzureCmd)
}


