package config

import (
	"os"

	"github.com/hjhsamuel/goenv"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

const (
	envPrefix = "ITOIO"
)

const (
	TurnModeTurn string = "turn" // both turn and stun
	TurnModeStun string = "stun" // stun only
)

type Config struct {
	Server ServerConfig `yaml:"server" env:"SERVER"`
	Turn   TurnConfig   `yaml:"turn" env:"TURN"`
	Log    LogConfig    `yaml:"log" env:"LOG"`
}

type ServerConfig struct {
	Node        int64  `yaml:"node" env:"NODE"`
	TLSCertFile string `yaml:"tls_cert_file" env:"TLSCERTFILE"`
	TLSKeyFile  string `yaml:"tls_key_file" env:"TLSKEYFILE"`
	Port        int    `yaml:"port" env:"PORT"` // http server port
	StoragePath string `yaml:"storage_path" env:"STORAGEPATH"`
}

type TurnConfig struct {
	Port     int    `yaml:"port" env:"PORT"` // eg. 3478
	PublicIP string `yaml:"public_ip" env:"PUBLICIP"`
	Realm    string `yaml:"realm" env:"REALM"`
	Mode     string `yaml:"mode" env:"MODE"`
}

type LogConfig struct {
	Level string `yaml:"level" env:"LEVEL"`
	Path  string `yaml:"path" env:"PATH"`
}

var gConf = &Config{
	Server: ServerConfig{
		Node:        1,
		Port:        5001,
		StoragePath: "data/ito.db",
	},
	Turn: TurnConfig{
		Port:     15432,
		Realm:    "ito-webrtc",
		Mode:     TurnModeTurn,
		PublicIP: "172.20.10.2",
	},
	Log: LogConfig{
		Level: "info",
		Path:  "log/ito.log",
	},
}

func Get() *Config {
	return gConf
}

func Init(path string) error {
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err = yaml.Unmarshal(content, gConf); err != nil {
			return err
		}
	} else {
		parser := goenv.NewEnvParser(goenv.WithPrefix(envPrefix))
		if err = parser.Start(&gConf); err != nil {
			return err
		}
	}

	if err := initLog(); err != nil {
		return err
	}
	return nil
}

func initLog() error {
	if gConf.Log.Path != "" {
		logrus.SetOutput(&lumberjack.Logger{
			Filename:   gConf.Log.Path,
			MaxSize:    50,
			MaxBackups: 3,
			LocalTime:  true,
			Compress:   true,
		})
	}

	f := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logrus.SetFormatter(f)

	ll, err := logrus.ParseLevel(gConf.Log.Level)
	if err != nil {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(ll)
	}
	return nil
}
