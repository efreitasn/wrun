package logs

import (
	"log"
	"os"

	"github.com/efreitasn/customo"
)

// CmdOut is the logger to print the data from the command's Stdout.
var CmdOut = log.New(os.Stdout, customo.Format(
	"CMD: ",
	customo.AttrBold,
	customo.AttrFgColor4BitsGreen,
), 0)

// CmdErr is the logger used to print errors from the command's Stderr.
var CmdErr = log.New(os.Stdout, customo.Format(
	"CMD: ",
	customo.AttrBold,
	customo.AttrFgColor4BitsRed,
), 0)

// Err is the logger used to print errors.
var Err = log.New(os.Stderr, customo.Format(
	"ERR: ",
	customo.AttrBold,
	customo.AttrFgColor4BitsRed,
), 0)

// Evt is the logger used to print events.
var Evt = log.New(os.Stdout, customo.Format(
	"EVT: ",
	customo.AttrBold,
	customo.AttrFgColor4BitsCyan,
), 0)
