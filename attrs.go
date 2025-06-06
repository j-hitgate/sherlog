package sherlog

import (
	"fmt"

	m "github.com/j-hitgate/sherlog/models"
)

type Labels []string
type Fields map[string]string

func (f Fields) AsAttrs() m.IAttrs {
	return &Attrs{Fields: f}
}

func WithFields(keyValues ...any) Fields {
	if len(keyValues)%2 == 1 {
		panic("Key dont have value")
	}
	fields := Fields{}

	for i := 0; i < len(keyValues); i += 2 {
		fields[fmt.Sprint(keyValues[i])] = fmt.Sprint(keyValues[i+1])
	}
	return fields
}

func (l Labels) AsAttrs() m.IAttrs {
	return &Attrs{Labels: l}
}

type Attrs struct {
	Labels Labels
	Fields Fields
}

func (a *Attrs) AsAttrs() m.IAttrs {
	return a
}
