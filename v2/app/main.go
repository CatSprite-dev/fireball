package main

import (
	"embed"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/CatSprite-dev/fireball/v2/internal/load"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
	"github.com/CatSprite-dev/fireball/v2/internal/ui"
)

const baseCurrency = "RUB"

// Фронтенд из frontend/dist встраивается в бинарник.
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Временно, до связки ключей: токен из v2/.env.
	if _, err := os.Stat("../.env"); err == nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Fatalf("load ../.env: %v", err)
		}
	}
	token := os.Getenv("T_INVEST_TOKEN_ALL")
	if token == "" {
		log.Fatal("T_INVEST_TOKEN_ALL is not set: put it in v2/.env or the environment")
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	client := tinvest.New(token, logger)
	loader := load.New(client, baseCurrency, logger)

	app := application.New(application.Options{
		Name:        "Fireball Outcome",
		Description: "Честный учёт инвестиций для клиентов Т-Инвестиций",
		Services: []application.Service{
			application.NewService(ui.NewPortfolioService(loader, baseCurrency, logger)),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Fireball Outcome",
		Width:  1000,
		Height: 618,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(6, 7, 15),
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
