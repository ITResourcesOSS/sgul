// Copyright 2019 Luca Stasio <joshuagame@gmail.com>
// Copyright 2019 IT Resources s.r.l.
//
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package sgul defines common structures and functionalities for applications.
// log.go defines commons for application logging.
package sgul

import (
	"os"
	"path"
	"sync"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var onceLogger sync.Once
var logger *zap.Logger

// GetLogger .
func GetLogger() *zap.Logger {
	onceLogger.Do(func() {
		conf := GetConfiguration().Log
		env := os.Getenv("ENV")
		var writerSyncer zapcore.WriteSyncer
		var encoder zapcore.Encoder

		if (env == "prod" || env == "production") && !conf.Console {
			writerSyncer = getLogWriter(conf)
		} else {
			writerSyncer = zapcore.NewMultiWriteSyncer(
				zapcore.AddSync(os.Stdout),
				getLogWriter(conf))
		}

		if conf.JSON {
			encoder = zapcore.NewJSONEncoder(getEncoderConfig())
		} else {
			encoder = zapcore.NewConsoleEncoder(getEncoderConfig())
		}

		lgLvl := zapcore.InfoLevel
		if err := (*zapcore.Level).UnmarshalText(&lgLvl, []byte(conf.Level)); err != nil {
			lgLvl = zapcore.InfoLevel
		}

		core := zapcore.NewCore(encoder, writerSyncer, lgLvl)

		if conf.Caller {
			logger = zap.New(core, zap.AddCaller())
		} else {
			logger = zap.New(core)
		}

	})

	return logger
}

func getEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return encoderConfig
}

func getLogWriter(conf Log) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   path.Join(conf.Path, conf.Filename),
		MaxSize:    conf.MaxSize,
		MaxBackups: conf.MaxBackups,
		MaxAge:     conf.MaxAge,
		Compress:   conf.Compress,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// import (
// 	"context"
// 	"fmt"
// 	"path"
// 	"sync"

// 	"github.com/go-chi/chi/middleware"
// 	"github.com/mattn/go-colorable"
// 	"github.com/sirupsen/logrus"
// 	"github.com/snowzach/rotatefilehook"
// 	prefixed "github.com/x-cray/logrus-prefixed-formatter"
// )

// var onceLogger sync.Once
// var loggerInstance *logrus.Logger

// // Logger is a wrapper around the app logrus logger instance.
// type Logger struct {
// 	logger    *logrus.Logger
// 	module    string
// 	requestID string
// }

// // NewLogger returns a new Logger instance
// func NewLogger(ctx context.Context, l *logrus.Logger, module string) Logger {
// 	requestID := middleware.GetReqID(ctx)

// 	wrapped := l
// 	if wrapped == nil {
// 		wrapped = GetLogger()
// 	}

// 	return Logger{
// 		logger:    wrapped,
// 		module:    module,
// 		requestID: requestID,
// 	}
// }

// // Logger returns the wrapped logger instance.
// func (l Logger) Logger() *logrus.Logger {
// 	return l.logger
// }

// func (l Logger) wrapFormat(format string) string {
// 	return fmt.Sprintf("[%s][%s] - %s", l.requestID, l.module, format)
// }

// func (l Logger) wrapVariadic(args ...interface{}) string {
// 	return fmt.Sprintf("[%s][%s] - %s", l.requestID, l.module, fmt.Sprint(args...))
// }

// // Tracef delegates to the logrous logger Tracef.
// func (l Logger) Tracef(format string, args ...interface{}) {
// 	l.logger.Tracef(l.wrapFormat(format), args...)
// }

// // Debugf delegates to the logrous logger Debugf.
// func (l Logger) Debugf(format string, args ...interface{}) {
// 	l.logger.Debugf(l.wrapFormat(format), args...)
// }

// // Infof delegates to the logrous logger Infof.
// func (l Logger) Infof(format string, args ...interface{}) {
// 	l.logger.Infof(l.wrapFormat(format), args...)
// }

// // Printf delegates to the logrous logger Printf.
// func (l Logger) Printf(format string, args ...interface{}) {
// 	l.logger.Printf(l.wrapFormat(format), args...)
// }

// // Warnf delegates to the logrous logger Warnf.
// func (l Logger) Warnf(format string, args ...interface{}) {
// 	l.logger.Warnf(l.wrapFormat(format), args...)
// }

// // Warningf delegates to the logrous logger Warningf.
// func (l Logger) Warningf(format string, args ...interface{}) {
// 	l.logger.Warningf(l.wrapFormat(format), args...)
// }

// // Errorf delegates to the logrous logger Errorf.
// func (l Logger) Errorf(format string, args ...interface{}) {
// 	l.logger.Errorf(l.wrapFormat(format), args...)
// }

// // Fatalf delegates to the logrous logger Fatalf.
// func (l Logger) Fatalf(format string, args ...interface{}) {
// 	l.logger.Fatalf(l.wrapFormat(format), args...)
// 	l.logger.Exit(1)
// }

// // Panicf delegates to the logrous logger Panicf.
// func (l Logger) Panicf(format string, args ...interface{}) {
// 	l.logger.Panicf(l.wrapFormat(format), args...)
// }

// // Trace delegates to the logrous logger Trace.
// func (l Logger) Trace(args ...interface{}) {
// 	l.logger.Trace(l.wrapVariadic(args...))
// }

// // Debug delegates to the logrous logger Debug.
// func (l Logger) Debug(args ...interface{}) {
// 	l.logger.Debug(l.wrapVariadic(args...))
// }

// // Info delegates to the logrous logger Info.
// func (l Logger) Info(args ...interface{}) {
// 	l.logger.Info(l.wrapVariadic(args...))
// }

// // Print delegates to the logrous logger Print.
// func (l Logger) Print(args ...interface{}) {
// 	l.logger.Print(l.wrapVariadic(args...))
// }

// // Warn delegates to the logrous logger Warn.
// func (l Logger) Warn(args ...interface{}) {
// 	l.logger.Warn(l.wrapVariadic(args...))
// }

// // Warning delegates to the logrous logger Warning.
// func (l Logger) Warning(args ...interface{}) {
// 	l.logger.Warning(l.wrapVariadic(args...))
// }

// // Error delegates to the logrous logger Error.
// func (l Logger) Error(args ...interface{}) {
// 	l.logger.Error(l.wrapVariadic(args...))
// }

// // Fatal delegates to the logrous logger Fatal.
// func (l Logger) Fatal(args ...interface{}) {
// 	l.logger.Fatal(l.wrapVariadic(args...))
// 	l.logger.Exit(1)
// }

// // Panic delegates to the logrous logger Panic.
// func (l Logger) Panic(args ...interface{}) {
// 	l.logger.Panic(l.wrapVariadic(args...))
// }

// // Traceln delegates to the logrous logger Traceln.
// func (l Logger) Traceln(args ...interface{}) {
// 	l.logger.Traceln(l.wrapVariadic(args...))
// }

// // Debugln delegates to the logrous logger Debugln.
// func (l Logger) Debugln(args ...interface{}) {
// 	l.logger.Debugln(l.wrapVariadic(args...))
// }

// // Infoln delegates to the logrous logger Infoln.
// func (l Logger) Infoln(args ...interface{}) {
// 	l.logger.Infoln(l.wrapVariadic(args...))
// }

// // Println delegates to the logrous logger Println.
// func (l Logger) Println(args ...interface{}) {
// 	l.logger.Println(l.wrapVariadic(args...))
// }

// // Warnln delegates to the logrous logger Warnln.
// func (l Logger) Warnln(args ...interface{}) {
// 	l.logger.Warnln(l.wrapVariadic(args...))
// }

// // Warningln delegates to the logrous logger Warningln.
// func (l Logger) Warningln(args ...interface{}) {
// 	l.logger.Warningln(l.wrapVariadic(args...))
// }

// // Errorln delegates to the logrous logger Errorln.
// func (l Logger) Errorln(args ...interface{}) {
// 	l.logger.Errorln(l.wrapVariadic(args...))
// }

// // Fatalln delegates to the logrous logger Fatalln.
// func (l Logger) Fatalln(args ...interface{}) {
// 	l.logger.Fatalln(l.wrapVariadic(args...))
// 	l.logger.Exit(1)
// }

// // Panicln delegates to the logrous logger Panicln.
// func (l Logger) Panicln(args ...interface{}) {
// 	l.logger.Panicln(l.wrapVariadic(args...))
// }

// // GetLogger returns the Logrus logger instance. Iff the instance is nil, then the instance will be initialized.
// func GetLogger() *logrus.Logger {
// 	onceLogger.Do(func() {
// 		conf := GetConfiguration().Log
// 		loggerInstance = logrus.New()
// 		if IsSet("log") {
// 			// file log with rotation
// 			rfh, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
// 				Filename:   path.Join(conf.Path, conf.Filename),
// 				MaxSize:    conf.MaxSize,
// 				MaxBackups: conf.MaxBackups,
// 				MaxAge:     conf.MaxAge,
// 				Level:      parseLevel(conf),
// 				Formatter:  logFormatter(conf),
// 			})

// 			if err != nil {
// 				panic(err)
// 			}

// 			loggerInstance.AddHook(rfh)

// 			// console log
// 			if conf.Console.Enabled {
// 				loggerInstance.SetLevel(parseLevel(conf))
// 				loggerInstance.SetOutput(colorable.NewColorableStdout())
// 				loggerInstance.SetFormatter(consoleFormatter(conf))
// 			}
// 		} else {
// 			// default logger
// 			Formatter := new(logrus.TextFormatter)
// 			Formatter.TimestampFormat = "02-01-2006 15:04:05"
// 			Formatter.FullTimestamp = true
// 			logrus.SetFormatter(Formatter)
// 		}

// 		loggerInstance.Debug("Config and Logger initialized")
// 	})

// 	return loggerInstance
// }

// // GetLoggerFor return a logrus logger entry tagged with field "module".
// func GetLoggerFor(module string) *logrus.Entry {
// 	return GetLogger().WithField("module", module)
// }

// func parseLevel(conf Log) logrus.Level {
// 	var logLevel logrus.Level

// 	logLevel, err := logrus.ParseLevel(conf.Level)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return logLevel
// }

// func logFormatter(conf Log) logrus.Formatter {
// 	if conf.JSON {
// 		return &logrus.JSONFormatter{
// 			TimestampFormat: conf.TimestampFormat,
// 		}
// 	}

// 	return &prefixed.TextFormatter{
// 		DisableColors:   true,
// 		ForceColors:     false,
// 		TimestampFormat: conf.TimestampFormat,
// 		FullTimestamp:   conf.FullTimestamp,
// 		ForceFormatting: conf.ForceFormatting,
// 	}
// }

// func consoleFormatter(conf Log) logrus.Formatter {
// 	return &prefixed.TextFormatter{
// 		DisableColors:   conf.Console.DisableColors,
// 		ForceColors:     conf.Console.Colors,
// 		TimestampFormat: conf.TimestampFormat,
// 		FullTimestamp:   conf.FullTimestamp,
// 		ForceFormatting: conf.ForceFormatting,
// 	}
// }
