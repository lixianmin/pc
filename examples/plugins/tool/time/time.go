package time

import (
	"fmt"
	"time"
)

type TimeTool struct{}

func NewTimeTool() *TimeTool {
	return &TimeTool{}
}

func (t *TimeTool) Now() map[string]any {
	now := time.Now()
	return map[string]any{
		"unix":       now.Unix(),
		"rfc3339":    now.Format(time.RFC3339),
		"date":       now.Format("2006-01-02"),
		"time":       now.Format("15:04:05"),
		"weekday":    now.Weekday().String(),
		"timestamp":  now.UnixMilli(),
	}
}

func (t *TimeTool) Format(unix int64, layout string) (string, error) {
	if layout == "" {
		layout = time.RFC3339
	}
	tm := time.Unix(unix, 0)
	return tm.Format(layout), nil
}

func (t *TimeTool) Parse(value, layout string) (map[string]any, error) {
	if layout == "" {
		layout = time.RFC3339
	}
	tm, err := time.Parse(layout, value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time: %w", err)
	}
	return map[string]any{
		"unix":      tm.Unix(),
		"rfc3339":   tm.Format(time.RFC3339),
		"timestamp": tm.UnixMilli(),
	}, nil
}

func (t *TimeTool) Sleep(seconds int) error {
	if seconds <= 0 || seconds > 60 {
		return fmt.Errorf("seconds must be between 1 and 60")
	}
	time.Sleep(time.Duration(seconds) * time.Second)
	return nil
}
