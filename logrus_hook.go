package wattpilot

import (
	log "github.com/sirupsen/logrus"
)

type CallHook struct {
	Call      func(string, string)
	LogLevels []log.Level
}

func (hook *CallHook) Fire(entry *log.Entry) error {
	line, err := entry.Bytes()
	if err != nil {
		return err
	}
	hook.Call(entry.Level.String(), string(line))
	return nil
}

// Levels define on which log levels this hook would trigger
func (hook *CallHook) Levels() []log.Level {
	return hook.LogLevels
}
