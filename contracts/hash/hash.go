package hash

//go:generate mockery --name=Hash
type Hash interface {
	// Make returns the hashed value of the given string.
	Make(string) string
	// Check checks if the given string matches the given hash.
	Check(string, string) bool
	// NeedsRehash checks if the given hash needs to be rehashed.
	NeedsRehash(string) bool
}
