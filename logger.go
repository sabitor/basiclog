package simplelog

import (
	"fmt"
	"io"
	// "strconv"
	// "time"
)

type logger struct {
	out io.Writer // log output, e.g. stdout or bufio.Writer
	buf []byte    // buffer for one line of prepared log data
}

func new2(out io.Writer) *logger {
	l := &logger{out: out}
	return l
}

func (l *logger) write(v []any) error {
	l.buf = l.buf[:0]
	l.buf = append(l.buf, fmt.Sprintln(v...)...)
	_, err := l.out.Write(l.buf)
	return err
}
