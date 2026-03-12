package repository

import (
	"fmt"
	"strings"

	todo "github.com/Toasterinee/Todo-app"
	"github.com/jmoiron/sqlx"
)

type TodoItemPostgres struct {
	db *sqlx.DB
}

func NewTodoItemPostgres(db *sqlx.DB) *TodoItemPostgres {
	return &TodoItemPostgres{db: db}
}

func (r *TodoItemPostgres) Create(listId int, item todo.ToDoItem) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var itemId int
	createItemQuery := fmt.Sprintf("INSERT INTO %s (title, description) VALUES ($1, $2) RETURNING id", todoItemsTable)
	row := tx.QueryRow(createItemQuery, item.Title, item.Description)
	err = row.Scan(&itemId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	createListItemQuery := fmt.Sprintf("INSERT INTO %s (list_id, item_id) VALUES ($1, $2)", listsItemsTable)
	_, err = tx.Exec(createListItemQuery, listId, itemId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return itemId, tx.Commit()
}

func (r *TodoItemPostgres) GetAll(userId, listId int) ([]todo.ToDoItem, error) {
	var items []todo.ToDoItem

	query := fmt.Sprintf(`
        SELECT ti.id, ti.title, ti.description, ti.done
        FROM %s ti
        INNER JOIN %s li ON li.item_id = ti.id
        INNER JOIN %s ul ON ul.list_id = li.list_id
        WHERE li.list_id = $1 AND ul.user_id = $2`,
		todoItemsTable, listsItemsTable, usersListsTable)

	err := r.db.Select(&items, query, listId, userId)
	return items, err
}

func (r *TodoItemPostgres) GetById(userId, listId, itemId int) (todo.ToDoItem, error) {
	var item todo.ToDoItem

	query := fmt.Sprintf("SELECT ti.id, ti.title, ti.description, ti.done FROM %s ti INNER JOIN %s tl ON ti.id = tl.item_id INNER JOIN %s ul ON ul.list_id = tl.list_id WHERE ti.id = $1 AND ul.user_id = $2 AND ul.list_id = $3", todoItemsTable, listsItemsTable, usersListsTable)
	if err := r.db.Get(&item, query, itemId, userId, listId); err != nil {
		return item, err
	}
	return item, nil
}

func (r *TodoItemPostgres) Update(userId, listId, itemId int, input todo.UpdateItemInput) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if input.Title != nil {
		setValues = append(setValues, fmt.Sprintf("title=$%d", argId))
		args = append(args, *input.Title)
		argId++
	}

	if input.Description != nil {
		setValues = append(setValues, fmt.Sprintf("description=$%d", argId))
		args = append(args, *input.Description)
		argId++
	}

	if input.Done != nil {
		setValues = append(setValues, fmt.Sprintf("done=$%d", argId))
		args = append(args, *input.Done)
		argId++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE %s ti SET %s FROM %s li, %s ul WHERE ti.id = li.item_id AND li.list_id = ul.list_id AND ti.id=$%d AND ul.user_id=$%d",
		todoItemsTable, setQuery, listsItemsTable, usersListsTable, argId, argId+1)
	args = append(args, itemId, userId)

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *TodoItemPostgres) Delete(userId, listId, itemId int) error {
	query := fmt.Sprintf("DELETE FROM %s ti USING %s tl, %s ul WHERE ti.id = tl.item_id AND tl.list_id = ul.list_id AND ti.id = $1 AND ul.user_id = $2", todoItemsTable, listsItemsTable, usersListsTable)
	_, err := r.db.Exec(query, itemId, userId)
	return err
}
