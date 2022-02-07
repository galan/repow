package say

import (
	"fmt"
	"os"
	"time"

	"strconv"
	"sync/atomic"

	. "github.com/logrusorgru/aurora"
)

var VerboseEnabled bool

func Repow() string {
	return Yellow("✪").Bold().String()
}

func Plain(message string, a ...interface{}) {
	fmt.Printf("%s\n", fmt.Sprintf(message, a...))
}

func Raw(message string) {
	fmt.Print(message)
}

func Verbose(message string, a ...interface{}) {
	if VerboseEnabled {
		fmt.Printf("%s\n", White(fmt.Sprintf(message, a...)))
	}
}

func InfoLn(message string, a ...interface{}) {
	fmt.Printf("%s\n", fmt.Sprintf(message, a...))
}

func Info(message string, a ...interface{}) {
	fmt.Printf("%s", fmt.Sprintf(message, a...))
}

func Header(message string, a ...interface{}) {
	fmt.Printf("%s\n", Cyan(fmt.Sprintf(message, a...)))
}

func Warn(message string, a ...interface{}) {
	fmt.Printf("%s\n", Yellow(fmt.Sprintf(message, a...)))
}

func Error(message string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", Red(fmt.Sprintf(message, a...)))
}

// Progress logs

func colorProject(name string) string {
	return BrightWhite(name).BgGray(4).String()
}

//TODO progress state into struct
func ProgressGeneric(counter *int32, total int, status string, name string, message string, a ...interface{}) {
	totalLen := len(strconv.Itoa(total))
	counterVal := atomic.AddInt32(counter, 1)
	Plain("(%*d/%d) [%s] %s %s", totalLen, counterVal, total, status, colorProject(name), fmt.Sprintf(message, a...))
}

func ProgressSuccess(counter *int32, total int, name string, message string, a ...interface{}) {
	ProgressGeneric(counter, total, Green("✔").Bold().String(), name, message, a...)
}

func ProgressWarn(counter *int32, total int, err error, name string, message string, a ...interface{}) {
	msg := message
	if err != nil {
		msg = msg + ": " + err.Error()
	}
	ProgressGeneric(counter, total, Yellow("!").Bold().String(), name, msg, a...)
}

func ProgressError(counter *int32, total int, err error, name string, message string, a ...interface{}) {
	msg := message
	if err != nil {
		msg = msg + ": " + err.Error()
	}
	ProgressGeneric(counter, total, Red("✘").Bold().String(), name, msg, a...)
}

func ProgressErrorArray(counter *int32, total int, errs []error, name string, message string, a ...interface{}) {
	msg := message
	totalLen := len(strconv.Itoa(total)) * 2
	for _, e := range errs {
		msg = msg + fmt.Sprintf("\n%*s        - %s", totalLen, "", e.Error())
	}
	ProgressGeneric(counter, total, Red("✘").Bold().String(), name, msg, a...)
}

func Timer(start time.Time) {
	func(start time.Time) {
		InfoLn("%s Finished, took %s", Repow(), time.Since(start))
	}(start)
}
