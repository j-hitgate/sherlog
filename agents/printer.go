package agents

import (
	"fmt"
	"strings"
	"time"

	m "github.com/j-hitgate/sherlog/models"
)

var _levelsMap = map[byte]string{
	7: "FATAL",
	6: "ERROR",
	5: "WARN",
	4: "INFO",
	3: "NOTE",
	2: "STAGE",
	1: "DEBUG",
	0: "MICRO",
}

type Printer struct {
	config           *m.Config
	lastLogTimestamp int64
}

func NewPrinter(config *m.Config) *Printer {
	return &Printer{
		config:           config,
		lastLogTimestamp: time.Now().UnixMilli(),
	}
}

func (p *Printer) datetimeAndDeltaBlock(l *m.Log) string {
	if p.config.NotShowDatetime {
		return ""
	}

	parts := make([]string, 0, 3)

	if p.config.ShowDate {
		parts = append(parts, l.Datetime.Format("02.01.2006"))
	}

	parts = append(parts, l.Datetime.Format("15:04:05"))

	if p.config.ShowTimeDelta {
		delta := float64(l.Timestamp-p.lastLogTimestamp) / 1000.0
		parts = append(parts, fmt.Sprintf("+%.3fs", delta))

		p.lastLogTimestamp = l.Timestamp
	}

	return fmt.Sprintf("─[%s]", strings.Join(parts, " "))

}

func (p *Printer) entityAndIdBlock(l *m.Log) string {
	if p.config.NotShowEntity {
		return ""
	}

	parts := make([]string, 0, 2)
	parts = append(parts, l.Entity)

	if !p.config.NotShowEntityId {
		parts = append(parts, l.EntityID)
	}

	return fmt.Sprintf("─( %s )", strings.Join(parts, " : "))
}

func (p *Printer) tracesAndModulesBlock(l *m.Log) string {
	parts := make([]string, len(l.Traces)*2+len(l.Modules)*2)
	separ := " > "
	i := 0

	if !p.config.NotShowTraces {
		for j := 0; j < len(l.Traces); j++ {
			parts[i], parts[i+1] = l.Traces[j], separ
			i += 2
		}
		i--
	}

	if !p.config.NotShowTraces && !p.config.NotShowModules && len(l.Modules) > 0 {
		parts[i] = " >> "
		i++
	}

	if !p.config.NotShowModules && len(l.Modules) > 0 {
		for j := 0; j < len(l.Modules); j++ {
			parts[i], parts[i+1] = l.Modules[j], separ
			i += 2
		}
		i--
	}

	if len(parts) == 0 {
		return ""
	}

	return fmt.Sprintf("─{ %s }", strings.Join(parts[:i], ""))
}

func (p *Printer) lastModuleBlock(l *m.Log) string {
	if p.config.NotShowModules {
		return ""
	}

	return fmt.Sprintf("─{ %s }", l.Modules[len(l.Modules)-1])
}

func (p *Printer) labelsBlock(l *m.Log) string {
	if p.config.NotShowLabels || len(l.Labels) == 0 {
		return ""
	}

	return fmt.Sprintf("─(%s)", strings.Join(l.Labels, ", "))
}

func (p *Printer) levelBlock(l *m.Log) string {
	if p.config.NotShowLevel {
		return ""
	}

	return fmt.Sprintf("─[ %s ]", _levelsMap[l.Level])
}

func (p *Printer) printFields(l *m.Log) {
	if p.config.NotShowFields || len(l.Fields) == 0 {
		return
	}

	for key, val := range l.Fields {
		println(fmt.Sprintf("  ╰─> %s: %s", key, val))
	}
}

func (p *Printer) PrintLog(l *m.Log) {
	// Example:
	// ╭─[18:33:42 +10ms]─( entity : id )─{ trace1 > trace2 >> module1 > module2 }─(label1, label2)
	// ╰─[ INFO ]─> Some message
	//   ╰─> key1: val1
	//   ╰─> key2: val2

	println(strings.Join([]string{"\n╭",
		p.datetimeAndDeltaBlock(l),
		p.entityAndIdBlock(l),
		p.tracesAndModulesBlock(l),
		p.labelsBlock(l),
	}, ""))

	println(strings.Join([]string{"╰", p.levelBlock(l), "─>"}, ""), l.Message)
	p.printFields(l)
}

func (p *Printer) PrintLogShort(l *m.Log) {
	// Example:
	// [18:33:42 +10ms]─[ INFO ]─( entity : id )─{ module2 }─> Some message

	println(strings.Join([]string{
		p.datetimeAndDeltaBlock(l),
		p.levelBlock(l),
		p.entityAndIdBlock(l),
		p.lastModuleBlock(l),
		">",
	}, "─"), l.Message)
}
