package translation

type Loader interface {
	// Load the messages for the given locale.
	Load(locale string, folder string) (map[string]map[string]any, error)
}
