package errors

import (
	"fmt"
)

type APIError interface {
	Error() string
	Message() string
	Code() int
	SetDetail(str string, a ...string) *apiError
	SetFields(d Fields) *apiError
}

var (
	// Generic Client Errors

	ErrUnauthorized          APIError = DefineError(70401, "unauthorized")           // client is not authenticated
	ErrInsufficientPrivilege APIError = DefineError(70403, "insufficient privilege") // client lacks privilege
	ErrDontBeSilly           APIError = DefineError(70470, "don't be silly")         // client is trying to do something stupid

	// Client Not Found

	ErrUnknownEmote    APIError = DefineError(70440, "unknown emote")   // can't find emote object
	ErrUnknownEmoteSet APIError = DefineError(70441, "unknown emote")   // can't find emote set object
	ErrUnknownUser     APIError = DefineError(70442, "unknown user")    // can't find user object
	ErrUnknownRole     APIError = DefineError(70443, "unknown role")    // can't find role object
	ErrUnknownReport   APIError = DefineError(70444, "unknown report")  // can't find report object
	ErrUnknownMessage  APIError = DefineError(70445, "unknown message") // can't find message object
	ErrUnknownBan      APIError = DefineError(70446, "unknown ban")     // can't find ban object

	// Client Type Errors

	ErrBadObjectID        APIError = DefineError(70410, "bad object id")
	ErrBadInt             APIError = DefineError(70411, "bad int")
	ErrValidationRejected APIError = DefineError(70412, "validation rejected")

	// Other Client Errors

	ErrEmoteNotEnabled      APIError = DefineError(704610, "emote not enabled")     // client wants to disable an emote which was not enabled to begin with
	ErrEmoteAlreadyEnabled  APIError = DefineError(704611, "emote already enabled") // client wants to enable an emote which is already added
	ErrEmoteNameConflict    APIError = DefineError(704612, "emote name conflict")   // client wants to enable an emote but its name conflict with another
	ErrMissingRequiredField APIError = DefineError(704613, "missing field")

	// Server Errors

	ErrInternalServerError        APIError = DefineError(70500, "internal server error")
	ErrInternalIncompleteMutation APIError = DefineError(70560, "incomplete mutation (internal)")
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
	e.s = e.s + ": " + fmt.Sprintf(str, a)
	return e
}

func (e *apiError) SetFields(d Fields) *apiError {
	e.d = d
	return e
}

func DefineError(code int, s string) APIError {
	return &apiError{
		s, code, Fields{},
	}
}
