package cmd

// Config represents the global configuration options for hazyctl
type Config struct {
	Azure AzureConfig `json:"azure"`
}

type AzureConfig struct {
	Subscription string `json:"subscription"`
	Migrate      MigrateConfig `json:"migrate"`
	Export       ExportConfig `json:"export"`
}

type MigrateConfig struct {
	Source       string `json:"source"`
	Destination  string `json:"destination"`
}

type ExportConfig struct {
	Name string `json:"name"`
	Output string `json:"output"`
}
