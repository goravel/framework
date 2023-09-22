package facades

import (
	"github.com/goravel/framework/contracts/translation"
)

func Lang() translation.Translator {
	return App().MakeLang()
}
