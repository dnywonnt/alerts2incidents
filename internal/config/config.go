package config // dnywonnt.me/alerts2incidents/internal/config

import (
	"errors"
	"time"

	"dnywonnt.me/alerts2incidents/internal/utils"
	"github.com/spf13/viper"
)

// Initialize Viper to automatically read environment variables
func init() {
	viper.AutomaticEnv()
}

// ApiConfig represents the configuration for the API
// ApiConfig contains configuration settings for the API server, including LDAP and JWT configurations.
type ApiConfig struct {
	Host                       string        `validate:"omitempty,hostname|ip"`    // Hostname or IP address for the API server. This field is optional and must contain a valid hostname or IP address if provided.
	Port                       int           `validate:"required,gte=1,lte=65535"` // Port number for the API server. This field is required and the value must be between 1 and 65535.
	LDAP                       *LDAPConfig   `validate:"required"`                 // Configuration settings for connecting to an LDAP server. This field is required.
	JwtSecretKey               string        `validate:"required,base64"`          // Secret key for signing JWT tokens. This field is required and must be a valid base64 encoded string.
	JwtTokenExpirationInterval time.Duration `validate:"required,min=1h,max=24h"`  // Expiration interval for JWT tokens. This field is required and must be between 1 hour and 1 day.
}

// LDAPConfig contains configuration settings for connecting to an LDAP server.
type LDAPConfig struct {
	Host          string   `validate:"required,hostname|ip"`     // Hostname or IP address for the LDAP server. This field is required and must contain a valid hostname or IP address.
	Port          int      `validate:"required,gte=1,lte=65535"` // Port number for the LDAP server. This field is required and the value must be between 1 and 65535.
	BaseDN        string   `validate:"required"`                 // Base Distinguished Name (BaseDN) for LDAP queries. This field is required.
	AllowedGroups []string `validate:"required,min=1"`           // List of allowed groups in LDAP. This field is required and must contain at least one element.
}

// DatabaseConfig represents the configuration for the database connection
type DatabaseConfig struct {
	Host           string `validate:"required,hostname|ip"`     // Hostname or IP for the database
	Port           int    `validate:"required,gte=1,lte=65535"` // Port for the database
	Name           string `validate:"required"`                 // Database name
	User           string `validate:"required"`                 // Database user
	Password       string `validate:"required"`                 // Database password
	MaxConnections int    `validate:"required,gte=1,lte=100"`   // Max number of connections
}

// ServiceConfig represents the configuration for the service
type ServiceConfig struct {
	DataChanMaxSize       int                     `validate:"required,gte=1,lte=100"`  // Max size for data channel
	AlertsChanMaxSize     int                     `validate:"required,gte=1,lte=100"`  // Max size for alerts channel
	IncidentsCacheMaxSize int                     `validate:"required,gte=1,lte=100"`  // Max size for incidents cache
	RulesCacheMaxSize     int                     `validate:"required,gte=-1,lte=100"` // Max size for rules cache
	GrafanaCollector      *GrafanaCollectorConfig `validate:"required"`                // Configuration for Grafana collector
	ZabbixCollector       *ZabbixCollectorConfig  `validate:"required"`                // Configuration for Zabbix collector
	AlertsParser          *AlertsParserConfig     `validate:"required"`                // Configuration for alerts parser
}

// GrafanaCollectorConfig represents the configuration for the Grafana collector
type GrafanaCollectorConfig struct {
	IsActive                bool          `validate:"-"`                                           // Whether the Grafana collector is active
	APIUrl                  string        `validate:"required_with=IsActive|url"`                  // API URL for Grafana
	Token                   string        `validate:"required_with=IsActive"`                      // Token for Grafana API
	IncludePrometheusAlerts bool          `validate:"-"`                                           // Whether to include Prometheus alerts
	PrometheusUIDs          []string      `validate:"required_with=IncludePrometheusAlerts|min=1"` // Prometheus UIDs
	CollectTimeout          time.Duration `validate:"required_with=IsActive|min=5s"`               // Collection timeout duration
}

// ZabbixCollectorConfig represents the configuration for the Zabbix collector
type ZabbixCollectorConfig struct {
	IsActive        bool          `validate:"-"`                                  // Whether the Zabbix collector is active
	APIUrl          string        `validate:"required_with=IsActive|url"`         // API URL for Zabbix
	Token           string        `validate:"required_with=IsActive"`             // Token for Zabbix API
	TriggerMinLevel int           `validate:"required_with=IsActive|gte=1,lte=5"` // Minimum trigger level
	CollectTimeout  time.Duration `validate:"required_with=IsActive|min=5s"`      // Collection timeout duration
}

// AlertsParserConfig defines the configuration for parsing alerts from different sources.
type AlertsParserConfig struct {
	AggregationInterval         time.Duration `validate:"required,min=5s"`                    // Interval for aggregating alerts (min 5s).
	GrafanaAMParseField         string        `validate:"required,oneof=summary description"` // Field to parse from Grafana Alertmanager alerts ("summary" or "description").
	GrafanaPrometheusParseField string        `validate:"required,oneof=summary description"` // Field to parse from Grafana Prometheus alerts ("summary" or "description").
}

// TelegramBotConfig represents the configuration for the Telegram bot
type TelegramBotConfig struct {
	Token                   string        `validate:"required"`                                // Token for the Telegram bot
	Chats                   []string      `validate:"required,min=1"`                          // List of chat IDs and optional thread IDs
	MessageCacheMaxSize     int           `validate:"required,gte=1,lte=100"`                  // Max size for the message cache
	MessageParseMode        string        `validate:"required,oneof=HTML Markdown MarkdownV2"` // Message parse mode
	MessageTemplateFilepath string        `validate:"required,filepath"`                       // Filepath for the message template
	RequestDelay            time.Duration `validate:"required,min=1s,max=30s"`                 // Delay between consecutive requests to avoid rate limits (1 to 30 seconds)
}

// LoadApiConfig loads the API configuration from environment variables
func LoadApiConfig() (*ApiConfig, error) {
	viper.SetEnvPrefix("API")

	ac := &ApiConfig{
		Host: viper.GetString("SERVER_HOST"),
		Port: viper.GetInt("SERVER_PORT"),
		LDAP: &LDAPConfig{
			Host:          viper.GetString("LDAP_HOST"),
			Port:          viper.GetInt("LDAP_PORT"),
			BaseDN:        viper.GetString("LDAP_BASE_DN"),
			AllowedGroups: viper.GetStringSlice("LDAP_ALLOWED_GROUPS"),
		},
		JwtSecretKey:               viper.GetString("JWT_SECRET_KEY"),
		JwtTokenExpirationInterval: viper.GetDuration("JWT_TOKEN_EXPIRATION_INTERVAL"),
	}

	// Validate the configuration
	if err := utils.ValidateStruct(ac); err != nil {
		return nil, err
	}

	return ac, nil
}

// LoadDatabaseConfig loads the database configuration from environment variables
func LoadDatabaseConfig() (*DatabaseConfig, error) {
	viper.SetEnvPrefix("DATABASE")

	dc := &DatabaseConfig{
		Host:           viper.GetString("HOST"),
		Port:           viper.GetInt("PORT"),
		Name:           viper.GetString("NAME"),
		User:           viper.GetString("USER"),
		Password:       viper.GetString("PASSWORD"),
		MaxConnections: viper.GetInt("MAX_CONNECTIONS"),
	}

	// Validate the configuration
	if err := utils.ValidateStruct(dc); err != nil {
		return nil, err
	}

	return dc, nil
}

// LoadServiceConfig loads the service configuration from environment variables
func LoadServiceConfig() (*ServiceConfig, error) {
	viper.SetEnvPrefix("SERVICE")

	sc := &ServiceConfig{
		DataChanMaxSize:       viper.GetInt("CHANNEL_DATA_MAX_SIZE"),
		AlertsChanMaxSize:     viper.GetInt("CHANNEL_ALERTS_MAX_SIZE"),
		IncidentsCacheMaxSize: viper.GetInt("CACHE_INCIDENTS_MAX_SIZE"),
		RulesCacheMaxSize:     viper.GetInt("CACHE_RULES_MAX_SIZE"),
		GrafanaCollector: &GrafanaCollectorConfig{
			IsActive:                viper.GetBool("COLLECTOR_GRAFANA_IS_ACTIVE"),
			APIUrl:                  viper.GetString("COLLECTOR_GRAFANA_API_URL"),
			Token:                   viper.GetString("COLLECTOR_GRAFANA_TOKEN"),
			IncludePrometheusAlerts: viper.GetBool("COLLECTOR_GRAFANA_INCLUDE_PROMETHEUS_ALERTS"),
			PrometheusUIDs:          viper.GetStringSlice("COLLECTOR_GRAFANA_PROMETHEUS_UIDS"),
			CollectTimeout:          viper.GetDuration("COLLECTOR_GRAFANA_COLLECT_TIMEOUT"),
		},
		ZabbixCollector: &ZabbixCollectorConfig{
			IsActive:        viper.GetBool("COLLECTOR_ZABBIX_IS_ACTIVE"),
			APIUrl:          viper.GetString("COLLECTOR_ZABBIX_API_URL"),
			Token:           viper.GetString("COLLECTOR_ZABBIX_TOKEN"),
			TriggerMinLevel: viper.GetInt("COLLECTOR_ZABBIX_TRIGGER_MIN_LEVEL"),
			CollectTimeout:  viper.GetDuration("COLLECTOR_ZABBIX_COLLECT_TIMEOUT"),
		},
		AlertsParser: &AlertsParserConfig{
			AggregationInterval:         viper.GetDuration("PARSER_AGGREGATION_INTERVAL"),
			GrafanaAMParseField:         viper.GetString("PARSER_GRAFANA_AM_PARSE_FIELD"),
			GrafanaPrometheusParseField: viper.GetString("PARSER_GRAFANA_PROMETHEUS_PARSE_FIELD"),
		},
	}

	// Ensure at least one collector is active
	if !sc.GrafanaCollector.IsActive && !sc.ZabbixCollector.IsActive {
		return nil, errors.New("at least one of the collectors must be active")
	}

	// Validate the configuration
	if err := utils.ValidateStruct(sc); err != nil {
		return nil, err
	}

	return sc, nil
}

// LoadTelegramBotConfig loads the Telegram bot configuration from environment variables
func LoadTelegramBotConfig() (*TelegramBotConfig, error) {
	viper.SetEnvPrefix("TELEGRAM")

	tbc := &TelegramBotConfig{
		Token:                   viper.GetString("BOT_TOKEN"),
		Chats:                   viper.GetStringSlice("CHATS"),
		MessageCacheMaxSize:     viper.GetInt("MESSAGE_CACHE_MAX_SIZE"),
		MessageParseMode:        viper.GetString("MESSAGE_PARSE_MODE"),
		MessageTemplateFilepath: viper.GetString("MESSAGE_TEMPLATE_FILEPATH"),
		RequestDelay:            viper.GetDuration("REQUEST_DELAY"),
	}

	// Validate the configuration
	if err := utils.ValidateStruct(tbc); err != nil {
		return nil, err
	}

	return tbc, nil
}
