package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"

	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

const baseCurrency = "RUB"

// Типы счетов, которые программа не показывает.
var excludedAccountTypes = map[string]bool{
	"ACCOUNT_TYPE_INVEST_BOX": true,
	"ACCOUNT_TYPE_DFA":        true,
}

type accountData struct {
	Account    tinvest.Account
	Operations []tinvest.Operation
	Portfolio  *tinvest.Portfolio
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	if err := run(logger); err != nil {
		logger.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return fmt.Errorf("load .env: %w", err)
		}
	}

	token := os.Getenv("T_INVEST_TOKEN_ALL")
	if token == "" {
		return errors.New("T_INVEST_TOKEN_ALL environment variable is not set")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client := tinvest.New(token, logger)

	accounts, err := client.GetAccounts(ctx)
	if err != nil {
		return fmt.Errorf("get accounts: %w", err)
	}

	accounts = filterAccounts(accounts)
	if len(accounts) == 0 {
		return errors.New("no supported accounts")
	}

	start := time.Now()
	data, err := loadAccounts(ctx, client, accounts)
	if err != nil {
		return fmt.Errorf("load accounts: %w", err)
	}
	logger.Info("all accounts loaded", "accounts", len(data), "duration", time.Since(start))

	for _, d := range data {
		logger.Info("account",
			"name", d.Account.Name,
			"operations", len(d.Operations),
			"total", d.Portfolio.TotalAmountPortfolio,
		)
	}
	return nil
}

// filterAccounts оставляет только счета, которые программа показывает.
func filterAccounts(accounts []tinvest.Account) []tinvest.Account {
	var supported []tinvest.Account
	for _, acc := range accounts {
		if excludedAccountTypes[acc.Type] {
			continue
		}
		supported = append(supported, acc)
	}
	return supported
}

// loadAccounts загружает операции и портфель всех счетов параллельно.
// Если хоть один счёт не загрузился, возвращает ошибку: неполная сводка хуже честной ошибки.
func loadAccounts(ctx context.Context, client *tinvest.Client, accounts []tinvest.Account) ([]accountData, error) {
	results := make([]accountData, len(accounts))

	g, gctx := errgroup.WithContext(ctx)
	for i, acc := range accounts {
		g.Go(func() error {
			ops, err := client.GetOperationsByCursor(gctx, tinvest.OperationsFilter{AccountID: acc.ID})
			if err != nil {
				return fmt.Errorf("account %s: %w", acc.ID, err)
			}
			portfolio, err := client.GetPortfolio(gctx, acc.ID, baseCurrency)
			if err != nil {
				return fmt.Errorf("account %s: %w", acc.ID, err)
			}
			results[i] = accountData{Account: acc, Operations: ops, Portfolio: portfolio}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}
