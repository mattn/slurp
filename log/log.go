package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/dustin/go-humanize"
)

var Rate = time.Millisecond * 300

var Flags = log.Ltime

type Log interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})

	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})

	ReadProgress(io.Reader, string, int64) io.ReadCloser
	Counter(string, int) *Counter

	New(string) Log
}

func New() Log {
	l := log.New(os.Stdout, "", Flags)
	return &logger{l, ""}
}

type printFormater interface {
	Printf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})

	Flags() int //Just to distingush from Log.
}

type logger struct {
	printFormater
	prefix string
}

func (l *logger) New(prefix string) Log {
	return &logger{l.printFormater, l.prefix + prefix}
}

func (l *logger) Printf(format string, v ...interface{}) {
	l.printFormater.Printf("%s%s", l.prefix, fmt.Sprintf(format, v...))
}

func (l *logger) Print(v ...interface{}) {
	l.printFormater.Printf("%s%s", l.prefix, fmt.Sprint(v...))
}

func (l *logger) Println(v ...interface{}) {
	l.printFormater.Printf("%s%s", l.prefix, fmt.Sprint(v...))
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	l.printFormater.Fatalf("%s%s", l.prefix, fmt.Sprintf(format, v...))
}

func (l *logger) Fatal(v ...interface{}) {
	l.printFormater.Fatalf("%s%s", l.prefix, fmt.Sprint(v...))
}

func (l *logger) Fatalln(v ...interface{}) {
	l.printFormater.Fatalf("%s%s", l.prefix, fmt.Sprint(v...))
}

func (l *logger) ReadProgress(r io.Reader, name string, size int64) io.ReadCloser {
	return &ProgressBar{r, name, size, 0, l, humanize.Bytes(uint64(size)), 0, NewRateLimit(Rate)}
}

func (l *logger) Counter(name string, size int) *Counter {
	return &Counter{name, size, 0, "", l, NewRateLimit(Rate / 2)}
}

type ProgressBar struct {
	io.Reader

	name string
	size int64

	done int64
	l    Log

	sizeHuman string //So we don't calcuate it in every read.
	last      int64

	limit *ratelimit
}

func (p *ProgressBar) print() {
	p.l.Printf("%s [%3d%%] %s of %s\n",
		p.name,
		p.done*100/p.size,
		humanize.Bytes(uint64(p.done)),
		p.sizeHuman)
}
func (p *ProgressBar) Read(b []byte) (int, error) {
	n, err := p.Reader.Read(b)
	p.done += int64(n)

	if (p.done-p.last) > (p.size/50) && !p.limit.Limit() {
		p.last = p.done
		p.print()
	}

	return n, err
}

func (p *ProgressBar) Close() error {
	p.print()
	var err error
	c, ok := p.Reader.(io.Closer)
	if ok {
		err = c.Close()
	}
	return err
}

type Counter struct {
	name string
	size int

	cur  int
	last string
	l    Log

	limit *ratelimit
}

func (c *Counter) Set(s int, last string) {
	c.cur = s
	c.last = last
	if !c.limit.Limit() || c.cur == c.size {
		c.print()
	}
}

func (c *Counter) print() {
	c.l.Printf("%s [%3d%%] %d of %d %s\n", c.name, c.cur*100/c.size, c.cur, c.size, c.last)
}
