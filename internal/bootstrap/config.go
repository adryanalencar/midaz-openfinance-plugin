package bootstrap

const ApplicationName = "plugin-openfinance"

// Config encapsulates all environment configurations for the Open Finance Plugin.
type Config struct {
	EnvName                     string `env:"ENV_NAME" default:"development"`
	LogLevel                    string `env:"LOG_LEVEL" default:"info"`
	ServerAddress               string `env:"SERVER_ADDRESS" default:"0.0.0.0:8085"`
	PluginHost                  string `env:"PLUGIN_HOST" default:"http://localhost:8085"`

	// Provider selection: 'openfinance' or 'pluggy'
	Provider                    string `env:"PROVIDER" default:"openfinance"`

	// Sync scheduling
	SyncIntervalMinutes         int    `env:"SYNC_INTERVAL_MINUTES" default:"60"`
	LookbackWindowHours         int    `env:"LOOKBACK_WINDOW_HOURS" default:"24"`

	// Midaz Core Settings
	MidazOnboardingURL          string `env:"MIDAZ_ONBOARDING_URL" default:"http://localhost:3000"`
	MidazTransactionURL         string `env:"MIDAZ_TRANSACTION_URL" default:"http://localhost:3001"`
	MidazOrganizationID         string `env:"MIDAZ_ORGANIZATION_ID"`
	MidazLedgerID               string `env:"MIDAZ_LEDGER_ID"`
	MidazClearingAccountAlias   string `env:"MIDAZ_CLEARING_ACCOUNT_ALIAS" default:"@external/BRL:openfinance"`
	MidazAuthToken              string `env:"MIDAZ_AUTH_TOKEN" default:""`

	// Direct Open Finance Settings
	OpenFinanceEnvironment      string `env:"OPENFINANCE_ENVIRONMENT" default:"sandbox"`
	OpenFinanceDirectoryURL     string `env:"OPENFINANCE_DIRECTORY_URL" default:"https://matls-auth.sandbox.directory.openbankingbrasil.org.br"`
	OpenFinanceClientID         string `env:"OPENFINANCE_CLIENT_ID" default:""`
	OpenFinanceTransportCert    string `env:"OPENFINANCE_TRANSPORT_CERT_PATH" default:""`
	OpenFinanceTransportKey     string `env:"OPENFINANCE_TRANSPORT_KEY_PATH" default:""`
	OpenFinanceSigningCert      string `env:"OPENFINANCE_SIGNING_CERT_PATH" default:""`
	OpenFinanceSigningKey       string `env:"OPENFINANCE_SIGNING_KEY_PATH" default:""`

	// Pluggy Aggregator Settings
	PluggyBaseURL               string `env:"PLUGGY_BASE_URL" default:"https://api.pluggy.ai"`
	PluggyClientID              string `env:"PLUGGY_CLIENT_ID" default:""`
	PluggyClientSecret          string `env:"PLUGGY_CLIENT_SECRET" default:""`

	// Telemetry & Logging
	OtelServiceName             string `env:"OTEL_RESOURCE_SERVICE_NAME" default:"plugin-openfinance"`
	OtelLibraryName             string `env:"OTEL_LIBRARY_NAME" default:"midaz-plugin-openfinance"`
	OtelServiceVersion          string `env:"OTEL_RESOURCE_SERVICE_VERSION" default:"v1.0.0"`
	OtelDeploymentEnv           string `env:"OTEL_RESOURCE_DEPLOYMENT_ENVIRONMENT" default:"development"`
	OtelColExporterEndpoint     string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" default:""`
	EnableTelemetry             bool   `env:"ENABLE_TELEMETRY" default:"false"`
}
