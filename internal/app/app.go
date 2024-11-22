package app

import (
	"neawn-backend/internal/app/handlers"

	"github.com/gin-gonic/gin"
)

type App struct {
	router *gin.Engine
}

func New() *App {
	app := &App{
		router: gin.Default(),
	}
	app.setupRoutes()
	return app
}

func (a *App) setupRoutes() {
	offerHandler := handlers.NewOfferHandler()

	api := a.router.Group("/api")
	{
		api.GET("/offers", offerHandler.GetOffers)
		api.POST("/offers", offerHandler.CreateOffers)
		api.DELETE("/offers", offerHandler.CleanupData)
	}
}

func (a *App) Run() error {
	return a.router.Run(":80")
}
