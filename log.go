// Copyright 2019 Luca Stasio <joshuagame@gmail.com>
// Copyright 2019 IT Resources s.r.l.
//
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package sgul defines common structures and functionalities for applications.
// log.go defines commons for application logging.
package sgul

import (
	"io"
	"path"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// Logger represent the minimal set of func to set a logger
type Logger interface {
	SetOutput(w io.Writer)
	Print(i ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})
	Debug(i ...interface{})
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	Info(i ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Warn(i ...interface{})
	Warnf(format string, args ...interface{})
	Warnln(args ...interface{})
	Error(i ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	Fatal(i ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Panic(i ...interface{})
	Panicf(format string, args ...interface{})
}

var loggerInstance *logrus.Logger

// GetLogger returns the logger instance. Iff the instance is nil, then the instance will be initialized.
func GetLogger() Logger {
	conf := GetConfiguration().Log
	if loggerInstance == nil {
		loggerInstance = logrus.New()
		if IsSet("log") {
			// file log with rotation
			rfh, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
				Filename:   path.Join(conf.Path, conf.Filename),
				MaxSize:    conf.MaxSize,
				MaxBackups: conf.MaxBackups,
				MaxAge:     conf.MaxAge,
				Level:      parseLevel(conf),
				Formatter:  logFormatter(conf),
			})

			if err != nil {
				panic(err)
			}

			loggerInstance.AddHook(rfh)

			// console log
			if conf.Console.Enabled {
				loggerInstance.SetLevel(parseLevel(conf))
				loggerInstance.SetOutput(colorable.NewColorableStdout())
				loggerInstance.SetFormatter(consoleFormatter(conf))
			}
		} else {
			// default logger
			Formatter := new(logrus.TextFormatter)
			Formatter.TimestampFormat = "02-01-2006 15:04:05"
			Formatter.FullTimestamp = true
			logrus.SetFormatter(Formatter)
		}
	}

	loggerInstance.Debug("Config and Logger initialized")

	return loggerInstance
}

func parseLevel(conf Log) logrus.Level {
	var logLevel logrus.Level

	logLevel, err := logrus.ParseLevel(conf.Level)
	if err != nil {
		panic(err)
	}

	return logLevel
}

func logFormatter(conf Log) logrus.Formatter {
	if conf.JSON {
		return &logrus.JSONFormatter{
			TimestampFormat: conf.TimestampFormat,
		}
	}

	return &prefixed.TextFormatter{
		DisableColors:   true,
		ForceColors:     false,
		TimestampFormat: conf.TimestampFormat,
		FullTimestamp:   conf.FullTimestamp,
		ForceFormatting: conf.ForceFormatting,
	}
}

func consoleFormatter(conf Log) logrus.Formatter {
	return &prefixed.TextFormatter{
		DisableColors:   conf.Console.DisableColors,
		ForceColors:     conf.Console.Colors,
		TimestampFormat: conf.TimestampFormat,
		FullTimestamp:   conf.FullTimestamp,
		ForceFormatting: conf.ForceFormatting,
	}
}
