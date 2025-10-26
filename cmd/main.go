package main

import (
	"log"

	todo "github.com/Toasterinee/Todo-app"
	"github.com/Toasterinee/Todo-app/pkg/handler"
	"github.com/Toasterinee/Todo-app/pkg/repository"
	"github.com/Toasterinee/Todo-app/pkg/service"
)

func main() {
	repos := repository.NewRepository()
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	srv := new(todo.Server)
	if err := srv.Run("8000", handlers.InitRoutes()); err != nil {
		log.Fatalf("error occured while running server: %s", err.Error())
	}
}
