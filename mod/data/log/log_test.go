package log_test

import (
	"os"
	"github.com/sail-services/sail-go/mod/data/log"
	"testing"
)

func Test(t *testing.T) {
	logger := log.New(os.Stdout, log.LEVEL_INFO, log.DATA_TIME)
	logger.Infoln("info message")
	logger.Warnln("warning message")
	logger.Errorln("error message")
}

func Test_File(t *testing.T) {
	logger := log.NewFile("test_log", log.LEVEL_INFO, log.DATA_TIME)
	logger.Infoln("info message")
	logger.Warnln("warning message")
	logger.Errorln("error message")
}
