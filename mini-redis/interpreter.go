package main

import (
	"fmt"
	"mini-redis/datastructure"
	"regexp"
	"strconv"
)

var simpleRegex, keyRegex, setRegex, setExRegex, zAddRegex *regexp.Regexp

func init() {
	simpleRegex = regexp.MustCompile("^DBSIZE$")
	keyRegex = regexp.MustCompile("^(?P<cmd>GET|DEL|INCR|ZCARD) (?P<key>[a-zA-Z0-9-_]+)$")
	setRegex = regexp.MustCompile("^SET (?P<key>[a-zA-Z0-9-_]+) (?P<value>[a-zA-Z0-9-_]+)$")
	setExRegex = regexp.MustCompile("^SET (?P<key>[a-zA-Z0-9-_]+) (?P<value>[a-zA-Z0-9-_]+) EX (?P<seconds>[0-9]+)$")
	zAddRegex = regexp.MustCompile("^ZADD (?P<key>[a-zA-Z0-9-_]+) (?P<score>[0-9]+) (?P<member>[a-zA-Z0-9-_]+)$")
}

type Interpreter struct {
	*Store
}

func (intr Interpreter) Exec(cmd string) (interface{}, error) {
	switch {
	case simpleRegex.MatchString(cmd):
		return intr.handleSimpleRegex(cmd)
	case keyRegex.MatchString(cmd):
		return intr.handleKeyRegex(cmd)
	case setRegex.MatchString(cmd):
		return intr.handleSetRegex(cmd)
	case setExRegex.MatchString(cmd):
		return intr.handleSetExRegex(cmd)
	case zAddRegex.MatchString(cmd):
		return intr.handleZAddSetRegex(cmd)
	}
	return errorReturn(cmd)
}

func (intr Interpreter) handleSimpleRegex(cmd string) (interface{}, error) {
	switch cmd {
	case "DBSIZE":
		return intr.DbSize(), nil
	}
	return errorReturn(cmd)
}

func (intr Interpreter) handleKeyRegex(cmd string) (interface{}, error) {
	values := scanVars(keyRegex, cmd, "cmd", "key")
	cmd, key := values[0], values[1]

	switch cmd {
	case "GET":
		if v, ok, err := intr.Get(key); err == nil {
			if ok {
				return v, nil
			} else {
				return nil, err
			}
		}
		return nil, nil
	case "DEL":
		return intr.Del(key), nil
	case "INCR":
		return intr.Incr(key)
	}

	return errorReturn(cmd)
}

func (intr Interpreter) handleSetRegex(cmd string) (interface{}, error) {
	values := scanVars(setRegex, cmd, "key", "value")
	key, value := values[0], values[1]

	return intr.Set(key, value)
}

func (intr Interpreter) handleSetExRegex(cmd string) (interface{}, error) {
	values := scanVars(setExRegex, cmd, "key", "value")
	key, value, secondStr := values[0], values[1], values[2]

	seconds, _ := strconv.Atoi(secondStr)
	return intr.SetEx(key, value, seconds)
}

func (intr *Interpreter) handleZAddSetRegex(cmd string) (interface{}, error) {
	values := scanVars(zAddRegex, cmd, "key", "score", "member")
	key, scoreStr, member := values[0], values[1], values[2]

	score, _ := strconv.Atoi(scoreStr)
	item := datastructure.SortedSetItem{Score: float64(score), Member: member}

	return intr.ZAdd(key, item)
}

func scanVars(regex *regexp.Regexp, str string, keys ...string) []string {
	groupNames := regex.SubexpNames()
	matches := regex.FindStringSubmatch(str)

	values := make([]string, 0)
	for index, value := range matches {
		for _, key := range keys {
			if groupNames[index] == key {
				values = append(values, value)
			}
		}
	}

	return values
}

func errorReturn(cmd string) (interface{}, error) {
	return nil, fmt.Errorf("miniredis: invalid command %q", cmd)
}
