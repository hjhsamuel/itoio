package app

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hjhsamuel/itoio/config"
	"github.com/hjhsamuel/itoio/internal/server/core"
	"github.com/hjhsamuel/itoio/internal/server/handler"
	"github.com/sirupsen/logrus"
)

func Start(path string) error {
	if err := config.Init(path); err != nil {
		return err
	}

	conf := config.Get()
	c, err := core.NewCore(conf.Server.Node)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Start(conf); err != nil {
		return err
	}

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = logrus.StandardLogger().Writer()
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	if err := handler.Routes(c, r); err != nil {
		return err
	}
	httpServer := &http.Server{
		Addr:    net.JoinHostPort("0.0.0.0", strconv.Itoa(conf.Server.Port)),
		Handler: r,
	}

	go func() {
		if conf.Server.TLSCertFile != "" && conf.Server.TLSKeyFile != "" {
			httpServer.TLSConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
					tls.TLS_FALLBACK_SCSV,
				},
			}
			_ = httpServer.ListenAndServeTLS(conf.Server.TLSCertFile, conf.Server.TLSKeyFile)
		} else {
			_ = httpServer.ListenAndServe()
		}
	}()

	logrus.Infof("Server started, listening on %s", httpServer.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down the server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = httpServer.Shutdown(ctx); err != nil {
		logrus.Warnf("Force shutdown server: %v", err)
	}
	return nil
}
