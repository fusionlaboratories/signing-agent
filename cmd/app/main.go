package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"

	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/qredo/signing-agent/internal/action"
	"github.com/qredo/signing-agent/internal/api"
	"github.com/qredo/signing-agent/internal/auth"
	"github.com/qredo/signing-agent/internal/autoapprover"
	"github.com/qredo/signing-agent/internal/config"
	"github.com/qredo/signing-agent/internal/defs"
	"github.com/qredo/signing-agent/internal/hub"
	"github.com/qredo/signing-agent/internal/hub/message"
	"github.com/qredo/signing-agent/internal/rest"
	"github.com/qredo/signing-agent/internal/service"
	"github.com/qredo/signing-agent/internal/store"
	"github.com/qredo/signing-agent/internal/util"
)

var (
	buildType    = ""
	buildVersion = ""
	buildDate    = ""
)

func main() {
	startText()

	var parser = flags.NewParser(nil, flags.Default)

	_, _ = parser.AddCommand("init", "init config", "write default config", &initCmd{})
	_, _ = parser.AddCommand("start", "start service", "", &startCmd{})
	_, _ = parser.AddCommand("version", "print version", "print service version and quit", &versionCmd{})
	_, _ = parser.AddCommand("gen-keys", "generate keys", "generates keys and quit", &genKeysCmd{})

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
}

func startText() {
	fmt.Printf("Signing Agent service %v (%v) build date: %v\n\n", buildType, buildVersion, buildDate)
}

type versionCmd struct{}

func (c *versionCmd) Execute([]string) error {
	return nil
}

type genKeysCmd struct {
}

func (g *genKeysCmd) Execute([]string) error {
	blsPublic, blsPriv, ecPubS, ecPrivS, err := defs.GenerateKeys()
	if err != nil {
		return err
	}

	fmt.Printf("BlsPublic: %s\nBlsPrivate: %s\nEcPublic: %s\nEcPrivate: %s\n", blsPublic, blsPriv, ecPubS, ecPrivS)
	return nil
}

type startCmd struct {
	ConfigFile string `short:"c" long:"config" description:"path to configuration file" default:"cc.yaml"`
}

func (c *startCmd) Execute([]string) error {
	var cfg config.Config
	cfg.Default()

	err := cfg.Load(c.ConfigFile)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	log := util.NewLogger(&cfg.Logging)
	log.Info("Loaded config file from " + c.ConfigFile)

	ver := &api.Version{
		BuildType: "dev",
	}

	if len(buildType) > 0 {
		ver.BuildType = buildType
	}
	if len(buildVersion) > 0 {
		ver.BuildVersion = buildVersion
	}
	if len(buildDate) > 0 {
		ver.BuildDate = buildDate
	}

	router, err := initRouter(log, cfg, *ver)
	if err != nil {
		log.Errorf("Failed to start the router, err: %v", err)
		log.Warn("exiting")
		os.Exit(1)
	}

	setCtrlC(router)

	if err = router.Start(); err != nil {
		log.Errorf("HTTP Listener error: %v", err)
		log.Warn("exiting")
		os.Exit(1)
	}

	return nil
}

func setCtrlC(r *rest.Router) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		r.Stop()
		time.Sleep(2 * time.Second) //wait for everything to close properly
		os.Exit(0)
	}()
}

func initRouter(log *zap.SugaredLogger, config config.Config, version api.Version) (*rest.Router, error) {
	agentStore, err := genAgentStore(config, log)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialise the store")
	}

	agentInfo, err := agentStore.GetAgentInfo()
	if err != nil && err != defs.ErrKVNotFound {
		return nil, errors.Wrap(err, "Failed to retrieve registered agent info")
	}

	headerProvider, agentKey, err := genHeaderProvider(config, agentInfo, log)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialise the header provider")
	}

	rds := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.LoadBalancing.RedisConfig.Host, config.LoadBalancing.RedisConfig.Port),
		Password: config.LoadBalancing.RedisConfig.Password,
		DB:       config.LoadBalancing.RedisConfig.DB,
	})

	var messageCache message.Cacher
	if !config.AutoApprove.Enabled {
		messageCache = message.NewCacher(config.LoadBalancing.Enable, log, rds)
	}

	feedHub := hub.NewFeedHub(hub.NewWebsocketSource(hub.NewDefaultDialer(), config.Websocket.QredoWebsocket, log, config.Websocket, headerProvider), log, messageCache)
	signer, err := action.NewSigner(config.Base.QredoAPI, headerProvider, log, agentKey)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialise the signer")
	}

	pool := goredis.NewPool(rds)
	rs := redsync.New(pool)
	syncronizer := action.NewSyncronizer(&config.LoadBalancing, rds, rs)

	autoApprover := genAutoApprover(config, log, signer, syncronizer)
	upgrader := hub.NewDefaultUpgrader(config.Websocket.ReadBufferSize, config.Websocket.WriteBufferSize)

	agentService := service.NewAgentService(config, headerProvider, agentStore, signer, feedHub, autoApprover, log, upgrader, agentInfo)
	actionService := service.NewActionService(syncronizer, log, config.LoadBalancing.Enable, messageCache, signer)

	return rest.NewRouter(log, config, version, agentService, actionService), nil
}

func genAgentStore(config config.Config, log *zap.SugaredLogger) (store.AgentStore, error) {
	log.Infof("Using %s store", config.Store.Type)
	kv := util.CreateStore(config)
	if kv == nil {
		log.Panicf("Unsupported store type: %s", config.Store.Type)
	}

	if err := kv.Init(); err != nil {
		return nil, err
	}

	return store.NewAgentStore(kv), nil
}

func genAutoApprover(config config.Config, log *zap.SugaredLogger, signer action.Signer, syncronizer action.ActionSync) autoapprover.AutoApprover {
	if !config.AutoApprove.Enabled {
		log.Debug("Auto-approval feature not enabled in config")
		return nil
	}

	log.Debug("Auto-approval feature enabled")
	return autoapprover.NewAutoApprover(log, config, syncronizer, signer)
}

func genHeaderProvider(config config.Config, agentInfo *store.AgentInfo, log *zap.SugaredLogger) (auth.HeaderProvider, string, error) {
	provider := auth.NewHeaderProvider(config.Base.QredoAPI, log)
	agentKey := defs.EmptyString

	if agentInfo != nil {
		agentKey = agentInfo.BLSPrivateKey
		if err := provider.Initiate(agentInfo.WorkspaceID, agentInfo.APIKeySecret, agentInfo.APIKeyID); err != nil {
			return nil, agentKey, err
		}
	}

	return provider, agentKey, nil
}

type initCmd struct {
	FileName string `short:"f" long:"file-name" description:"output file name" default:"cc.yaml"`
}

func (c *initCmd) Execute([]string) error {
	var cfg config.Config
	cfg.Default()
	if err := cfg.Save(c.FileName); err != nil {
		return err
	}

	fmt.Printf("written file %s\n\n", c.FileName)
	return nil
}
