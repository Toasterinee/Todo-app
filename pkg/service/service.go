package service

import (
	todo "github.com/Toasterinee/Todo-app"
	"github.com/Toasterinee/Todo-app/pkg/repository"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type Authorization interface {
	CreateUser(user todo.User) (int, error)
	GenerateToken(username, password string) (string, error)
	ParseToken(token string) (int, error)
}

type ToDoList interface {
	Create(userId int, list todo.ToDoList) (int, error)
	GetAll(userId int) ([]todo.ToDoList, error)
	GetById(userId, listId int) (todo.ToDoList, error)
	Delete(userId, listId int) error
	Update(userId, listId int, input todo.UpdateListInput) error
}

type ToDoItem interface {
	Create(userId, listId int, item todo.ToDoItem) (int, error)
	GetAll(userId, listId int) ([]todo.ToDoItem, error)
	GetById(userId, listId, itemId int) (todo.ToDoItem, error)
	Update(userId, listId, itemId int, input todo.UpdateItemInput) error
	Delete(userId, listId, itemId int) error
}

type Service struct {
	Authorization
	ToDoList
	ToDoItem
	redisClient *redis.Client
}

func NewService(repos *repository.Repository, redisClient *redis.Client) *Service {
	logrus.Infof("NewService: repos.ToDoList = %v", repos.ToDoList)
	return &Service{
		Authorization: NewAuthService(repos.Authorization),
		ToDoList:      NewToDoListService(repos.ToDoList, redisClient),
		ToDoItem:      NewToDoItemService(repos.ToDoItem, repos.ToDoList, redisClient),
		redisClient:   redisClient,
	}
}
