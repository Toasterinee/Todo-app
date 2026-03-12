package repository

import (
	todo "github.com/Toasterinee/Todo-app"
	"github.com/jmoiron/sqlx"
)

type Authorization interface {
	CreateUser(user todo.User) (int, error)
	GetUser(username, password string) (todo.User, error)
}

type ToDoList interface {
	Create(userId int, list todo.ToDoList) (int, error)
	GetAll(userId int) ([]todo.ToDoList, error)
	GetById(userId, listId int) (todo.ToDoList, error)
	Delete(userId, listId int) error
	Update(userId, listId int, input todo.UpdateListInput) error
}

type ToDoItem interface {
	Create(listId int, item todo.ToDoItem) (int, error)
	GetAll(userId, listId int) ([]todo.ToDoItem, error)
	GetById(userId, listId, itemId int) (todo.ToDoItem, error)
	Update(userId, listId, itemId int, input todo.UpdateItemInput) error
	Delete(userId, listId, itemId int) error
}

type Repository struct {
	Authorization
	ToDoList
	ToDoItem
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
		ToDoList:      NewToDoListPostgres(db),
		ToDoItem:      NewTodoItemPostgres(db),
	}
}
