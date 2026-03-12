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

type TodoListService struct {
	repo  repository.ToDoList
	redis *redis.Client
}

func NewToDoListService(repo repository.ToDoList, redis *redis.Client) *TodoListService {
	return &TodoListService{repo: repo, redis: redis}
}

func (s *TodoListService) Create(userId int, list todo.ToDoList) (int, error) {
	id, err := s.repo.Create(userId, list)
	if err != nil {
		return 0, err
	}

	ctx := context.Background()
	allListsCacheKey := fmt.Sprintf("user:%d:lists", userId)
	if err := s.redis.Del(ctx, allListsCacheKey).Err(); err != nil {
		logrus.Warnf("failed to delete cache for user %d lists: %v", userId, err)
	}
	return id, nil
}

func (s *TodoListService) GetAll(userId int) ([]todo.ToDoList, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:lists", userId)

	// Try to get data from Redis cache
	cachedLists, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var lists []todo.ToDoList
		if err := json.Unmarshal(cachedLists, &lists); err == nil {
			return lists, nil
		}
		logrus.Warn("failed to unmarshal cached lists.")
	}
	lists, err := s.repo.GetAll(userId)
	if err != nil {
		return nil, err
	}

	// Cache the result in Redis
	if data, err := json.Marshal(lists); err == nil {
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	} else {
		logrus.Warn("failed to marshal lists for caching.")
	}
	return lists, nil
}

func (s *TodoListService) GetById(userId, listId int) (todo.ToDoList, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:list:%d", userId, listId)

	// Try to get data from Redis cache
	cachedList, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var list todo.ToDoList
		if err := json.Unmarshal(cachedList, &list); err == nil {
			return list, nil
		}
		logrus.Warn("failed to unmarshal cached list.")
	}
	list, err := s.repo.GetById(userId, listId)
	if err != nil {
		return todo.ToDoList{}, err
	}

	// Cache the result in Redis
	if data, err := json.Marshal(list); err == nil {
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	} else {
		logrus.Warn("failed to marshal list for caching.")
	}
	return list, nil
}

func (s *TodoListService) Delete(userId, listId int) error {
	err := s.repo.Delete(userId, listId)
	if err != nil {
		return err
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:list:%d", userId, listId)
	if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
		logrus.Warnf("failed to delete cache for list %d: %v", listId, err)
	}

	allListsCacheKey := fmt.Sprintf("user:%d:lists", userId)
	s.redis.Del(ctx, allListsCacheKey)
	return nil
}

func (s *TodoListService) Update(userId, listId int, input todo.UpdateListInput) error {
	if err := input.Validate(); err != nil {
		return err
	}
	err := s.repo.Update(userId, listId, input)
	if err != nil {
		return err
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:list:%d", userId, listId)
	if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
		logrus.Warnf("failed to delete cache for list %d: %v", listId, err)
	}

	allListsCacheKey := fmt.Sprintf("user:%d:lists", userId)
	s.redis.Del(ctx, allListsCacheKey)
	return nil
}
