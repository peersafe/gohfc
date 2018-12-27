package gohfc

import (
	"github.com/op/go-logging"
	"os"
)

//设置log级别
func SetLogLevel(logMap map[string]string) error {
	format := logging.MustStringFormatter("%{shortfile} %{time:2006-01-02 15:04:05.000} [%{module}] %{level:.4s} : %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	for name, level := range logMap {
		logLevel, err := logging.LogLevel(level)
		if err != nil {
			return err
		}
		logging.SetBackend(backendFormatter).SetLevel(logLevel, name)
		logger.Debugf("SetLogLevel level: %s, levelName: %s\n", level, name)
	}

	return nil
}
