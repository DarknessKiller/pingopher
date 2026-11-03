package app

import (
	"log"

	"github.com/DarknessKiller/pingopher/internal/bootstrap"
	"github.com/DarknessKiller/pingopher/internal/config"
	"github.com/DarknessKiller/pingopher/internal/router"
	"github.com/DarknessKiller/pingopher/internal/service/uptime"
	"github.com/gin-gonic/gin"
)

type App struct {
	engine          *gin.Engine
	config          *config.Config
	uptimeScheduler *uptime.Scheduler
}

func New() *App {
	app := &App{}

	bs := bootstrap.New()
	engine := router.Router(bs)
	app.engine = engine
	app.config = bs.Config
	app.uptimeScheduler = bs.UptimeScheduler

	return app
}

func (a *App) Run() {
	go a.uptimeScheduler.Start()
	if err := a.engine.Run(a.config.Host + ":" + a.config.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
