package relays

import (
	"bytes"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/j-hitgate/sherlog/agents"
	m "github.com/j-hitgate/sherlog/models"
)

type LogsSaver struct {
	logs   *agents.Deque[[]byte]
	queue  chan *m.Log
	wg     *sync.WaitGroup
	config *m.Config
}

func NewLogsSaver(config *m.Config) *LogsSaver {
	ls := &LogsSaver{
		logs:   &agents.Deque[[]byte]{},
		queue:  make(chan *m.Log, 10),
		wg:     &sync.WaitGroup{},
		config: config,
	}
	ls.wg.Add(1)
	go ls.logHandler()

	return ls
}

func (*LogsSaver) appendFile(name string, data []byte) error {
	dir := filepath.Dir(name)
	err := os.MkdirAll(dir, 0755)

	if err != nil {
		return err
	}

	file, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	_, err = file.Write(data)
	file.Close()
	return err
}

func (ls *LogsSaver) Add(l *m.Log) {
	ls.queue <- l
}

func (ls *LogsSaver) Close() {
	ls.queue <- nil
	ls.wg.Wait()
}

func (ls *LogsSaver) DumpLogs() {
	if ls.logs.Len() == 0 {
		return
	}

	ls.logs.AppendLeft([]byte{})
	logs := ls.logs.ToSlice()

	name := time.Now().Format("02.01.2006.log")

	err := ls.appendFile(
		path.Join(ls.config.LogsDir, name),
		bytes.Join(logs, []byte{'\n'}),
	)

	if err != nil {
		panic(err)
	}

	ls.logs.Clear()
}

func (ls *LogsSaver) logHandler() {
	for l := range ls.queue {
		if l == nil {
			break
		}

		data, err := json.Marshal(l)

		if err != nil {
			panic(err)
		}

		ls.logs.Append(data)

		if ls.config.AutodumpAfter > 0 && ls.logs.Len() >= ls.config.AutodumpAfter {
			ls.DumpLogs()
		}
	}

	if ls.config.AutodumpAfter > 0 {
		ls.DumpLogs()
	}
	ls.wg.Done()
}
