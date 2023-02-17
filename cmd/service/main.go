package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/jessevdk/go-flags"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/rest"
	"github.com/qredo/signing-agent/rest/version"
	"github.com/qredo/signing-agent/util"
)

var (
	buildType    = ""
	buildVersion = ""
	buildDate    = ""
)

func startText() {
	fmt.Printf("Signing Agent service %v (%v) build date: %v\n\n", buildType, buildVersion, buildDate)
}

type versionCmd struct{}

func (c *versionCmd) Execute([]string) error {
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

	ver := version.DefaultVersion()
	if len(buildType) > 0 {
		ver.BuildType = buildType
	}
	if len(buildVersion) > 0 {
		ver.BuildVersion = buildVersion
	}
	if len(buildDate) > 0 {
		ver.BuildDate = buildDate
	}

	router, err := rest.NewQRouter(log, &cfg, ver)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	setCtrlC(router)

	if err = router.Start(); err != nil {
		log.Error("HTTP Listener error", "err", err)
	}

	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	return nil
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

func main() {
	startText()

	var parser = flags.NewParser(nil, flags.Default)

	_, _ = parser.AddCommand("init", "init config", "write default config", &initCmd{})
	_, _ = parser.AddCommand("start", "start service", "", &startCmd{})
	_, _ = parser.AddCommand("version", "print version", "print service version and quit", &versionCmd{})

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
}

func setCtrlC(router *rest.Router) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		router.Stop()
		time.Sleep(2 * time.Second) //wait for everything to close properly
		os.Exit(0)
	}()
}
