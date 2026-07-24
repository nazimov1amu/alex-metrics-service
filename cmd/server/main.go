package main

import (
	"github.com/Alexunder2003/alex-metrics-service/internal/app"
)

func main() {
	app := app.NewApp()
	app.Run()
}
