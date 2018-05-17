package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
)

type Logger struct {
	mu        sync.RWMutex
	prefix    string
	flag      int
	out       io.Writer
	level     int
	calldepth int
}

func New(out io.Writer, prefix string, flag, level, calldepth int) *Logger {
	if out == nil {
		out = os.Stdin
	}
	if level > _max_ {
		level = FatalLevel
	}

	return &Logger{
		out:       out,
		prefix:    prefix,
		flag:      flag,
		level:     level,
		calldepth: calldepth,
	}
}

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	l.out = w
	l.mu.Unlock()
}

func (l *Logger) SetLevel(lvl int) {
	if lvl >= _max_ {
		lvl = FatalLevel
	}
	l.mu.Lock()
	l.level = lvl
	l.mu.Unlock()
}

func (l *Logger) Level() (lvl int) {
	l.mu.RLock()
	lvl = l.level
	l.mu.RUnlock()
	return
}

func (l *Logger) Flags() (flag int) {
	l.mu.RLock()
	flag = l.flag
	l.mu.RUnlock()
	return
}

func (l *Logger) SetFlags(flag int) {
	l.mu.Lock()
	l.flag = flag
	l.mu.Unlock()
}

func (l *Logger) Prefix() (prefix string) {
	l.mu.RLock()
	prefix = l.prefix
	l.mu.RUnlock()
	return
}

func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	l.prefix = prefix
	l.mu.Unlock()
}

func (l *Logger) formatHeader(buf *[]byte, prefix string, t time.Time, file string, line int) {
	if prefix != "" {
		*buf = append(*buf, prefix...)
	} else {
		*buf = append(*buf, l.prefix...)
	}

	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&LUTC != 0 {
			t = t.UTC()
		}
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	if l.flag&(Lshortfile|Llongfile) != 0 {
		if l.flag&Lshortfile != 0 {
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					file = file[i+1:]
					break
				}
			}
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ':', ' ')
	}
}

func (l *Logger) write(calldepth int, prefix, s string) error {
	now := time.Now()
	var file string
	var line int
	l.mu.RLock()
	if l.flag&(Lshortfile|Llongfile) != 0 {
		l.mu.RUnlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.RLock()
	}

	buf := make([]byte, 0)
	l.formatHeader(&buf, prefix, now, file, line)
	l.mu.RUnlock()
	buf = append(buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		buf = append(buf, '\n')
	}
	_, err := l.out.Write(buf)
	return err
}

func (l *Logger) Printf(format string, msg ...interface{}) {
	s := fmt.Sprintf(format, msg...)
	buf := make([]byte, 0)
	buf = append(buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		buf = append(buf, '\n')
	}
	l.out.Write(buf)
}

func (l *Logger) Print(msg ...interface{}) {
	s := fmt.Sprint(msg...)
	buf := make([]byte, 0)
	buf = append(buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		buf = append(buf, '\n')
	}
	l.out.Write(buf)
}

func (l *Logger) Infof(format string, msg ...interface{}) {
	if l.level > InfoLevel {
		return
	}
	l.write(l.calldepth, "[info] ", fmt.Sprintf(format, msg...))
}

func (l *Logger) Info(msg ...interface{}) {
	if l.level > InfoLevel {
		return
	}
	l.write(l.calldepth, "[info] ", fmt.Sprint(msg...))
}

func (l *Logger) Debugf(format string, msg ...interface{}) {
	if l.level > DebugLevel {
		return
	}
	l.write(l.calldepth, "[debug] ", fmt.Sprintf(format, msg...))
}

func (l *Logger) Debug(msg ...interface{}) {
	if l.level > DebugLevel {
		return
	}
	l.write(l.calldepth, "[debug] ", fmt.Sprint(msg...))
}

func (l *Logger) Warnf(format string, msg ...interface{}) {
	if l.level > WarnLevel {
		return
	}
	l.write(l.calldepth, "[warn] ", fmt.Sprintf(format, msg...))
}

func (l *Logger) Warn(msg ...interface{}) {
	if l.level > WarnLevel {
		return
	}
	l.write(l.calldepth, "[warn] ", fmt.Sprint(msg...))
}

func (l *Logger) Errorf(format string, msg ...interface{}) {
	if l.level > ErrorLevel {
		return
	}
	l.write(l.calldepth, "[error] ", fmt.Sprintf(format, msg...))
}

func (l *Logger) Error(msg ...interface{}) {
	if l.level > ErrorLevel {
		return
	}
	l.write(l.calldepth, "[error] ", fmt.Sprint(msg...))
}

func (l *Logger) Panicf(format string, msg ...interface{}) {
	if l.level > PanicLevel {
		return
	}
	message := fmt.Sprintf(format, msg...)
	l.write(l.calldepth, "[panic] ", message)
	panic(message)
}

func (l *Logger) Panic(msg ...interface{}) {
	if l.level > PanicLevel {
		return
	}
	message := fmt.Sprint(msg...)
	l.write(l.calldepth, "[panic] ", message)
	panic(message)
}

func (l *Logger) Fatalf(format string, msg ...interface{}) {
	if l.level > FatalLevel {
		return
	}

	l.write(l.calldepth, "[fatal] ", fmt.Sprintf(format, msg...))
	os.Exit(-1)
}

func (l *Logger) Fatal(msg ...interface{}) {
	if l.level > FatalLevel {
		return
	}

	l.write(l.calldepth, "[fatal] ", fmt.Sprint(msg...))
	os.Exit(-1)
}

func (l *Logger) IsTTY() bool {
	file, ok := l.out.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd())
}
