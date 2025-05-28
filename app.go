package main

import (
	"github.com/gin-gonic/gin"
	"red-feed/internal/events"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
}
