package structures

import "github.com/SevenTV/Common/errors"

type EmoteValidator struct {
	v *Emote
}

func (e *Emote) Validator() EmoteValidator {
	return EmoteValidator{e}
}

func (x EmoteValidator) Name() error {
	if RegExpEmoteName.MatchString(x.v.Name) {
		return nil
	}
	return errors.ErrEmoteNameInvalid()
}
