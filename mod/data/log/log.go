package log

import (
	"github.com/sail-services/sail-go/com/base"
	"github.com/sail-services/sail-go/com/sys/fs"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type (
	Log struct {
		prefix            string
		mutex             *sync.Mutex
		level             log_level
		data_type         data_type
		formatter         Formatter
		writer            io.Writer
		write_closer      io.WriteCloser
		file_path         string
		file_name         string
		file_writer       io.Writer
		file_write_closer io.WriteCloser
	}
	Formatter interface {
		Format(time.Time, log_level, string, *Log) string
	}
	formatter struct{}
	log_level int
	data_type int
)

const (
	LEVEL_INFO    log_level = iota // 信息
	LEVEL_DATA                     // 数据
	LEVEL_WARNING                  // 警告
	LEVEL_ERROR                    // 错误
	LEVEL_FATAL                    // 严重错误

	DATA_NONE      data_type = iota // 无信息
	DATA_BASIC                      // 简单信息
	DATA_FILE_CODE                  // 代码位置
	DATA_TIME                       // 时间
	DATA_ALL                        // 时间及代码位置
)

func New(writer io.Writer, level log_level, dtype data_type) *Log {
	log := Log{
		level:     level,
		data_type: dtype,
		writer:    writer,
		formatter: &formatter{},
		mutex:     &sync.Mutex{},
	}
	if wc, ok := writer.(io.WriteCloser); ok {
		log.write_closer = wc
	}
	return &log
}

func NewFile(path string, level log_level, dtype data_type) (log *Log) {
	return &Log{
		level:     level,
		data_type: dtype,
		file_path: path,
		formatter: &formatter{},
		mutex:     &sync.Mutex{},
	}
}

func NewWriterAndFile(writer io.Writer, path string, level log_level, dtype data_type) *Log {
	log := Log{
		level:     level,
		data_type: dtype,
		writer:    writer,
		file_path: path,
		formatter: &formatter{},
		mutex:     &sync.Mutex{},
	}
	if wc, ok := writer.(io.WriteCloser); ok {
		log.write_closer = wc
	}
	return &log
}

func LevelGetByS(str string) log_level {
	str = strings.ToUpper(str)
	switch str {
	case "INFO":
		return LEVEL_INFO
	case "DATA":
		return LEVEL_DATA
	case "WARN":
		return LEVEL_WARNING
	case "ERROR":
		return LEVEL_ERROR
	case "FATAL":
		return LEVEL_FATAL
	default:
		return -1
	}
}

// ======================
// Log
// ======================

func (log *Log) LevelSet(level log_level) {
	log.level = level
}

func (log *Log) LevelGet() log_level {
	return log.level
}

func (log *Log) Close() {
	log.mutex.Lock()
	if log.write_closer != nil {
		log.write_closer.Close()
	}
	if log.file_write_closer != nil {
		log.file_write_closer.Close()
	}
	log.mutex.Unlock()
}

func (log *Log) Writer(str string) {
	if log.writer != nil {
		log.writer.Write([]byte(str))
	}
	if log.file_writer != nil {
		log.file_writer.Write([]byte(str))
	}
}

func (log *Log) WriterGet() *io.Writer {
	return &log.writer
}

func (log *Log) Format(t time.Time, level log_level, message string) string {
	log.mutex.Lock()
	if len(log.file_path) != 0 && log.file_name != time.Now().Format("02PM") {
		fs.PathNew(log.file_path+"/"+time.Now().Format("200601"), 0750)
		log.file_name = time.Now().Format("02pm")
		file, err := os.OpenFile(fmt.Sprintf("%s/%s/%s.log", log.file_path, time.Now().Format("200601"), log.file_name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
		if err != nil {
			return ""
		}
		log.file_writer = file
		log.file_write_closer = file
	}
	var msg string
	if log.formatter != nil {
		msg = log.formatter.Format(t, level, message, log)
	}
	log.mutex.Unlock()
	return msg
}

func (log *Log) Print(v ...interface{}) {
	log.Writer(fmt.Sprint(v...))
}

func (log *Log) Println(v ...interface{}) {
	log.Writer(fmt.Sprintln(v...))
}

func (log *Log) Printf(format string, v ...interface{}) {
	log.Writer(fmt.Sprintf(format, v...))
}

func (log *Log) log(level log_level, v ...interface{}) {
	if level >= log.level {
		log.Writer(log.Format(time.Now(), level, fmt.Sprint(v...)))
	}
}

func (log *Log) logf(level log_level, format string, v ...interface{}) {
	if level >= log.level {
		log.Writer(log.Format(time.Now(), level, fmt.Sprintf(format, v...)))
	}
}

func (log *Log) logln(level log_level, v ...interface{}) {
	if level >= log.level {
		log.Writer(log.Format(time.Now(), level, fmt.Sprintln(v...)))
	}
}

func (log *Log) Info(v ...interface{}) {
	log.log(LEVEL_INFO, v...)
}

func (log *Log) Infof(format string, v ...interface{}) {
	log.logf(LEVEL_INFO, format, v...)
}

func (log *Log) Infoln(v ...interface{}) {
	log.logln(LEVEL_INFO, v...)
}

func (log *Log) Data(v ...interface{}) {
	log.log(LEVEL_DATA, v...)
}

func (log *Log) Dataf(format string, v ...interface{}) {
	log.logf(LEVEL_DATA, format, v...)
}

func (log *Log) Dataln(v ...interface{}) {
	log.logln(LEVEL_DATA, v...)
}

func (log *Log) Warn(v ...interface{}) {
	log.log(LEVEL_WARNING, v...)
}

func (log *Log) Warnf(format string, v ...interface{}) {
	log.logf(LEVEL_WARNING, format, v...)
}

func (log *Log) Warnln(v ...interface{}) {
	log.logln(LEVEL_WARNING, v...)
}

func (log *Log) Error(v ...interface{}) {
	log.log(LEVEL_ERROR, v...)
}

func (log *Log) Errorf(format string, v ...interface{}) {
	log.logf(LEVEL_ERROR, format, v...)
}

func (log *Log) Errorln(v ...interface{}) {
	log.logln(LEVEL_ERROR, v...)
}

func (log *Log) Fatal(v ...interface{}) {
	log.log(LEVEL_FATAL, v...)
	os.Exit(1)
}

func (log *Log) Fatalf(format string, v ...interface{}) {
	log.logf(LEVEL_FATAL, format, v...)
	log.Close()
	os.Exit(1)
}

func (log *Log) Fatalln(v ...interface{}) {
	log.logln(LEVEL_FATAL, v...)
	log.Close()
	os.Exit(1)
}

func (log *Log) Panic(v ...interface{}) {
	log.log(LEVEL_FATAL, v...)
	log.Close()
	panic(nil)
}

func (log *Log) Panicf(format string, v ...interface{}) {
	log.logf(LEVEL_FATAL, format, v...)
	log.Close()
	panic(nil)
}

func (log *Log) Panicln(v ...interface{}) {
	log.logln(LEVEL_FATAL, v...)
	log.Close()
	panic(nil)
}

// ======================
// formatter
// ======================

func (f *formatter) Format(t time.Time, level log_level, message string, log *Log) string {
	time_str := t.Format("2006-01-02 15:04:05")
	var format string
	switch log.data_type {
	case DATA_NONE:
		format = fmt.Sprintf("%s", message)
	case DATA_BASIC:
		format = fmt.Sprintf("[%s] %s", level_to_string(level), message)
	case DATA_FILE_CODE:
		format = fmt.Sprintf("(%s) [%s] %s", base.CallerInfoGet(), level_to_string(level), message)
	case DATA_TIME:
		format = fmt.Sprintf("(%s) [%s] %s", time_str, level_to_string(level), message)
	case DATA_ALL:
		format = fmt.Sprintf("(%s) [%s] <%s> %s", time_str, level_to_string(level), base.CallerInfoGet(), message)
	}
	return format
}

// ======================
// func
// ======================

func level_to_string(level log_level) string {
	switch level {
	case LEVEL_INFO:
		return "INFO"
	case LEVEL_DATA:
		return "DATA"
	case LEVEL_WARNING:
		return "WARN"
	case LEVEL_ERROR:
		return "ERRO"
	case LEVEL_FATAL:
		return "FATL"
	default:
		return "UNKN"
	}
}
