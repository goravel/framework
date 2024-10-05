package errors

var (
	ErrConfigFacadeNotSet = New("config facade is not initialized")
	ErrJSONParserNotSet   = New("JSON parser is not initialized")
	ErrCacheFacadeNotSet  = New("cache facade is not initialized")
	ErrOrmFacadeNotSet    = New("orm facade is not initialized")
	ErrLogFacadeNotSet    = New("log facade is not initialized")
	ErrQueueFacadeNotSet  = New("queue facade is not initialized")
	ErrApplicationNotSet  = New("application instance is not initialized")

	ErrCacheSupportRequired = New("cache support is required")
	ErrCacheForeverFailed   = New("cache forever is failed")

	ErrSessionNotFound              = New("session [%s] not found")
	ErrSessionDriverIsNotSet        = New("session driver is not set")
	ErrSessionDriverNotSupported    = New("session driver [%s] not supported")
	ErrSessionDriverAlreadyExists   = New("session driver [%s] already exists")
	ErrSessionDriverExtensionFailed = New("session failed to extend session [%s] driver [%v]")

	ErrAuthRefreshTimeExceeded = New("authentication refresh time limit exceeded")
	ErrAuthTokenExpired        = New("authentication token has expired")
	ErrAuthNoPrimaryKeyField   = New("no primary key field found in the model, ensure primary key is set, e.g., orm.Model")
	ErrAuthEmptySecret         = New("authentication secret is missing or required")
	ErrAuthTokenDisabled       = New("authentication token has been disabled")
	ErrAuthParseTokenFirst     = New("authentication token must be parsed first")
	ErrAuthInvalidClaims       = New("authentication token contains invalid claims")
	ErrAuthInvalidToken        = New("authentication token is invalid")
	ErrAuthInvalidKey          = New("authentication key is invalid")

	ErrCacheDriverNotSupported        = New("invalid driver: %s, only support memory, custom")
	ErrCacheStoreContractNotFulfilled = New("%s doesn't implement contracts/cache/store")
	ErrCacheMemoryInvalidIntValueType = New("value type of %s is not *atomic.Int64 or *int64 or *atomic.Int32 or *int32")

	ErrCryptAppKeyNotSet        = New("APP_KEY is required in artisan environment")
	ErrCryptInvalidAppKeyLength = New("invalid APP_KEY length: %d bytes")
	ErrCryptMissingIVKey        = New("decrypt payload error: missing iv key")
	ErrCryptMissingValueKey     = New("decrypt payload error: missing value key")

	ErrEventListenerNotBind = New("event %v doesn't bind listeners")
)
