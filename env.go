package goil

import (
	"fmt"
	"os"
	"strings"
)

const ENV_KEY = "GOIL_MODE"
const (
	PRD = "PRD"
	UAT = "UAT"
	DBG = "DBG"
)

const (
	DBG_MODE = iota
	UAT_MODE
	PRD_MODE
)

//the app running env
//the env is dbg or uat or prd
var run_mode = DBG_MODE

func init() {
	env := os.Getenv(ENV_KEY)
	switch strings.ToUpper(env) {
	case PRD:
		run_mode = PRD_MODE
	case UAT:
		run_mode = UAT_MODE
	default:
		run_mode = DBG_MODE
	}
}

func RunMode() string {
	switch run_mode {
	case PRD_MODE:
		return PRD
	case UAT_MODE:
		return UAT
	case DBG_MODE:
		return DBG
	}
	panic(fmt.Sprintf("unsupport run mode: %d", run_mode))
}

//if the server runing
//the var aim to prevent changing app state in running
var start = false

func starting() {
	start = true
}

//exec only unstart
func RunBeforeStart(f func()) {
	if !start {
		f()
	}
}
