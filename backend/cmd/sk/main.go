package main

import (
	"context"
	"flag"
	"github.com/BronOS/secret-keeper/internal/app/apiserver"
	"github.com/BronOS/secret-keeper/internal/pkg/db"
	"github.com/BronOS/secret-keeper/internal/pkg/passwords"
	"github.com/BronOS/secret-keeper/internal/pkg/security"
	"github.com/BronOS/secret-keeper/internal/pkg/uid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// dependencies (di container)
var (
	configPath string
	config     *Config
	logger     *logrus.Logger
	router     *mux.Router
	storage    db.Interface
	startedAt  time.Time
	pg         passwords.GeneratorInterface
	kg         uid.GeneratorInterface
	cipher     security.CipherInterface
)

func init() {
	flag.StringVar(&configPath, "config-path", "/etc/sk/config.yaml", "path to config file")
}

func main() {
	flag.Parse()

	config = readConfig(configPath)
	setDeps(config)
	setMiddlewares(router)
	setRoutes(router)

	srv := &http.Server{
		Addr:    config.Server.Bind,
		Handler: router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Printf("Listening on addr: %s", config.Server.Bind)

	<-done
	log.Print("Graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		if err := storage.Disconnect(); err != nil {
			log.Printf("Failed to close DB: %s\n", err)
		}

		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server stopped")
}

func setDeps(c *Config) {
	startedAt = time.Now()

	// logging
	logger = logrus.New()
	level, err := logrus.ParseLevel(c.Logger.Level)
	if err != nil {
		log.Fatal(err)
	}
	logger.SetLevel(level)

	// routing
	router = mux.NewRouter()

	// storage
	storage = db.NewMongoDB(&db.Config{
		Addr: c.Database.Addr,
		Port: c.Database.Port,
		User: c.Database.User,
		Pass: c.Database.Pass,
		Name: c.Database.Name,
	})
	if err := storage.Connect(); err != nil {
		log.Fatalf("Failed to connect to database:%+v", err)
	}

	// password generator
	pg = passwords.NewGenerator(&passwords.Config{
		Length:     c.PasswordGenerator.Length,
		NumDigits:  c.PasswordGenerator.NumDigits,
		NumSymbols: c.PasswordGenerator.NumSymbols,
	})

	// key generator
	kg = uid.NewGenerator()

	// cipher security
	cipher = security.NewAes(config.Security.CipherKey)
}

func readConfig(configPath string) *Config {
	config := &Config{}
	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}
	if err := yaml.Unmarshal(b, config); err != nil {
		log.Fatalf("failed to decode config file: %v", err)
	}

	return config
}

func setRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/ping", apiserver.HandlePing(startedAt)).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/secret/create", apiserver.HandleCreate(logger, storage, pg, kg, cipher)).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/secret/info", apiserver.HandleInfo(logger, storage, config.Security.MaxPinTries)).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/secret/view", apiserver.HandleView(logger, storage, config.Security.MaxPinTries, cipher)).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/secret/delete", apiserver.HandleDelete(logger, storage)).Methods(http.MethodDelete)
}

func setMiddlewares(router *mux.Router) {
	router.Use(apiserver.LoggingMiddleware(logger))
}
