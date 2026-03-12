package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	todo "github.com/Toasterinee/Todo-app"
	"github.com/Toasterinee/Todo-app/pkg/repository"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type TodoItemService struct {
	repo     repository.ToDoItem
	listrepo repository.ToDoList
	redis    *redis.Client
}

func NewToDoItemService(repo repository.ToDoItem, listrepo repository.ToDoList, redis *redis.Client) *TodoItemService {
	logrus.Infof("NewToDoItemService: listrepo = %v", listrepo)
	return &TodoItemService{repo: repo, listrepo: listrepo, redis: redis}
}
func (s *TodoItemService) Create(userId, listId int, item todo.ToDoItem) (int, error) {
	_, err := s.listrepo.GetById(userId, listId)
	if err != nil {
		return 0, err
	}
	id, err := s.repo.Create(listId, item)
	if err != nil {
		return 0, err
	}

	ctx := context.Background()
	allItemsCacheKey := fmt.Sprintf("user:%d:list:%d:items", userId, listId)
	if err := s.redis.Del(ctx, allItemsCacheKey).Err(); err != nil {
		logrus.Warnf("failed to delete cache for user %d list %d items: %v", userId, listId, err)
	}
	return id, nil
}

func (s *TodoItemService) GetAll(userId, listId int) ([]todo.ToDoItem, error) {
	_, err := s.listrepo.GetById(userId, listId)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:list:%d:items", userId, listId)

	// Try to get data from Redis cache
	cachedItems, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var items []todo.ToDoItem
		if err := json.Unmarshal(cachedItems, &items); err == nil {
			return items, nil
		}
		logrus.Warn("failed to unmarshal cached items.")
	}
	items, err := s.repo.GetAll(userId, listId)
	if err != nil {
		return nil, err
	}

	// Cache the result in Redis
	if data, err := json.Marshal(items); err == nil {
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	} else {
		logrus.Warn("failed to marshal items for caching.")
	}
	return items, nil
}

func (s *TodoItemService) GetById(userId, listId, itemId int) (todo.ToDoItem, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:list:%d:item:%d", userId, listId, itemId)

	// Try to get data from Redis cache
	cachedItem, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var item todo.ToDoItem
		if err := json.Unmarshal(cachedItem, &item); err == nil {
			return item, nil
		}
		logrus.Warn("failed to unmarshal cached item.")
	}
	item, err := s.repo.GetById(userId, listId, itemId)
	if err != nil {
		return todo.ToDoItem{}, err
	}

	// Cache the result in Redis
	if data, err := json.Marshal(item); err == nil {
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	} else {
		logrus.Warn("failed to marshal item for caching.")
	}
	return item, nil
}

func (s *TodoItemService) Update(userId, listId, itemId int, input todo.UpdateItemInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	err := s.repo.Update(userId, listId, itemId, input)
	if err != nil {
		return err
	}

	ctx := context.Background()
	itemCacheKey := fmt.Sprintf("user:%d:item:%d", userId, itemId)
	if err := s.redis.Del(ctx, itemCacheKey).Err(); err != nil {
		logrus.Warnf("failed to delete cache for user %d item %d: %v", userId, itemId, err)
	}

	// Invalidate the cache for the list's items
	listItemsCacheKey := fmt.Sprintf("user:%d:list:%d:items", userId, listId)
	if err := s.redis.Del(ctx, listItemsCacheKey).Err(); err != nil {
		logrus.Warnf("failed to delete cache for user %d list items: %v", userId, err)
	}
	return nil
}

func (s *TodoItemService) Delete(userId, listId, itemId int) error {
	err := s.repo.Delete(userId, listId, itemId)
	if err != nil {
		return err
	}

	ctx := context.Background()
	itemCacheKey := fmt.Sprintf("user:%d:list:%d:item:%d", userId, listId, itemId)
	if err := s.redis.Del(ctx, itemCacheKey).Err(); err != nil {
		logrus.Warnf("failed to delete cache for user %d item %d: %v", userId, itemId, err)
	}

	// Invalidate the cache for the list's items
	listItemsCacheKey := fmt.Sprintf("user:%d:list:%d:items", userId, listId)
	if err := s.redis.Del(ctx, listItemsCacheKey).Err(); err != nil {
		logrus.Warnf("failed to delete cache for user %d list items: %v", userId, err)
	}
	return nil
}
