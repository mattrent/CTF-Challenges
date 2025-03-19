package config

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

var Values = initConfig()

type Annotations map[string]string

func (a *Annotations) Decode(value string) error {
	return json.Unmarshal([]byte(value), a)
}

type Config struct {
	DbHost                   string
	DbPort                   int
	DbUser                   string
	DbPassword               string
	DbName                   string
	DbConn                   string
	JwtSecret                []byte
	UploadPath               string
	MinVMMemory              string
	MaxVMMemory              string
	VMCPUs                   uint32
	VMSSHPUBLICKEY           string
	ChallengeLifetimeMinutes int
	TestLifetimeMinutes      int
	BackendUrl               string
	ChallengeDomain          string
	VMImageUrl               string
	ContainerImageUrl        string
	AdminPassword            string
	CTFDURL                  string
	CTFDAPIToken             string
	IngressClassName         string
	IngressHttpAnnotations   Annotations
	JwksUrl                  string
	RootCert                 string
	MaxConcurrentTests       int
	ImagePullSecret          string
	Unleash                  UnleashConfig
	ChallengeReadinessProbe  KubernetesProbeConfig
	ChallengeLivenessProbe   KubernetesProbeConfig
	// Currently not supported
	ChallengeStartupProbe KubernetesProbeConfig
}

type KubernetesProbeConfig struct {
	InitialDelaySeconds int32
	PeriodSeconds       int32
	TimeoutSeconds      int32
	FailureThreshold    int32
}

type UnleashConfig struct {
	Url         string
	ApiKey      string
	Environment string
}

func initConfig() Config {
	_ = godotenv.Load()

	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal("Error loading config")
	}

	if len(cfg.DbConn) == 0 {
		cfg.DbConn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.DbUser, cfg.DbPassword, cfg.DbHost, cfg.DbPort, cfg.DbName)
	}

	return cfg
}
