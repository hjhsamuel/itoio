package conf

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hjhsamuel/goenv"
)

type Config struct {
	CorsAllowOrigins []string `env:"CORS_ALLOW_ORIGINS"`
	CheckOrigin      func(origin string) bool
}

var Conf = &Config{}

func Init() error {
	parser := goenv.NewEnvParser(
		goenv.WithSplitChar("ITO"),
		goenv.WithSplitChar("|"),
	)

	err := parser.Start(&Conf)
	if err != nil {
		return err
	}

	// compile allowed origins
	var compiledAllowedOrigins []*regexp.Regexp
	for _, item := range Conf.CorsAllowOrigins {
		compile, err := regexp.Compile(item)
		if err != nil {
			return fmt.Errorf("invalid CORS_ALLOW_ORIGINS: %s", err)
		}
		compiledAllowedOrigins = append(compiledAllowedOrigins, compile)
	}
	Conf.CheckOrigin = func(origin string) bool {
		if origin == "" {
			return true
		}
		for _, item := range compiledAllowedOrigins {
			if item.Match([]byte(strings.ToLower(origin))) {
				return true
			}
		}
		return false
	}
}
