package facades

import "github.com/goravel/framework/contracts/translation"

func Translation() translation.Translation {
	return App().MakeTranslation()
}

func Language(language string) translation.Language {
	return Translation().Language(language)
}
