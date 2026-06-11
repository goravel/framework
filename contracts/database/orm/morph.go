package orm

// ModelWithMorphClass lets a model override the value written to and matched against polymorphic
// `*_type` columns. The override takes precedence over both the global morph map (registered via
// orm.MorphMap) and GORM's default of using the parent's table name.
//
// A model that wants to be aliased as e.g. "post" in polymorphic relations declares:
//
//	func (Post) MorphClass() string { return "post" }
//
// This is the recommended primary mechanism for aliasing morph types because it co-locates the
// alias with the model definition. The global morph map is provided as a fallback for models the
// caller cannot modify (e.g. third-party types) or for teams that prefer a single boot-time
// registration.
type ModelWithMorphClass interface {
	MorphClass() string
}
