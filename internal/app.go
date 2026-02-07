package internal

import (
	"context"
	"embed"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	command_inbound_adapter "MikrOps/internal/adapter/inbound/command"
	gin_inbound_adapter "MikrOps/internal/adapter/inbound/gin"
	rabbitmq_inbound_adapter "MikrOps/internal/adapter/inbound/rabbitmq"
	crypto_adapter "MikrOps/internal/adapter/outbound/crypto"
	jwt_adapter "MikrOps/internal/adapter/outbound/jwt"
	mikrotik_outbound_adapter "MikrOps/internal/adapter/outbound/mikrotik"
	postgres_outbound_adapter "MikrOps/internal/adapter/outbound/postgres"
	rabbitmq_outbound_adapter "MikrOps/internal/adapter/outbound/rabbitmq"
	redis_outbound_adapter "MikrOps/internal/adapter/outbound/redis"
	"MikrOps/internal/domain"
	_ "MikrOps/internal/migration/postgres"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils"
	"MikrOps/utils/activity"
	"MikrOps/utils/database"
	"MikrOps/utils/encryption"
	"MikrOps/utils/log"
	"MikrOps/utils/rabbitmq"
	"MikrOps/utils/redis"
)

var embedMigrations embed.FS
var databaseDriverList = []string{"postgres"}
var httpDriverList = []string{"gin"}
var messageDriverList = []string{"rabbitmq"}
var outboundDatabaseDriver string
var outboundMessageDriver string
var outboundCacheDriver string
var inboundHttpDriver string
var inboundMessageDriver string

type App struct {
	ctx    context.Context
	domain domain.Domain
}

func NewApp() *App {
	ctx := activity.NewContext(context.Background(), "app_start")
	ctx = activity.WithClientID(ctx, "system")
	godotenv.Load(".env")
	configureLogging()

	outboundDatabaseDriver = os.Getenv("OUTBOUND_DATABASE_DRIVER")
	outboundMessageDriver = os.Getenv("OUTBOUND_MESSAGE_DRIVER")
	outboundCacheDriver = os.Getenv("OUTBOUND_CACHE_DRIVER")
	inboundHttpDriver = os.Getenv("INBOUND_HTTP_DRIVER")
	inboundMessageDriver = os.Getenv("INBOUND_MESSAGE_DRIVER")

	// Initialize encryption service for service credentials
	encryptionSvc := initEncryptionService(ctx)

	// Initialize JWT and password hasher ports
	jwtPort := jwt_adapter.NewJWTAdapter(os.Getenv("JWT_SECRET"))
	passwordHasherPort := crypto_adapter.NewPasswordHasherAdapter()

	domain := domain.NewDomain(
		databaseOutbound(ctx),
		messageOutbound(ctx),
		cacheOutbound(ctx),
		mikrotik_outbound_adapter.NewMikrotikClientFactory(),
		encryptionSvc,
		jwtPort,
		passwordHasherPort,
	)

	return &App{
		ctx:    ctx,
		domain: domain,
	}
}

func (a *App) Run(option string) {
	switch option {
	case "http":
		a.httpInbound()
	case "message":
		a.messageInbound()
	default:
		a.commandInbound()
	}
}

func databaseOutbound(ctx context.Context) outbound_port.DatabasePort {
	if !utils.IsInList(databaseDriverList, outboundDatabaseDriver) {
		log.WithContext(ctx).Fatal("database driver is not supported")
		os.Exit(1)
	}

	// InitDatabase now returns *gorm.DB instead of *sql.DB
	db := database.InitDatabase(ctx, outboundDatabaseDriver)

	switch outboundDatabaseDriver {
	case "postgres":
		return postgres_outbound_adapter.NewAdapter(db)
	}

	return nil
}

func messageOutbound(ctx context.Context) outbound_port.MessagePort {
	if !utils.IsInList(messageDriverList, outboundMessageDriver) {
		log.WithContext(ctx).Fatal("message driver is not supported")
		os.Exit(1)
	}

	switch outboundMessageDriver {
	case "rabbitmq":
		rabbitmq.InitMessage()
		return rabbitmq_outbound_adapter.NewAdapter()
	}

	return nil
}

func cacheOutbound(ctx context.Context) outbound_port.CachePort {
	if !utils.IsInList([]string{"redis"}, outboundCacheDriver) {
		log.WithContext(ctx).Fatal("cache driver is not supported")
		os.Exit(1)
	}

	switch outboundCacheDriver {
	case "redis":
		redis.InitDatabase()
		redis.InitPubsub()
		return redis_outbound_adapter.NewAdapter()
	}

	return nil
}

func (a *App) httpInbound() {
	ctx := a.ctx
	if !utils.IsInList(httpDriverList, inboundHttpDriver) {
		log.WithContext(ctx).Fatal("http driver is not supported")
		os.Exit(1)
	}

	switch inboundHttpDriver {
	case "gin":
		app := gin.New()
		inboundHttpAdapter := gin_inbound_adapter.NewAdapter(a.domain)
		gin_inbound_adapter.InitRoute(ctx, app, inboundHttpAdapter)

		go func() {
			if err := app.Run(":" + os.Getenv("SERVER_PORT")); err != nil {
				log.WithContext(ctx).Fatalf("failed to listen and serve: %+v", err)
			}
		}()
	}

	ctx, shutdown := context.WithTimeout(ctx, 5*time.Second)
	defer shutdown()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)
	<-quit

	log.WithContext(ctx).Info("http server stopped")
}

func (a *App) messageInbound() {
	ctx := a.ctx
	if !utils.IsInList(messageDriverList, inboundMessageDriver) {
		log.WithContext(ctx).Fatal("message driver is not supported")
		os.Exit(1)
	}

	switch inboundMessageDriver {
	case "rabbitmq":
		inboundMessageAdapter := rabbitmq_inbound_adapter.NewAdapter(a.domain, mikrotik_outbound_adapter.NewMikrotikClientFactory())
		rabbitmq_inbound_adapter.InitRoute(ctx, os.Args, inboundMessageAdapter)

		// NOTE: Traffic aggregation uses STREAMING (not polling)
		// Data flows: MikroTik /interface/monitor-traffic (PUSH)
		//   -> Monitor.publishTrafficData()
		//   -> Redis cache (for customer portal)
		// No worker needed - streaming happens on-demand when customers connect
	}
}

func (a *App) commandInbound() {
	ctx := a.ctx
	inboundCommandAdapter := command_inbound_adapter.NewAdapter(a.domain)
	command_inbound_adapter.InitRoute(ctx, os.Args, inboundCommandAdapter)
}

func initEncryptionService(ctx context.Context) *encryption.Service {
	// Get encryption key from environment (must be 32 bytes for AES-256)
	encryptionKey := os.Getenv("SERVICE_CREDENTIAL_KEY")
	if encryptionKey == "" {
		log.WithContext(ctx).Fatal("SERVICE_CREDENTIAL_KEY environment variable is required (must be 32 bytes)")
		os.Exit(1)
	}

	// Initialize encryption service
	encryptionSvc, err := encryption.NewService(encryptionKey)
	if err != nil {
		log.WithContext(ctx).Fatalf("failed to initialize encryption service: %v", err)
		os.Exit(1)
	}

	log.WithContext(ctx).Info("Encryption service initialized successfully")
	return encryptionSvc
}

func configureLogging() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.AddHook(utils.LogrusSourceContextHook{})

	if os.Getenv("APP_MODE") != "release" {
		logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	} else {
		// TODO: Fix joonix import issue
		// logrus.SetFormatter(&joonix.FluentdFormatter{})
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}
