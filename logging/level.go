package logging

type Level string

const (
	LevelCritical Level = "critical"
	LevelError    Level = "error"
	LevelWarning  Level = "warning"
	LevelNotice   Level = "notice"
)

func (l Level) IsSevere() bool {
	return l == LevelCritical || l == LevelError || l == LevelWarning
}
