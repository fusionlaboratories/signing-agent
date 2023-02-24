package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// swagger:model ConfigResponse
type Config struct {
	Base          Base            `yaml:"base" json:"base"`
	HTTP          HttpSettings    `yaml:"http" json:"http"`
	Logging       Logging         `yaml:"logging" json:"logging"`
	LoadBalancing LoadBalancing   `yaml:"loadBalancing" json:"loadBalancing"`
	Store         Store           `yaml:"store" json:"store"`
	AutoApprove   AutoApprove     `yaml:"autoApproval" json:"autoApproval"`
	Websocket     WebSocketConfig `yaml:"websocket" json:"websocket"`
}

type Base struct {
	// The pin number to use to provide a zero knowledge proof token for communication with the Partner API.
	// example: 123456
	PIN int `yaml:"pin" json:"pin"`

	// The URL of the Qredo API.
	// example: https://sandbox-api.qredo.network
	QredoAPI string `yaml:"qredoAPI" json:"qredoAPI"`
}

type TLSConfig struct {
	// Enable TLS for the internal HTTP server.
	// example: true
	Enabled bool `yaml:"enabled" json:"enabled"`

	// The cert file to use for the TLS server.
	// example: tls/domain.crt
	CertFile string `yaml:"certFile" json:"certFile"`

	// The key file to use for the TLS server.
	// example: tls/domain.key
	KeyFile string `yaml:"keyFile" json:"keyFile"`
}

type AutoApprove struct {
	// Activate the automatic approval of every transaction that is received.
	// example: true
	Enabled bool `yaml:"enabled" json:"enabled"`

	// The maximum time over which the Signing Agent retries action approval. After this period, it’s considered as a failure.
	// example: 300
	RetryIntervalMax int `yaml:"retryIntervalMaxSec" json:"retryIntervalMaxSec"`

	// The interval over which the Signing Agent attempts to approve an action. It will cycle this interval until the `retryIntervalMaxSec` is reached.
	// example: 5
	RetryInterval int `yaml:"retryIntervalSec" json:"retryIntervalSec"`
}

type WebSocketConfig struct {
	// The URL of the Qredo WebSocket feed.
	// example: wss://sandbox-api.qredo.network/api/v1/p/coreclient/feed
	QredoWebsocket string `yaml:"qredoWebsocket" json:"qredoWebsocket"`

	// The reconnect timeout in seconds.
	// example: 300
	ReconnectTimeOut int `yaml:"reconnectTimeoutSec" json:"reconnectTimeoutSec"`

	// The reconnect interval in seconds.
	// example: 5
	ReconnectInterval int `yaml:"reconnectIntervalSec" json:"reconnectIntervalSec"`

	// The ping period for the ping handler in seconds.
	// example: 5
	PingPeriod int `yaml:"pingPeriodSec" json:"pingPeriodSec"`

	// The pong wait for the pong handler in seconds.
	// example: 10
	PongWait int `yaml:"pongWaitSec" json:"pongWaitSec"`

	// The write wait in seconds.
	// example: 10
	WriteWait int `yaml:"writeWaitSec" json:"writeWaitSec"`

	// The WebSocket upgrader read buffer size in bytes.
	// example: 512
	ReadBufferSize int `yaml:"readBufferSize" json:"readBufferSize"`

	// The WebSocket upgrader write buffer size in bytes.
	// example: 1024
	WriteBufferSize int `yaml:"writeBufferSize" json:"writeBufferSize"`
}

type Store struct {
	// The type of store to use to store the private key information for the Signing Agent.
	// enum: file,oci,aws
	// example: file
	Type string `default:"file" yaml:"type" json:"type"`

	// The path to the storage file when `file` store is used.
	// example: /volume/ccstore.db
	FileConfig string `yaml:"file" json:"file"`

	OciConfig OciConfig `yaml:"oci" json:"oci"`
	AwsConfig AWSConfig `yaml:"aws" json:"aws"`
}

// OciConfig is the OCI configuration when Base store type is set to oci
type OciConfig struct {
	// The OCID where the vault and encryption key reside.
	// example: ocid1.tenancy.oc1...
	Compartment string `yaml:"compartment" json:"compartment"`

	// The OCID of the vault where the secret will be stored.
	// example: ocid1.vault.oc1...
	Vault string `yaml:"vault" json:"vault"`

	// The encryption key used for both the secret and the data inside the secret.
	// example: ocid1.key.oc1...
	SecretEncryptionKey string `yaml:"secretEncryptionKey" json:"secretEncryptionKey"`

	// The name of secret that will be used to store the data.
	// example: automated_approver_config
	ConfigSecret string `yaml:"configSecret" json:"configSecret"`
}

// AWSConfig is the AWS configuration when Base store type is set to aws
type AWSConfig struct {
	// The AWS region where the secret is stored.
	// example: eu-west-3
	Region string `yaml:"region" json:"region"`

	// The name of the AWS Secrets Manager secret containing the encrypted data.
	// example: secrets_manager_secret
	SecretName string `yaml:"configSecret" json:"configSecret"`
}

type HttpSettings struct {
	// The address and port the service runs on [the bind address and port the build in API endpoints].
	// example: 0.0.0.0:8007
	Addr string `yaml:"addr" json:"addr"`

	// The value of the Access-Control-Allow-Origin of the responses of the build in API.
	// example: '*'
	CORSAllowOrigins []string `yaml:"CORSAllowOrigins" json:"CORSAllowOrigins"`

	// Log all incoming requests to the build in API.
	// example: true
	LogAllRequests bool `yaml:"logAllRequests" json:"logAllRequests"`

	TLS TLSConfig `yaml:"TLS" json:"TLS"`
}

type Logging struct {
	// Output format of the log.
	// enum: text,json
	// example: json
	Format string `yaml:"format" json:"format"`

	// Log level to be output.
	// enum: info,warn,error,debug
	// example: debug
	Level string `yaml:"level" json:"level"`
}

type LoadBalancing struct {
	// Enables the load-balancing logic.
	// example: true
	Enable bool `yaml:"enable" json:"enable"`

	// The on lock timeout in milliseconds.
	// example: 300
	OnLockErrorTimeOutMs int `yaml:"onLockErrorTimeoutMs" json:"onLockErrorTimeoutMs"`

	// The expiration of actionID variable in Redis in seconds.
	// example: 6
	ActionIDExpirationSec int         `yaml:"actionIDExpirationSec" json:"actionIDExpirationSec"`
	RedisConfig           RedisConfig `yaml:"redis" json:"redis"`
}

// RedisConfig is the redis configuration when LoadBalancing is enabled
type RedisConfig struct {
	// The Redis host.
	// example: redis
	Host string `yaml:"host" json:"host"`

	// The Redis port.
	// example: 6379
	Port int `yaml:"port" json:"port"`

	// The Redis password.
	// example: just a password
	Password string `yaml:"password" json:"password"`

	// Redis database to be selected after connecting to the server.
	// example: 0
	DB int `yaml:"db" json:"db"`
}

// Default creates configuration with default values.
func (c *Config) Default() {
	c.HTTP = HttpSettings{
		Addr:             "127.0.0.1:8007",
		CORSAllowOrigins: []string{"*"},
		TLS: TLSConfig{
			Enabled: false,
		},
	}

	c.Base.PIN = 0
	c.Base.QredoAPI = "https://sandbox-api.qredo.network/api/v1/p"
	c.AutoApprove = AutoApprove{
		Enabled:          false,
		RetryIntervalMax: 300,
		RetryInterval:    5,
	}
	c.Websocket = WebSocketConfig{
		ReconnectTimeOut:  300,
		ReconnectInterval: 5,
		QredoWebsocket:    "wss://sandbox-api.qredo.network/api/v1/p/coreclient/feed",
		PingPeriod:        5,
		PongWait:          10,
		WriteWait:         10,
		ReadBufferSize:    512,
		WriteBufferSize:   1024,
	}
	c.Logging.Level = "info"
	c.Logging.Format = "json"
	c.Store.Type = "file"
	c.Store.FileConfig = "ccstore.db"
	c.LoadBalancing = LoadBalancing{
		Enable:                false,
		OnLockErrorTimeOutMs:  300,
		ActionIDExpirationSec: 6,
		RedisConfig: RedisConfig{
			Host:     "redis",
			Port:     6379,
			Password: "",
			DB:       0,
		},
	}
}

// Load reads and parses yaml config.
func (c *Config) Load(fileName string) error {
	f, err := os.ReadFile(fileName)
	if err != nil {
		return errors.Wrap(err, "read config file")
	}

	c.Default()
	if err := yaml.Unmarshal(f, c); err != nil {
		return errors.Wrap(err, "parse config file")
	}

	return nil
}

// Save saves yaml config.
func (c *Config) Save(fileName string) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(fileName, b, 0600)
}
