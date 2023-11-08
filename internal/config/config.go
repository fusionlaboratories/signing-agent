package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

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
	QredoAPI string `yaml:"qredoAPI" json:"qredoAPI"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	CertFile string `yaml:"certFile" json:"certFile"`
	KeyFile  string `yaml:"keyFile" json:"keyFile"`
}

type AutoApprove struct {
	Enabled          bool `yaml:"enabled" json:"enabled"`
	RetryIntervalMax int  `yaml:"retryIntervalMaxSec" json:"retryIntervalMaxSec"`
	RetryInterval    int  `yaml:"retryIntervalSec" json:"retryIntervalSec"`
}

type WebSocketConfig struct {
	QredoWebsocket    string `yaml:"qredoWebsocket" json:"qredoWebsocket"`
	ReconnectTimeOut  int    `yaml:"reconnectTimeoutSec" json:"reconnectTimeoutSec"`
	ReconnectInterval int    `yaml:"reconnectIntervalSec" json:"reconnectIntervalSec"`
	PingPeriod        int    `yaml:"pingPeriodSec" json:"pingPeriodSec"`
	PongWait          int    `yaml:"pongWaitSec" json:"pongWaitSec"`
	WriteWait         int    `yaml:"writeWaitSec" json:"writeWaitSec"`
	ReadBufferSize    int    `yaml:"readBufferSize" json:"readBufferSize"`
	WriteBufferSize   int    `yaml:"writeBufferSize" json:"writeBufferSize"`
}

type Store struct {
	Type       string    `default:"file" yaml:"type" json:"type"`
	FileConfig string    `yaml:"file" json:"file"`
	OciConfig  OciConfig `yaml:"oci" json:"oci"`
	AwsConfig  AWSConfig `yaml:"aws" json:"aws"`
	GcpConfig  GCPConfig `yaml:"gcp" json:"gcp"`
}

type OciConfig struct {
	Compartment         string `yaml:"compartment" json:"compartment"`
	Vault               string `yaml:"vault" json:"vault"`
	SecretEncryptionKey string `yaml:"secretEncryptionKey" json:"secretEncryptionKey"`
	ConfigSecret        string `yaml:"configSecret" json:"configSecret"`
}

type AWSConfig struct {
	Region     string `yaml:"region" json:"region"`
	SecretName string `yaml:"configSecret" json:"configSecret"`
}

type GCPConfig struct {
	ProjectID  string `yaml:"projectID" json:"projectID"`
	SecretName string `yaml:"configSecret" json:"configSecret"`
}

type HttpSettings struct {
	Addr             string    `yaml:"addr" json:"addr"`
	CORSAllowOrigins []string  `yaml:"CORSAllowOrigins" json:"CORSAllowOrigins"`
	LogAllRequests   bool      `yaml:"logAllRequests" json:"logAllRequests"`
	TLS              TLSConfig `yaml:"TLS" json:"TLS"`
}

type Logging struct {
	Format string `yaml:"format" json:"format"`
	Level  string `yaml:"level" json:"level"`
}

type LoadBalancing struct {
	Enable                bool        `yaml:"enable" json:"enable"`
	OnLockErrorTimeOutMs  int         `yaml:"onLockErrorTimeoutMs" json:"onLockErrorTimeoutMs"`
	ActionIDExpirationSec int         `yaml:"actionIDExpirationSec" json:"actionIDExpirationSec"`
	RedisConfig           RedisConfig `yaml:"redis" json:"redis"`
}

type RedisConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
}

func (c *Config) Default() {
	c.HTTP = HttpSettings{
		Addr:             "127.0.0.1:8007",
		CORSAllowOrigins: []string{"*"},
		TLS: TLSConfig{
			Enabled: false,
		},
	}

	c.Base.QredoAPI = "https://backend.uat.qredo.network/api/v2" //TODO - check if this is the right env. Atm is down
	c.AutoApprove = AutoApprove{
		Enabled:          false,
		RetryIntervalMax: 300,
		RetryInterval:    5,
	}
	c.Websocket = WebSocketConfig{
		ReconnectTimeOut:  300,
		ReconnectInterval: 5,
		QredoWebsocket:    "wss://backend.uat.qredo.network/api/v2/actions/signrequests", //TODO - check if this is the right env. Atm is down
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
