package p2p

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

func init() {
	logConfig := zap.NewDevelopmentConfig()
	logConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	log, err := logConfig.Build()
	if err != nil {
		panic(err)
	}
	defer log.Sync()
	logger = log.Named("p2p").WithOptions().Sugar()
}