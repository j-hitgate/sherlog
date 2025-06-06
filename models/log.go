package models

import "time"

type Log struct {
	Datetime  time.Time         `json:"-"`
	Timestamp int64             `json:"timestamp"`
	Level     byte              `json:"level"`
	Traces    []string          `json:"traces"`
	Modules   []string          `json:"modules"`
	Entity    string            `json:"entity"`
	EntityID  string            `json:"entity_id"`
	Message   string            `json:"message"`
	Labels    []string          `json:"labels"`
	Fields    map[string]string `json:"fields"`
}
