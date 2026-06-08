package main

import (
	"fmt"
	"log"
	"net/http"

	"gox/pkg/core"
)

type App struct {
	UserController      *UserController      `inject:""`
	NotificationService *NotificationService `inject:""`
}

func (a *App) Start() {
	// Initialize NotificationService to listen to events
	a.NotificationService.Init()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", core.Handler(a.UserController.Create))

	fmt.Println("Server started on :8081 with Auto-Wiring!")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
