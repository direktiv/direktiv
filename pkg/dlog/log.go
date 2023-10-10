package dlog

import (
	"os"
	"strconv"

	"github.com/direktiv/direktiv/pkg/util"
	"github.com/direktiv/direktiv/pkg/version"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const direktivDebug = "DIREKTIV_DEBUG"

// ApplicationLogger returns logger for applications.
func ApplicationLogger(component string) (*zap.SugaredLogger, error) {
	appLogger, err := customLogger()
	if err != nil {
		return nil, err
	}
	return appLogger.With(zap.String("component", component), zap.String("build", version.Version)).Sugar(), nil
}

// FunctionsLogger returns logger for functions.
func FunctionsLogger() (*zap.SugaredLogger, error) {
	fnLogger, err := customLogger()
	if err != nil {
		return nil, err
	}
	return fnLogger.With(zap.String("component", "function")).Sugar(), nil
}

func customLogger() (*zap.Logger, error) {
	l, _ := strconv.ParseBool(os.Getenv(direktivDebug))

	inLvl := zapcore.InfoLevel
	if l {
		inLvl = zapcore.DebugLevel
	}

	errOut := zapcore.Lock(os.Stderr)

	logLvl := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= inLvl
	})

	// console
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)

	jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, errOut, logLvl),
	)

	if os.Getenv(util.DirektivLogJSON) == "json" {
		core = zapcore.NewTee(
			zapcore.NewCore(jsonEncoder, errOut, logLvl),
		)
	}

	return zap.New(core, zap.AddCaller()), nil
}
