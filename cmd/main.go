package main

import (
	"log"

	todo "github.com/Toasterinee/Todo-app"
	"github.com/Toasterinee/Todo-app/pkg/handler"
)

func main() {
	handlers := new(handler.Handler)
	srv := new(todo.Server)
	if err := srv.Run("8000", handlers.InitRoutes()); err != nil {
		log.Fatalf("error occured while running server: %s", err.Error())
	}
}
