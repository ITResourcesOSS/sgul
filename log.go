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
var logger *zap.SugaredLogger

// GetLogger .
func GetLogger() *zap.SugaredLogger {
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
			logger = zap.New(core, zap.AddCaller()).Sugar()
		} else {
			logger = zap.New(core).Sugar()
		}

	})

	return logger
}

// GetLoggerByConf .
func GetLoggerByConf(conf Log) *zap.SugaredLogger {
	onceLogger.Do(func() {
		//conf := GetConfiguration().Log
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
			logger = zap.New(core, zap.AddCaller()).Sugar()
		} else {
			logger = zap.New(core).Sugar()
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
