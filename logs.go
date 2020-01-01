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
