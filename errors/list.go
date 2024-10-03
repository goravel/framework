package errors

var (
	ErrConfigFacadeNotSet = New("config facade is not initialized")
	ErrJSONParserNotSet   = New("JSON parser is not initialized")

	ErrSessionNotFound              = New("session [%s] not found")
	ErrSessionDriverIsNotSet        = New("session driver is not set")
	ErrSessionDriverNotSupported    = New("session driver [%s] not supported")
	ErrSessionDriverAlreadyExists   = New("session driver [%s] already exists")
	ErrSessionDriverExtensionFailed = New("failed to extend session [%s] driver [%v]")
)
