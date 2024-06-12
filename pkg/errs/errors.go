package errs

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrBadRequest    = errors.New("bad request")
	ErrBrokenUserReq = errors.New("invalid request")
	ErrInvalidCred   = errors.New("invalid user credentials")
	ErrUnknown       = errors.New("internal server error")
	ErrNotFound      = errors.New("entity not found")
	ErrUserExisted   = errors.New("user already existed")
)

type CustomError struct {
	msg     string
	baseerr error
	context map[string]interface{}
}

func New(msg error, args ...any) error {
	err := &CustomError{baseerr: msg}

	errMeta := []string{}
	if len(args) != 0 {
		err.context = make(map[string]interface{})
		if len(args)%2 != 0 {
			err.context["data"] = args[len(args)-1]
			errMeta = append(errMeta, fmt.Sprintf("data=%v", args[len(args)-1]))

			args = args[:len(args)-1]
		}

		for i := 0; i < len(args); i += 2 {
			err.context[args[i].(string)] = args[i+1]
			errMeta = append(errMeta, fmt.Sprintf("%s=%v", args[i], args[i+1]))
		}

		err.msg = fmt.Sprintf("%s; %s", msg.Error(), strings.Join(errMeta, ","))
	} else {
		err.msg = msg.Error()
	}

	return err
}

func (e *CustomError) Error() string {
	return e.msg
}

func (e *CustomError) Is(err error) bool {
	return e.baseerr == err
}
