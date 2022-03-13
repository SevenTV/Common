package errors

import (
	"fmt"
	"strings"

	"github.com/SevenTV/Common/utils"
)

type APIError interface {
	Error() string
	Message() string
	Code() int
	SetDetail(str string, a ...string) *apiError
	SetFields(d Fields) *apiError
	GetFields() Fields
	ExpectedHTTPStatus() int
	WithHTTPStatus(s int) *apiError
}

type apiErrorFn func() APIError

var (
	// Generic Client Errors

	ErrUnauthorized          apiErrorFn = DefineError(70401, "Sign-In Required", 401)       // client is not authenticated
	ErrInsufficientPrivilege apiErrorFn = DefineError(70403, "Insufficient Privilege", 403) // client lacks privilege
	ErrDontBeSilly           apiErrorFn = DefineError(70470, "Don't Be Silly", 403)         // client is trying to do something stupid

	// Client Not Found

	ErrUnknownEmote          apiErrorFn = DefineError(70440, "Unknown Emote", 404)           // can't find emote object
	ErrUnknownEmoteSet       apiErrorFn = DefineError(70441, "Unknown Emote Set", 404)       // can't find emote set object
	ErrUnknownUser           apiErrorFn = DefineError(70442, "Unknown User", 404)            // can't find user object
	ErrUnknownUserConnection apiErrorFn = DefineError(70442, "Unknown User Connection", 404) // can't find user connection object
	ErrUnknownRole           apiErrorFn = DefineError(70443, "Unknown Role", 404)            // can't find role object
	ErrUnknownReport         apiErrorFn = DefineError(70444, "Unknown Report", 404)          // can't find report object
	ErrUnknownMessage        apiErrorFn = DefineError(70445, "Unknown Message", 404)         // can't find message object
	ErrUnknownBan            apiErrorFn = DefineError(70446, "Unknown Ban", 404)             // can't find ban object
	ErrNoItems               apiErrorFn = DefineError(70449, "No Items Found", 404)          // search returned nothing

	// Client Type Errors

	ErrInvalidRequest     apiErrorFn = DefineError(70410, "Invalid Request", 400)
	ErrBadObjectID        apiErrorFn = DefineError(70411, "Bad Object ID", 400)
	ErrBadInt             apiErrorFn = DefineError(70412, "Bad Int", 400)
	ErrValidationRejected apiErrorFn = DefineError(70413, "Validation Rejected", 400)
	ErrInternalField      apiErrorFn = DefineError(70414, "Internal Field", 400)
	ErrUnknownRoute       apiErrorFn = DefineError(70441, "Unknown Route", 400) // the requested api endpoint doesn't exist

	// Other Client Errors

	ErrEmoteNotEnabled      apiErrorFn = DefineError(704610, "Emote Not Enabled", 400)     // client wants to disable an emote which was not enabled to begin with
	ErrEmoteAlreadyEnabled  apiErrorFn = DefineError(704611, "Emote Already Enabled", 400) // client wants to enable an emote which is already added
	ErrEmoteNameConflict    apiErrorFn = DefineError(704612, "Emote Name Conflict", 400)   // client wants to enable an emote but its name conflict with another
	ErrEmoteNameInvalid     apiErrorFn = DefineError(704613, "Bad Emote Name", 400)        // client sent an emote name that did not pass validation
	ErrNoSpaceAvailable     apiErrorFn = DefineError(704620, "No Space Available", 403)    // the target object is full
	ErrMissingRequiredField apiErrorFn = DefineError(704680, "Missing Field", 400)

	// Server Errors

	ErrInternalServerError        apiErrorFn = DefineError(70500, "Internal Server Error", 500)
	ErrMissingInternalDependency  apiErrorFn = DefineError(70510, "Missing Internal Dependency", 503)
	ErrInternalIncompleteMutation apiErrorFn = DefineError(70560, "Incomplete Mutation (internal)", 500)
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
	s                  string
	code               int
	d                  Fields
	expectedHttpStatus int
}

type Fields map[string]interface{}

func (e *apiError) Error() string {
	return fmt.Sprintf("[%d] %s", e.code, strings.ToLower(e.s))
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

func (e *apiError) ExpectedHTTPStatus() int {
	return e.expectedHttpStatus
}

func (e *apiError) WithHTTPStatus(s int) *apiError {
	e.expectedHttpStatus = s
	return e
}

func DefineError(code int, s string, httpStatus int) func() APIError {
	return func() APIError {
		return &apiError{
			s, code, Fields{}, httpStatus,
		}
	}
}
