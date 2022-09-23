package structures

import (
	"regexp"

	"github.com/seventv/common/errors"
)

type EmoteSetValidator struct {
	v *EmoteSet
}

var RegExpEmoteSetName = regexp.MustCompile(`^[a-zA-Z0-9&'_-~# ]{1,40}$`)

func (e *EmoteSet) Validator() EmoteSetValidator {
	return EmoteSetValidator{e}
}

func (x EmoteSetValidator) Name() error {
	if RegExpEmoteSetName.MatchString(x.v.Name) {
		return nil
	}
	return errors.ErrNameInvalid()
}
