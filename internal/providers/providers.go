package providers

type SecretSource interface {
    ListSecrets(vaultName string) ([]Secret, error)
    GetSecret(vaultName, secretName string) (Secret, error)
}

type SecretDestination interface {
    PutSecret(vaultName string, secret Secret) error
    DeleteSecret(vaultName, secretName string) error
}

type Secret struct {
    Name        string
    Value       string
    Metadata    map[string]string
}