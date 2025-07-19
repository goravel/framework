package pluralizer

type Inflector interface {
	Plural(word string) string
	Singular(word string) string
}
