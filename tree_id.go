package main

import (
	"errors"
	"fmt"
	"strings"
)

var ErrMalformedTreeID = errors.New("malformed tree ID")

type TreeID []string

func NewTreeID(path ...string) TreeID {
	result := []string{}
	for _, p := range path {
		result = append(result, strings.Split(p, "/")...)
	}
	return result
}

func (t TreeID) And(any interface{}) TreeID {
	s := fmt.Sprintf("%v", any)
	newID := make([]string, len(t)+1)
	for i := range t {
		newID[i] = t[i]
	}
	newID[len(t)] = s
	return TreeID(newID)
}

func (t TreeID) String() string {
	return strings.Join(t, "/")
}

func (t TreeID) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *TreeID) UnmarshalText(text []byte) error {
	elements := strings.Split(string(text), "/")
	*t = elements
	return nil
}

func (t *TreeID) Scan(src interface{}) error {
	asString, ok := src.(string)
	if !ok {
		return fmt.Errorf("cannot convert %T into TreeID", src)
	}
	return t.UnmarshalText([]byte(asString))
}
