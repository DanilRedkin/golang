package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/project/library/config"
	migrations "github.com/project/library/db"
	generated "github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/internal/usecase/outbox"
	"github.com/project/library/internal/usecase/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	exitDelay             = 3 * time.Second
	dialTimeout           = 30 * time.Second
	keepAlivePeriod       = 180 * time.Second
	maxIdleConns          = 100
	maxConnsPerHost       = 100
	idleConnTimeout       = 90 * time.Second
	tlsHandshakeTimeout   = 15 * time.Second
	expectContinueTimeout = 2 * time.Second
)

func Run(logger *zap.Logger, cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.PG.URL)

	if err != nil {
		logger.Error("can not create pgxpool", zap.Error(err))
		return
	}

	defer dbPool.Close()

	if err := migrations.SetupPostgres(dbPool, logger); err != nil {
		logger.Error("failed to set up Postgres", zap.Error(err))
		return
	}

	repo := repository.NewPostgresRepository(dbPool)
	outboxRepository := repository.NewOutbox(dbPool)

	transactor := repository.NewTransactor(dbPool)
	logger.Info("outbox endpoints",
		zap.String("authorURL", cfg.Outbox.AuthorSendURL),
		zap.String("bookURL", cfg.Outbox.BookSendURL),
	)

	runOutbox(ctx, cfg, logger, outboxRepository, transactor)

	useCases := library.New(logger, repo, repo, outboxRepository, transactor)

	ctrl := controller.New(logger, useCases, useCases)

	go runRest(ctx, cfg, logger)
	go runGrpc(cfg, logger, ctrl)

	<-ctx.Done()
	time.Sleep(exitDelay)
}

func runOutbox(
	ctx context.Context,
	cfg *config.Config,
	logger *zap.Logger,
	outboxRepository repository.OutboxRepository,
	transactor repository.Transactor,
) {
	dialer := &net.Dialer{
		Timeout:   dialTimeout,
		KeepAlive: keepAlivePeriod,
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          maxIdleConns,
		MaxConnsPerHost:       maxConnsPerHost,
		IdleConnTimeout:       idleConnTimeout,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		ExpectContinueTimeout: expectContinueTimeout,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}

	client := new(http.Client)
	client.Transport = transport

	globalHandler := globalOutboxHandler(client, cfg.Outbox.BookSendURL, cfg.Outbox.AuthorSendURL, logger)
	outboxService := outbox.New(logger, outboxRepository, globalHandler, cfg, transactor)

	outboxService.Start(
		ctx,
		cfg.Outbox.Workers,
		cfg.Outbox.BatchSize,
		cfg.Outbox.WaitTimeMS,
		cfg.Outbox.InProgressTTLMS,
	)
}

func globalOutboxHandler(
	client *http.Client,
	bookURL, authorURL string,
	logger *zap.Logger,
) outbox.GlobalHandler {
	return func(kind repository.OutboxKind) (outbox.KindHandler, error) {
		switch kind {
		case repository.OutboxKindBook:
			return bookOutboxHandler(logger, client, bookURL), nil
		case repository.OutboxKindAuthor:
			return authorOutboxHandler(logger, client, authorURL), nil
		default:
			return nil, fmt.Errorf("unsupported outbox kind: %d", kind)
		}
	}
}

func bookOutboxHandler(
	logger *zap.Logger,
	client *http.Client,
	url string) outbox.KindHandler {
	return func(_ context.Context, data []byte) error {
		var book entity.Book
		if err := json.Unmarshal(data, &book); err != nil {
			logger.Error("book unmarshalling failure", zap.Error(err))
			return fmt.Errorf("cannot deserialize book outbox data: %w", err)
		}

		response, err := client.Post(url, "text/plain", strings.NewReader(book.ID))
		if err != nil {
			logger.Error("POST book request failed", zap.String("url", url), zap.Error(err))
			return fmt.Errorf("POST to book endpoint: %w", err)
		}
		defer func() {
			if respErr := response.Body.Close(); respErr != nil {
				logger.Error("failed to close response body", zap.Error(respErr))
			}
		}()

		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			err := fmt.Errorf("book endpoint returned status %d", response.StatusCode)
			logger.Error("book delivery failed", zap.Int("status", response.StatusCode), zap.Error(err))
			return err
		}

		logger.Info("successfully delivered book",
			zap.String("bookID", book.ID),
			zap.String("endpoint", url))

		return nil
	}
}

func authorOutboxHandler(
	logger *zap.Logger,
	client *http.Client,
	url string,
) outbox.KindHandler {
	return func(_ context.Context, data []byte) error {
		var author entity.Author
		if err := json.Unmarshal(data, &author); err != nil {
			logger.Error("author unmarshalling failure", zap.Error(err))
			return fmt.Errorf("cannot deserialize author outbox data: %w", err)
		}

		response, err := client.Post(url, "text/plain", strings.NewReader(author.ID))
		if err != nil {
			logger.Error("POST author request failed", zap.String("url", url), zap.Error(err))
			return fmt.Errorf("failed to POST author ID to %q: %w", url, err)
		}
		defer func() {
			if respErr := response.Body.Close(); respErr != nil {
				logger.Error("failed to close response", zap.Error(respErr))
			}
		}()

		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			err := fmt.Errorf("author endpoint returned status %d", response.StatusCode)
			logger.Error("author delivery failed", zap.Int("status", response.StatusCode), zap.Error(err))
			return err
		}

		logger.Info("successfully delivered author",
			zap.String("authorID", author.ID),
			zap.String("endpoint", url))

		return nil
	}
}

func runRest(ctx context.Context, cfg *config.Config, logger *zap.Logger) {
	mux := grpcruntime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	address := "localhost:" + cfg.GRPC.Port
	err := generated.RegisterLibraryHandlerFromEndpoint(ctx, mux, address, opts)
	if err != nil {
		logger.Error("cannot register grpc gateway", zap.Error(err))
		return
	}

	gatewayPort := ":" + cfg.GRPC.GatewayPort
	logger.Info("gateway listening at port", zap.String("port", gatewayPort))

	if err = http.ListenAndServe(gatewayPort, mux); err != nil {
		logger.Error("gateway listen error", zap.Error(err))
	}
}

func runGrpc(cfg *config.Config, logger *zap.Logger, libraryService generated.LibraryServer) {
	port := ":" + cfg.GRPC.Port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error("cannot open tcp socket", zap.Error(err))
		return
	}

	s := grpc.NewServer()
	reflection.Register(s)
	generated.RegisterLibraryServer(s, libraryService)

	logger.Info("grpc server listening at port", zap.String("port", port))

	if err = s.Serve(lis); err != nil {
		logger.Error("grpc server listen error", zap.Error(err))
	}
}
