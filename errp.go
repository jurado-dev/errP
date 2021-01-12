package errp

import (
	"encoding/json"
	"regexp"
	"runtime"
)

//	ErrP is a implementation of Error interface that allows to add more information about the error
type ErrP struct {
	Info       string      `json:"info"`
	StatusCode StatusCode  `json:"statusCode"`
	Trace      ErrTrace    `json:"trace"`
	Queue      QueueAction `json:"queue"`
}

type StatusCode struct {
	Code int
}

type ErrTrace struct {
	Line     int
	File     string
	Function string
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

		if code, ok := option.(StatusCode); ok {
			errP.StatusCode = code
		}

		if trace, ok := option.(ErrTrace); ok {
			errP.Trace = trace
		}

		if queue, ok := option.(QueueAction); ok {
			errP.Queue = queue
		}
	}

	return errP
}

//	Code sets a custom status code to the error
func Code(code int) StatusCode {
	return StatusCode{Code: code}
}

//	Trace sets the error trace
func Trace() ErrTrace {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	function := runtime.FuncForPC(pc[0])
	file, line := function.FileLine(pc[0])

	// matching only the file name
	matches := regexp.MustCompile(`(?i)/([\w\d_+*()\[\]%=\-]+\.\w+)$`).FindStringSubmatch(file)
	if len(matches) > 0 {
		file = matches[1]
	}

	funcName := function.Name()
	matches = regexp.MustCompile(`(?i)/([\w\d_+%=*()\[\]\-]+\.[\w\d_+*()\[\]%=\-]+)$`).FindStringSubmatch(funcName)
	if len(matches) > 0 {
		funcName = matches[1]
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
