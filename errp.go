package errp

import (
	"encoding/json"
	"regexp"
	"runtime"
)

//	ErrP is a implementation of Error interface that allows to add more information about the error
type ErrP struct {
	Info  string      `json:"info"`
	Err   string      `json:"err"`
	Code  int         `json:"code"`
	Trace ErrTrace    `json:"trace"`
	Queue QueueAction `json:"queue"`
}

//	Deprecated - StatusCode
type StatusCode struct {
	Code int
}

type ErrTrace struct {
	Line     int    `json:"line"`
	File     string `json:"file"`
	Function string `json:"function"`
}

type QueueAction struct {
	Requeue bool
}

func (e ErrP) Error() string {
	eJson, _ := json.Marshal(e)
	return string(eJson)
}

//	New creates a new err plus error. The input may be a string or a error interface.
//	You can pass other options such as status code, error trace, queue actions.
func New(input interface{}, options ...interface{}) ErrP {

	var errP ErrP

	if str, ok := input.(string); ok {
		errP.Info = str
	}

	if err, ok := input.(error); ok {
		errP.Info = err.Error()
	}

	if len(options) == 0 {
		return errP
	}

	for _, option := range options {

		if code, ok := option.(int); ok {
			errP.Code = code
		}

		if statusCode, ok := option.(StatusCode); ok {
			errP.Code = statusCode.Code
		}

		if trace, ok := option.(ErrTrace); ok {
			errP.Trace = trace
		}

		if queue, ok := option.(QueueAction); ok {
			errP.Queue = queue
		}

		if e, ok := option.(error); ok {
			errP.Err = e.Error()
		}
	}

	return errP
}

//	Deprecated - Code sets a custom status code to the error
func Code(code int) StatusCode {
	return StatusCode{Code: code}
}

//	Trace instances a new ErrTrace using the caller trace
func Trace() ErrTrace {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	function := runtime.FuncForPC(pc[0])
	file, line := function.FileLine(pc[0])

	// matching only the file name
	rgx, err := regexp.Compile(`(?i)/([\w\d_+*()\[\]%=\-]+\.\w+)$`)
	if err == nil {
		matches := rgx.FindStringSubmatch(file)
		if len(matches) > 0 {
			file = matches[1]
		}
	}

	funcName := function.Name()
	rgx, err = regexp.Compile(`(?i)(/[\w\d_\-]+/[\w\d_*().\-]+$)`)
	if err == nil {
		matches := rgx.FindStringSubmatch(funcName)
		if len(matches) > 0 {
			funcName = matches[1]
		}
	}
	return ErrTrace{Line: line, File: file, Function: funcName}
}

//	Queue sets the requeue option. Is useful when you are working with message brokers.
func Queue(requeue bool) QueueAction {
	return QueueAction{Requeue: requeue}
}

//	Decode tries to decode a ErrP json
func Decode(input error) (ErrP, error) {

	var errP ErrP

	err := json.Unmarshal([]byte(input.Error()), &errP)
	if err != nil {
		return errP, New(err, Trace())
	}

	return errP, nil
}
