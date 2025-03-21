/*
 * Copyright (C) 2020-2022, IrineSistiana
 *
 * This file is part of mosdns.
 *
 * mosdns is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * mosdns is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package logger

import (
	"fmt"
	"os"

	"github.com/haierkeys/obsidian-better-sync-service/pkg/fileurl"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	// Level, See also zapcore.ParseLevel.
	Level string `yaml:"level"`

	// File that logger will be writen into.
	// Default is stderr.
	File string `yaml:"file"`

	// Production enables json output.
	Production bool `yaml:"production"`
}

var (
	stderr = zapcore.Lock(os.Stderr)
	lvl    = zap.NewAtomicLevelAt(zap.InfoLevel)
	l      = zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), stderr, lvl))
	s      = l.Sugar()

	nop = zap.NewNop()
)

func NewLogger(lc Config) (*zap.Logger, error) {

	if !fileurl.IsExist(lc.File) {
		fileurl.CreatePath(lc.File, os.ModePerm)
	}

	lvl, err := zapcore.ParseLevel(lc.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	var fileOut zapcore.WriteSyncer
	if lf := lc.File; len(lf) > 0 {
		f, _, err := zap.Open(lf)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}
		fileOut = zapcore.Lock(f)

		var fileEncoder zapcore.Encoder
		if lc.Production {
			fileEncoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		} else {
			fileEncoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		}

		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(stderr)), lvl)
		fileCore := zapcore.NewCore(fileEncoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileOut)), lvl)

		// 使用 zapcore.NewTee 合并两个 Core
		return zap.New(zapcore.NewTee(consoleCore, fileCore)), nil

	} else {
		return zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), stderr, lvl)), nil
	}
}

// L is a global logger.
func L() *zap.Logger {
	return l
}

// SetLevel sets the log level for the global logger.
func SetLevel(l zapcore.Level) {
	lvl.SetLevel(l)
}

// S is a global logger.
func S() *zap.SugaredLogger {
	return s
}

// Nop is a logger that never writes out logs.
func Nop() *zap.Logger {
	return nop
}
