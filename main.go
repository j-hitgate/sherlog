package sherlog

import (
	"github.com/j-hitgate/sherlog/agents"
	m "github.com/j-hitgate/sherlog/models"
	"github.com/j-hitgate/sherlog/relays"
)

var _config *m.Config
var _filter *Filter
var _printer *agents.Printer
var _logsSaver *relays.LogsSaver
var _isInit = false

func Init(config Config, filter *Filter) {
	if _isInit {
		return
	}
	cfg := config
	_config = (*m.Config)(&cfg)
	_filter = filter

	if !config.NotShowLogs {
		_printer = agents.NewPrinter(_config)
	}
	if config.LogsDir != "" {
		_logsSaver = relays.NewLogsSaver(_config)
	}
	_isInit = true
}

func Close() {
	if _logsSaver != nil {
		_logsSaver.Close()
	}
}

func DumpLogs() {
	_logsSaver.DumpLogs()
}

func CloseTraces(traces map[string]*Trace) {
	for _, t := range traces {
		t.Close()
	}
}
