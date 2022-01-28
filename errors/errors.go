package errors

import (
	"fmt"

	"github.com/SevenTV/Common/utils"
)

type APIError interface {
	Error() string
	Message() string
	Code() int
	SetDetail(str string, a ...string) *apiError
	SetFields(d Fields) *apiError
	GetFields() Fields
}

type apiErrorFn func() APIError

var (
	// Generic Client Errors

	ErrUnauthorized          apiErrorFn = DefineError(70401, "unauthorized")           // client is not authenticated
	ErrInsufficientPrivilege apiErrorFn = DefineError(70403, "insufficient privilege") // client lacks privilege
	ErrDontBeSilly           apiErrorFn = DefineError(70470, "don't be silly")         // client is trying to do something stupid

	// Client Not Found

	ErrUnknownEmote    apiErrorFn = DefineError(70440, "unknown emote")   // can't find emote object
	ErrUnknownEmoteSet apiErrorFn = DefineError(70441, "unknown emote")   // can't find emote set object
	ErrUnknownUser     apiErrorFn = DefineError(70442, "unknown user")    // can't find user object
	ErrUnknownRole     apiErrorFn = DefineError(70443, "unknown role")    // can't find role object
	ErrUnknownReport   apiErrorFn = DefineError(70444, "unknown report")  // can't find report object
	ErrUnknownMessage  apiErrorFn = DefineError(70445, "unknown message") // can't find message object
	ErrUnknownBan      apiErrorFn = DefineError(70446, "unknown ban")     // can't find ban object

	// Client Type Errors

	ErrBadObjectID        apiErrorFn = DefineError(70410, "bad object id")
	ErrBadInt             apiErrorFn = DefineError(70411, "bad int")
	ErrValidationRejected apiErrorFn = DefineError(70412, "validation rejected")
	ErrInternalField      apiErrorFn = DefineError(70413, "internal field")
	ErrUnknownRoute       apiErrorFn = DefineError(70441, "unknown route") // the requested api endpoint doesn't exist

	// Other Client Errors

	ErrEmoteNotEnabled      apiErrorFn = DefineError(704610, "emote not enabled")     // client wants to disable an emote which was not enabled to begin with
	ErrEmoteAlreadyEnabled  apiErrorFn = DefineError(704611, "emote already enabled") // client wants to enable an emote which is already added
	ErrEmoteNameConflict    apiErrorFn = DefineError(704612, "emote name conflict")   // client wants to enable an emote but its name conflict with another
	ErrEmoteNameInvalid     apiErrorFn = DefineError(704613, "bad emote name")        // client sent an emote name that did not pass validation
	ErrMissingRequiredField apiErrorFn = DefineError(704680, "missing field")

	// Server Errors

	ErrInternalServerError        apiErrorFn = DefineError(70500, "internal server error")
	ErrInternalIncompleteMutation apiErrorFn = DefineError(70560, "incomplete mutation (internal)")
)

/*
	API Error Code Format

	7 - error code namespace
	0 - always zero
	X - 4: user error, 5: server error
	X - 0: generic, 1: type error, 4: not found, 6: bad mutation, 7: misc.
	X[...] - any increment.
*/

type apiError struct {
	s    string
	code int
	d    Fields
}

type Fields map[string]string

func (e *apiError) Error() string {
	return fmt.Sprintf("[%d] %s", e.code, e.s)
}

func (e *apiError) Message() string {
	return e.s
}

func (e *apiError) Code() int {
	return e.code
}

func (e *apiError) SetDetail(str string, a ...string) *apiError {
	e.s = e.s + ": " + utils.Ternary(len(a) > 0, fmt.Sprintf(str, a), str).(string)
	return e
}

func (e *apiError) SetFields(d Fields) *apiError {
	e.d = d
	return e
}

func (e *apiError) GetFields() Fields {
	return e.d
}

func DefineError(code int, s string) func() APIError {
	return func() APIError {
		return &apiError{
			s, code, Fields{},
		}
	}
}
