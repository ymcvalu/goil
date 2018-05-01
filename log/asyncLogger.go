package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	isatty "github.com/mattn/go-isatty"
)

const (
	nohead = 1 << iota
	wait
)

type signal struct{}

type entry struct {
	msg  string
	time time.Time
	file string
	line int
	flag int8
	cb   chan signal
	tag  string
}

type AsyncLogger struct {
	flag   int
	level  int
	out    io.Writer
	prefix string
	mq     chan *entry
	pool   sync.Pool
}

func NewAsync(out io.Writer, prefix string, flag, level, bs int) (l *AsyncLogger) {
	if out == nil {
		out = os.Stdin
	}
	if level > _max_ {
		level = FatalLevel
	}
	l = &AsyncLogger{
		flag:   flag,
		level:  level,
		prefix: prefix,
		out:    out,
		mq:     make(chan *entry, bs),
		pool: sync.Pool{
			New: func() interface{} {
				return new(entry)
			},
		},
	}

	go poll(l)

	return
}

func poll(l *AsyncLogger) {
	for {
		entry, ok := <-l.mq
		if !ok {
			break
		}
		buf := make([]byte, 0)
		if entry.flag&nohead == 0 {
			l.formatHeader(&buf, entry.tag, entry.time, entry.file, entry.line)
		}
		buf = append(buf, entry.msg...)
		if buf[len(buf)-1] != '\n' {
			buf = append(buf, '\n')
		}

		l.write(&buf)

		if entry.flag&wait == wait {
			entry.cb <- signal{}
			continue
		}

		l.putEntry(entry)
	}
}

func (l *AsyncLogger) formatHeader(buf *[]byte, tag string, t time.Time, file string, line int) {

	if l.prefix != "" {
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

		*buf = append(*buf, tag...)
		*buf = append(*buf, ' ')
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ':', ' ')
	}
}

func (l *AsyncLogger) write(buf *[]byte) error {
	_, err := l.out.Write(*buf)
	return err
}

func (l *AsyncLogger) Printf(format string, msg ...interface{}) {
	entry := l.getEntry()
	entry.flag |= nohead
	entry.msg = fmt.Sprintf(format, msg...)
	l.mq <- entry
}

func (l *AsyncLogger) Print(msg ...interface{}) {
	entry := l.getEntry()
	entry.flag |= nohead
	entry.msg = fmt.Sprint(msg...)
	l.mq <- entry
}

func (l *AsyncLogger) Infof(format string, msg ...interface{}) {
	if l.level > InfoLevel {
		return
	}
	entry := l.getEntry()
	entry.tag = "[info]"
	entry.init(l)
	entry.msg = fmt.Sprintf(format, msg...)
	l.mq <- entry
}

func (l *AsyncLogger) Info(msg ...interface{}) {
	if l.level > InfoLevel {
		return
	}
	entry := l.getEntry()
	entry.tag = "[info]"
	entry.init(l)
	entry.msg = fmt.Sprint(msg...)
	l.mq <- entry
}

func (l *AsyncLogger) Debugf(format string, msg ...interface{}) {
	if l.level > DebugLevel {
		return
	}
	entry := l.getEntry()
	entry.tag = "[debug]"
	entry.init(l)
	entry.msg = fmt.Sprintf(format, msg...)
	l.mq <- entry
}

func (l *AsyncLogger) Debug(msg ...interface{}) {
	if l.level > DebugLevel {
		return
	}
	entry := l.getEntry()
	entry.tag = "[debug]"
	entry.init(l)
	entry.msg = fmt.Sprint(msg...)
	l.mq <- entry
}

func (l *AsyncLogger) Warnf(format string, msg ...interface{}) {
	if l.level > WarnLevel {
		return
	}
	entry := l.getEntry()
	entry.tag = "[warn]"
	entry.init(l)
	entry.msg = fmt.Sprintf(format, msg...)
	l.mq <- entry
}

func (l *AsyncLogger) Warn(msg ...interface{}) {
	if l.level > WarnLevel {
		return
	}
	entry := l.getEntry()
	entry.tag = "[warn]"
	entry.init(l)
	entry.msg = fmt.Sprint(msg...)
	l.mq <- entry
}

func (l *AsyncLogger) Errorf(format string, msg ...interface{}) {
	if l.level > ErrorLevel {
		return
	}
	entry := l.getEntry()
	entry.tag = "[error]"
	entry.init(l)
	entry.msg = fmt.Sprintf(format, msg...)
	l.mq <- entry
}

func (l *AsyncLogger) Error(msg ...interface{}) {
	if l.level > ErrorLevel {
		return
	}
	entry := l.getEntry()
	entry.tag = "[error]"
	entry.init(l)
	entry.msg = fmt.Sprint(msg...)
	l.mq <- entry
}

func (l *AsyncLogger) Panicf(format string, msg ...interface{}) {
	if l.level > PanicLevel {
		return
	}
	entry := l.getEntry()
	entry.tag = "[panic]"
	entry.init(l)
	m := fmt.Sprintf(format, msg...)
	entry.msg = m
	entry.cb = make(chan signal)
	entry.flag |= wait
	l.mq <- entry
	<-entry.cb
	l.putEntry(entry)
	panic(m)
}

func (l *AsyncLogger) Panic(msg ...interface{}) {
	if l.level > PanicLevel {
		return
	}
	entry := l.getEntry()
	entry.tag = "[panic]"
	entry.init(l)
	m := fmt.Sprint(msg...)
	entry.msg = m
	entry.flag |= wait
	entry.cb = make(chan signal)
	l.mq <- entry
	<-entry.cb
	l.putEntry(entry)
	panic(m)
}

func (l *AsyncLogger) Fatalf(format string, msg ...interface{}) {
	if l.level > FatalLevel {
		return
	}
	entry := l.getEntry()
	entry.tag = "[fatal]"
	entry.init(l)
	entry.msg = fmt.Sprintf(format, msg...)
	entry.flag |= wait
	entry.cb = make(chan signal)
	l.mq <- entry
	<-entry.cb
	l.putEntry(entry)
	os.Exit(-1)
}

func (l *AsyncLogger) Fatal(msg ...interface{}) {
	if l.level > FatalLevel {
		return
	}
	entry := l.getEntry()
	entry.tag = "[fatal]"
	entry.init(l)
	entry.msg = fmt.Sprint(msg...)
	entry.flag |= wait
	entry.cb = make(chan signal)
	l.mq <- entry
	<-entry.cb
	l.putEntry(entry)
	os.Exit(-1)
}

func (l *AsyncLogger) callerInfo(calldepth int) (file string, line int) {
	if l.flag&(Lshortfile|Llongfile) != 0 {
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
	}
	return
}

func (l *AsyncLogger) getEntry() *entry {
	return l.pool.Get().(*entry)
}

func (l *AsyncLogger) putEntry(entry *entry) {
	l.pool.Put(entry)
}

func (e *entry) init(l *AsyncLogger) {
	file, line := l.callerInfo(3)
	e.file = file
	e.line = line
	e.time = time.Now()
	e.flag = 0
}

func (l *AsyncLogger) IsTTY() bool {
	file, ok := l.out.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd())
}
