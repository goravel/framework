package errors

var (
	ErrConfigFacadeNotSet = New("config facade is not initialized")
	ErrJSONParserNotSet   = New("JSON parser is not initialized")
	ErrCacheFacadeNotSet  = New("cache facade is not initialized")
	ErrOrmFacadeNotSet    = New("orm facade is not initialized")

	ErrCacheSupportRequired = New("cache support is required")
	ErrCacheForeverFailed   = New("cache forever is failed")

	ErrSessionNotFound              = New("session [%s] not found", ModuleSession)
	ErrSessionDriverIsNotSet        = New("driver is not set", ModuleSession)
	ErrSessionDriverNotSupported    = New("driver [%s] not supported", ModuleSession)
	ErrSessionDriverAlreadyExists   = New("driver [%s] already exists")
	ErrSessionDriverExtensionFailed = New("failed to extend session [%s] driver [%v]", ModuleSession)

	ErrAuthRefreshTimeExceeded = New("authentication refresh time limit exceeded")
	ErrAuthTokenExpired        = New("authentication token has expired")
	ErrAuthNoPrimaryKeyField   = New("no primary key field found in the model, ensure primary key is set, e.g., orm.Model")
	ErrAuthEmptySecret         = New("authentication secret is missing or required")
	ErrAuthTokenDisabled       = New("authentication token has been disabled")
	ErrAuthParseTokenFirst     = New("authentication token must be parsed first")
	ErrAuthInvalidClaims       = New("authentication token contains invalid claims")
	ErrAuthInvalidToken        = New("authentication token is invalid")
	ErrAuthInvalidKey          = New("authentication key is invalid")
)
