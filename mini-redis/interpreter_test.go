package main

import (
	mini_redis_ops "mini-redis-ops"
	"testing"
)

func TestInterpreter_Exec(t *testing.T) {
	store := new(mini_redis_ops.Store)
	intr := Interpreter{store}

	t.Run("set key to store", func(t *testing.T) {
		if actual, err := intr.Exec("SET foo bar"); err == nil {
			if expected := true; actual != expected {
				t.Errorf("expected %v, got %v", expected, actual)
			}
		} else {
			t.Errorf("expected no error, but got %q", err)
		}
	})

	t.Run("get key from store", func(t *testing.T) {
		if actual, err := intr.Exec("GET foo"); err == nil {
			if expected := "bar"; actual != expected {
				t.Errorf("expected %v, got %v", expected, actual)
			}
		} else {
			t.Errorf("expected no error, but got %q", err)
		}
	})

	t.Run("get the store size", func(t *testing.T) {
		if actual, err := intr.Exec("DBSIZE"); err == nil {
			if expected := 1; actual != expected {
				t.Errorf("expected %v, got %v", expected, actual)
			}
		} else {
			t.Errorf("expected no error, but got %q", err)
		}
	})

	t.Run("delete key from store", func(t *testing.T) {
		if actual, err := intr.Exec("DEL foo"); err == nil {
			if expected := 1; actual != expected {
				t.Errorf("expected %v, got %v", expected, actual)
			}
		} else {
			t.Errorf("expected no error, but got %q", err)
		}
	})

	t.Run("increment key from store", func(t *testing.T) {
		intr.Exec("INCR bar")
		if actual, err := intr.Exec("INCR bar"); err == nil {
			if expected := 2; actual != expected {
				t.Errorf("expected %v, got %v", expected, actual)
			}
		} else {
			t.Errorf("expected no error, but got %q", err)
		}
	})

	t.Run("set sorted set item to key", func(t *testing.T) {
		if actual, err := intr.Exec("ZADD fizz 3 three"); err == nil {
			if expected := 1; actual != expected {
				t.Errorf("expected %v, got %v", expected, actual)
			}
		} else {
			t.Errorf("expected no error, but got %q", err)
		}
	})
}
