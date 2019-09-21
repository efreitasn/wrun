package main

import (
	"log"
	"os"

	"github.com/efreitasn/customo"
)

var logCmdOut = log.New(os.Stdout, customo.Format(
	"CMD: ",
	customo.AttrBold,
	customo.AttrFgColor4BitsGreen,
), 0)

var logCmdErr = log.New(os.Stdout, customo.Format(
	"CMD: ",
	customo.AttrBold,
	customo.AttrFgColor4BitsRed,
), 0)

var logPreCmdOut = log.New(os.Stdout, customo.Format(
	"PRECMD: ",
	customo.AttrBold,
	customo.AttrFgColor4BitsGreen,
), 0)

var logPreCmdErr = log.New(os.Stdout, customo.Format(
	"PRECMD: ",
	customo.AttrBold,
	customo.AttrFgColor4BitsRed,
), 0)

var logErr = log.New(os.Stderr, customo.Format(
	"ERR: ",
	customo.AttrBold,
	customo.AttrFgColor4BitsRed,
), 0)

var logEvt = log.New(os.Stdout, customo.Format(
	"EVT: ",
	customo.AttrBold,
	customo.AttrFgColor4BitsCyan,
), 0)

// CmdLogger is a log.Logger that implements the os.Writer interface.
type CmdLogger struct {
	l *log.Logger
}

// NewCmdLogger returns a new CmdLogger.
func NewCmdLogger(l *log.Logger) *CmdLogger {
	return &CmdLogger{
		l,
	}
}

func (cLogger *CmdLogger) Write(bs []byte) (n int, err error) {
	cLogger.l.Print(string(bs))

	return len(bs), nil
}
