package sherlog

import (
	"fmt"
	"log"
	"os"
	"time"

	m "github.com/j-hitgate/sherlog/models"
)

type Trace struct {
	traceChain []string
	modules    []string
	entity     string
	entityID   string

	queue       chan *m.LogMsg
	isClosed    bool
	groups      []string
	lastLogTime time.Time
}

func NewTrace(name string) *Trace {
	if name == "" {
		panic("no name trace")
	}

	var queue chan *m.LogMsg

	if _config.SyncPrint {
		queue = make(chan *m.LogMsg)
	} else {
		queue = make(chan *m.LogMsg, 10)
	}

	t := &Trace{
		traceChain:  []string{name},
		modules:     []string{},
		groups:      []string{},
		lastLogTime: time.Now(),
		queue:       queue,
	}
	go t.logHandler()
	return t
}

func contains[T int | string](arr []T, item T) bool {
	for i := range arr {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// Getters

func (t *Trace) Modules() []string {
	modules := make([]string, len(t.modules))
	copy(modules, t.modules)
	return modules
}

func (t *Trace) TraceChain() []string {
	traceChain := make([]string, len(t.traceChain))
	copy(traceChain, t.traceChain)
	return traceChain
}

func (t *Trace) Entity() string {
	return t.entity
}

func (t *Trace) EntityID() string {
	return t.entityID
}

// Setters

func (t *Trace) SetEntity(name, id string) {
	t.entity = name
	t.entityID = id
}

// Operations

func (*Trace) filter(l *m.Log) bool {
	f := _filter

	if len(f.Entities) > 0 {
		if contains(f.Entities, l.Entity) {
			return true
		}
	}

	if len(f.Traces) > 0 {
		for _, trace := range l.Traces {
			if contains(f.Traces, trace) {
				return true
			}
		}
	}

	if len(f.Modules) > 0 {
		for _, module := range l.Modules {
			if contains(f.Modules, module) {
				return true
			}
		}
	}

	if len(f.Labels) > 0 {
		for _, label := range l.Labels {
			if contains(f.Labels, label) {
				return true
			}
		}
	}

	if len(f.Fields) > 0 {
		for key, val1 := range l.Fields {
			if val2, ok := f.Fields[key]; ok && val1 == val2 {
				return true
			}
		}
	}

	return false
}

func (t *Trace) logHandler() {
	for lm := range t.queue {
		now := time.Now()

		l := &m.Log{
			Datetime:  now,
			Timestamp: now.UnixMilli(),
			Traces:    t.traceChain,
			Level:     lm.Level,
			Entity:    lm.Entity,
			EntityID:  lm.EntityID,
			Message:   fmt.Sprint(lm.Message...),
			Modules:   lm.Modules,
			Labels:    Labels{},
			Fields:    Fields{},
		}

		if lm.Attrs != nil {
			attrs := lm.Attrs.AsAttrs().(*Attrs)

			if len(attrs.Labels) > 0 {
				l.Labels = attrs.Labels
			}
			if len(attrs.Fields) > 0 {
				l.Fields = attrs.Fields
			}
		}

		if _filter == nil || _filter.Invert == t.filter(l) {
			if _logsSaver != nil {
				_logsSaver.Add(l)
			}

			if _printer != nil {
				if _config.ShortMode {
					_printer.PrintLogShort(l)
				} else {
					_printer.PrintLog(l)
				}
			}
		}

		if lm.Level == 7 {
			Close()
			os.Exit(1)
		}
	}
}

func (t *Trace) AddModule(group, module string) (popModule func()) {
	if t == nil {
		return func() {}
	}

	if module == "" {
		panic("Module empty")
	}

	isNewGroup := false

	if group != "" && (len(t.groups) == 0 || t.groups[len(t.groups)-1] != group) {
		t.modules = append(t.modules, group)
		t.groups = append(t.groups, group)
		isNewGroup = true
	}
	t.modules = append(t.modules, module)

	isCalled := false

	return func() {
		if isCalled {
			return
		}

		if isNewGroup {
			t.modules = t.modules[:len(t.modules)-2]
			t.groups = t.groups[:len(t.groups)-1]
		} else {
			t.modules = t.modules[:len(t.modules)-1]
		}

		isCalled = true
	}
}

func (t *Trace) WithModule(group, module string, fn func()) {
	popModule := t.AddModule(group, module)
	fn()
	popModule()
}

func (t *Trace) Fork(traceName string) *Trace {
	if t == nil {
		return nil
	}

	traceChain := make([]string, len(t.traceChain)+1)
	copy(traceChain, t.traceChain)
	traceChain[len(traceChain)-1] = traceName

	modules := make([]string, len(t.modules))
	copy(modules, t.modules)

	groups := make([]string, len(t.groups))
	copy(groups, t.groups)

	var queue chan *m.LogMsg

	if _config.SyncPrint {
		queue = make(chan *m.LogMsg)
	} else {
		queue = make(chan *m.LogMsg, 10)
	}

	t_ := &Trace{
		traceChain: traceChain,
		modules:    modules,
		entity:     t.entity,
		entityID:   t.entityID,

		queue:       queue,
		groups:      groups,
		lastLogTime: t.lastLogTime,
	}
	go t_.logHandler()
	return t_
}

func (t *Trace) ForkOnMap(keys ...string) map[string]*Trace {
	m := make(map[string]*Trace, len(keys))

	for i := range keys {
		m[keys[i]] = t.Fork(keys[i])
	}
	return m
}

func (t *Trace) Close() {
	if t != nil && !t.isClosed {
		t.isClosed = true
		close(t.queue)
	}
}

// Log

func (t *Trace) sendLog(level byte, attrs m.IAttrs, message []any) {
	if t == nil || level < _config.Level {
		return
	}

	modules := make([]string, len(t.modules))
	copy(modules, t.modules)

	t.queue <- &m.LogMsg{
		Level:    level,
		Entity:   t.entity,
		EntityID: t.entityID,
		Modules:  modules,
		Message:  message,
		Attrs:    attrs,
	}
}

func (t *Trace) FATAL(attrs m.IAttrs, message ...any) {
	if t != nil {
		t.sendLog(7, attrs, message)
	} else {
		log.Fatalln(message...)
	}
}

func (t *Trace) ERROR(attrs m.IAttrs, message ...any) {
	t.sendLog(6, attrs, message)
}

func (t *Trace) WARN(attrs m.IAttrs, message ...any) {
	t.sendLog(5, attrs, message)
}

func (t *Trace) INFO(attrs m.IAttrs, message ...any) {
	t.sendLog(4, attrs, message)
}

func (t *Trace) NOTE(attrs m.IAttrs, message ...any) {
	t.sendLog(3, attrs, message)
}

func (t *Trace) STAGE(attrs m.IAttrs, message ...any) {
	t.sendLog(2, attrs, message)
}

func (t *Trace) DEBUG(attrs m.IAttrs, message ...any) {
	t.sendLog(1, attrs, message)
}

func (t *Trace) MICRO(attrs m.IAttrs, message ...any) {
	t.sendLog(0, attrs, message)
}

// Special log

func (t *Trace) FATAL_if_err(err error, attrs m.IAttrs, message ...any) {
	if err == nil {
		return
	}

	if t != nil {
		t.sendLog(7, attrs, message)
	} else {
		log.Fatalln(message...)
	}
}

func (t *Trace) ERROR_if_err(err error, attrs m.IAttrs, message ...any) {
	if err != nil {
		t.sendLog(6, attrs, message)
	}
}

func (t *Trace) WARN_if_err(err error, attrs m.IAttrs, message ...any) {
	if err != nil {
		t.sendLog(5, attrs, message)
	}
}

func (t *Trace) INFO_if_not_err(err error, attrs m.IAttrs, message ...any) {
	if err == nil {
		t.sendLog(4, attrs, message)
	}
}

func (t *Trace) NOTE_if_err(err error, attrs m.IAttrs, message ...any) {
	if err != nil {
		t.sendLog(3, attrs, message)
	}
}

func (t *Trace) STAGE_if_not_err(err error, attrs m.IAttrs, message ...any) {
	if err == nil {
		t.sendLog(2, attrs, message)
	}
}

func (t *Trace) DEBUG_if_not_err(err error, attrs m.IAttrs, message ...any) {
	if err == nil {
		t.sendLog(1, attrs, message)
	}
}
