package errors

import "fmt"

type APIError interface {
	Error() string
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

	ErrBadObjectID APIError = DefineError(70410, "bad object id")
	ErrBadInt      APIError = DefineError(70411, "bad int")

	// Other Client Errors

	ErrEmoteNotEnabled     APIError = DefineError(704610, "emote not enabled")     // client wants to disable an emote which was not enabled to begin with
	ErrEmoteAlreadyEnabled APIError = DefineError(704611, "emote already enabled") // client wants to enable an emote which is already added
	ErrEmoteNameConflict   APIError = DefineError(704611, "emote name conflict")

	// Server Errors

	ErrInternalServerError APIError = DefineError(70500, "internal server error")
)

/*
	API Error Code Format

	7 - error code namespace
	0 - always zero
	X - 4: user error, 5: server error
	X - 0: generic, 1: type error, 4: not found, 6: bad mutation, 7: misc.
	X[...] - any increment.
*/

func DefineError(code int, msg string) error {
	return fmt.Errorf("[%d] %s", code, msg)
}

func AppendErrorDetail(err error, txt string, placeholders ...string) error {
	return fmt.Errorf(fmt.Sprintf(err.Error()+": "+txt, placeholders))
}
