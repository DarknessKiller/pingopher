package app

import (
	"github.com/DarknessKiller/pingopher/internal/bootstrap"
	"github.com/DarknessKiller/pingopher/internal/config"
	"github.com/DarknessKiller/pingopher/internal/router"
	"github.com/gin-gonic/gin"
)

type App struct {
	engine *gin.Engine
	config *config.Config
}

func New() *App {
	app := &App{}

	bs := bootstrap.New()
	engine := router.Router(bs)
	app.engine = engine
	app.config = bs.Config

	return app
}

func (a *App) Run() {
	a.engine.Run(a.config.Host + ":" + a.config.Port)
}
