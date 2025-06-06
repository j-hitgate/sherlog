package models

type LogMsg struct {
	Level    byte
	Entity   string
	EntityID string
	Modules  []string
	Message  []any
	Attrs    IAttrs
}
