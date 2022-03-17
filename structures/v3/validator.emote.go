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

type EmoteVersionValidator struct {
	v *EmoteVersion
}

func (e *EmoteVersion) Validator() EmoteVersionValidator {
	return EmoteVersionValidator{e}
}

func (x EmoteVersionValidator) Name() error {
	if RegExpEmoteVersionName.MatchString(x.v.Name) {
		return nil
	}
	return errors.ErrEmoteVersionNameInvalid()
}

func (x EmoteVersionValidator) Description() error {
	if RegExpEmoteVersionDescription.MatchString(x.v.Description) {
		return nil
	}
	return errors.ErrEmoteVersionDescriptionInvalid()
}
