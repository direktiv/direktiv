package dlog

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	appTCP = "http://127.0.0.1:8889"
	fnTCP  = "http://127.0.0.1:8888"
)

var (
	appLogger, fnLogger *zap.Logger
)

func init() {
	if err := zap.RegisterSink("http", NewHTTPSink); err != nil {
		panic(err)
	}

	var (
		err error
	)
	// startup probes don't work here. reporting success too early
	for i := 0; i < 60; i++ {
		log.Printf("connecting to logging %v\n", appTCP)
		_, err = http.Post(appTCP, "application/json",
			bytes.NewBuffer([]byte("")))

		time.Sleep(1 * time.Second)

		if err == nil {
			break
		}
	}

	if err != nil {
		log.Fatalf("can not start logging: %v", err)
	}

}

// ApplicationLogger returns logger for applications
func ApplicationLogger(component string) (*zap.SugaredLogger, error) {

	var err error
	if appLogger == nil {
		appLogger, err = customLogger(appTCP)
		if err != nil {
			return nil, err
		}
	}

	return appLogger.With(zap.String("component", component)).Sugar(), nil
}

// FunctionsLogger returns logger for functions
func FunctionsLogger() (*zap.SugaredLogger, error) {

	var err error
	if fnLogger == nil {
		fnLogger, err = customLogger(fnTCP)
		if err != nil {
			return nil, err
		}
	}
	return fnLogger.With(zap.String("component", "functions")).Sugar(), nil
}

func customLogger(tcp string) (*zap.Logger, error) {

	l := os.Getenv("DIREKTIV_DEBUG")

	inLvl := zapcore.InfoLevel
	if l == "true" {
		inLvl = zapcore.DebugLevel
	}

	consoleOut := zapcore.Lock(os.Stdout)
	logLvl := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= inLvl
	})

	// console
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)

	// tcp
	tcpEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	writer, _, err := zap.Open(tcp)
	if err != nil {
		return nil, err
	}

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleOut, logLvl),
		zapcore.NewCore(tcpEncoder, writer, logLvl),
	)

	return zap.New(core, zap.AddCaller()), nil

}
