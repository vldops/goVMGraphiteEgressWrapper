package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

var logger *zap.Logger
var err error
var logLevel zapcore.Level
var appInstance string
var environment string
var ok bool
var configPath *string
var victoriaMetricsAddress string

// Config struct
type Config struct {
	Logger struct {
		Level string `yaml:"level"`
	} `yaml:"logger"`
	VictoriaMetrics string `yaml:"victoriaMetrics"`
	AppPort         string `yaml:"appPort"`
	MetricName      string `yaml:"metricName"`
	RgxLevelOne     string `yaml:"rgxLevelOne"`
	RgxLevelTwo     string `yaml:"rgxLevelTwo"`
}

var config Config

func init() {
	configPath = flag.String("config", "", "path2ConfigFile")
	flag.Parse()
	if *configPath == "" {
		log.Fatalf("VictoriaMetrics host must be specified!\n Example: ./myApp --config /path/to/config.yaml")
	}
	parseConfig(configPath)
}

func init() {
	if config.Logger.Level == "" || config.Logger.Level == "info" {
		logLevel = 0
	} else if config.Logger.Level == "debug" {
		logLevel = -1
	} else {
		logLevel = 1
	}

	if config.VictoriaMetrics == "" {
		log.Fatalf("VictoriaMetrics host must be specified!\n Example: VictoriaMetrics: http://VictoriaMetrics:8428")
	} else {
		victoriaMetricsAddress = config.VictoriaMetrics
	}

	if config.AppPort == "" {
		config.AppPort = "8080"
		config.AppPort = fmt.Sprintf(":%v", config.AppPort)
	} else {
		config.AppPort = fmt.Sprintf(":%v", config.AppPort)
	}

	if config.MetricName == "" {
		config.MetricName = "graphite"
	}

	if config.RgxLevelOne == "" {
		config.RgxLevelOne = `graphite{target="(.*)"}`
	}
	if config.RgxLevelTwo == "" {
		config.RgxLevelTwo = `.*[\[\]\{\}\*].*`
	}
}

func init() {
	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(logLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{

			MessageKey:   "message",
			LevelKey:     "level",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			TimeKey:      "time",
			EncodeTime:   zapcore.RFC3339TimeEncoder,
			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
		InitialFields: map[string]interface{}{
			"appName": "goPromxyWrapper",
		},
	}

	logger, err = cfg.Build()

	if err != nil {
		log.Fatal(err)
	}

	logger.Info("Configuring logger was successful.")

}

func parseConfig(configPath *string) {
	f, err := os.Open(*configPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	configEncoder := yaml.NewDecoder(f)
	err = configEncoder.Decode(&config)
	if err != nil {
		log.Fatalf("Can't decode config values, becase: %v\n", err)
		return
	}

}
