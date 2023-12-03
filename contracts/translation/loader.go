package translation

type Loader interface {
	// Load the messages for the given locale.
	Load(folder string, locale string) (map[string]map[string]string, error)
}
