package sherlog

type Config struct {
	Level     byte
	SyncPrint bool

	NotShowLogs     bool
	NotShowDatetime bool
	ShowDate        bool
	ShowTimeDelta   bool
	NotShowLevel    bool
	NotShowTraces   bool
	NotShowModules  bool
	NotShowEntity   bool
	NotShowEntityId bool
	NotShowLabels   bool
	NotShowFields   bool
	ShortMode       bool

	LogsDir       string
	AutodumpAfter int
}
