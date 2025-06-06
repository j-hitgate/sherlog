package sherlog

type Filter struct {
	Invert   bool
	Traces   []string
	Modules  []string
	Entities []string
	Labels   []string
	Fields   map[string]string
}
