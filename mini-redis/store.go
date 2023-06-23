package main

import (
	"fmt"
	"mini-redis/datastructure"
	"strconv"
	"sync"
	"time"
)

type Store struct {
	values sync.Map
	locks  sync.Map
	timers sync.Map
}

type UnlockCallback func()
type Value interface{}

type Operations interface {
	LockKey(key string) UnlockCallback
	Set(key string, value Value) (bool, error)
	SetEx(key string, value Value, seconds int) (bool, error)
	Get(key string) (string, bool, error)
	Del(keys ...string) int
	DbSize() int
	Incr(key string) (int, error)
}

func (store *Store) LockKey(key string) UnlockCallback {
	// LoadOrStore returns the existing value for the key if present.
	// Otherwise, it stores and returns the given value.
	// The loaded result is true if the value was loaded, false if stored
	actual, _ := store.locks.LoadOrStore(key, new(sync.Mutex))
	// typecasting to actual mutex
	mutex := actual.(*sync.Mutex)
	// Lock locks key k.
	// If the lock is already in use, the calling goroutine blocks until the mutex is available.
	mutex.Lock()
	// return a callback function for the caller to call
	return func() {
		mutex.Unlock()
	}
}

func (store *Store) Set(key string, value Value) (bool, error) {
	return store.SetEx(key, value, -1)
}

func (store *Store) SetEx(key string, value Value, seconds int) (bool, error) {
	// how is this approach better than simply acquiring a lock in this fn itself
	unlock := store.LockKey(key)
	defer unlock()

	store.clearTtlTimer(key)
	store.values.Store(key, value)

	if seconds > -1 {
		store.setTtlTimer(key, seconds)
	}

	return true, nil
}

func (store *Store) setTtlTimer(key string, seconds int) {
	duration := time.Second * time.Duration(seconds)
	timer := time.AfterFunc(duration, func() {
		store.del(key)
	})

	store.timers.Store(key, timer)
}

func (store *Store) clearTtlTimer(key string) {
	if actual, ok := store.timers.Load(key); ok {
		timer := actual.(*time.Timer)
		timer.Stop()

		store.timers.Delete(key)
	}
}

func (store *Store) Get(key string) (string, bool, error) {
	unlock := store.LockKey(key)
	defer unlock()

	if actual, ok := store.values.Load(key); ok {
		switch typed := actual.(type) {
		case string:
			return typed, true, nil
		case int:
			return strconv.Itoa(typed), true, nil
		default:
			return "", false, fmt.Errorf("miniredis: cant return %v of type %T as string", typed, typed)
		}
	}

	return "", false, nil
}

func (store *Store) Del(keys ...string) int {
	count := 0
	for _, key := range keys {
		if store.del(key) {
			count++
		}
	}

	return count
}

func (store *Store) del(key string) bool {
	unlock := store.LockKey(key)
	defer unlock()

	if _, ok := store.values.Load(key); ok {
		store.clearTtlTimer(key)
		store.values.Delete(key)

		return true
	}

	return false
}

func (store *Store) DbSize() int {
	count := 0
	store.values.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

func (store *Store) Incr(key string) (int, error) {
	unlock := store.LockKey(key)
	defer unlock()

	actual, _ := store.values.LoadOrStore(key, 0)

	var num int
	switch typed := actual.(type) {
	case int:
		num = typed
	case string:
		value, err := strconv.Atoi(typed)
		if err != nil {
			return 0, fmt.Errorf("miniredis: conversion of %q to integer failed with message %q", typed, err)
		}
		num = value
	default:
		return 0, fmt.Errorf("miniredis: cant convert value %q to integer", typed)
	}
	num++
	store.values.Store(key, num)

	return num, nil
}

func (store *Store) ZAdd(key string, sets ...datastructure.SortedSetItem) (int, error) {
	unlock := store.LockKey(key)
	defer unlock()

	actual, _ := store.values.LoadOrStore(key, datastructure.MakeSortedSet())

	var sortedSet *datastructure.SortedSet
	switch typed := actual.(type) {
	case *datastructure.SortedSet:
		sortedSet = typed
	default:
		return 0, fmt.Errorf("miniredis: key %q value is not a sorted set: %q", key, typed)
	}

	count := 0
	for _, set := range sets {
		if sortedSet.Set(set.Score, set.Member) {
			count++
		}
	}
	return count, nil
}
