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

	ErrFilesystemDiskNotSet          = New("please set default disk")
	ErrFilesystemDriverNotSupported  = New("invalid driver: %s, only support local, custom")
	ErrFilesystemInvalidCustomDriver = New("init %s disk fail: via must be implement filesystem.Driver or func() (filesystem.Driver, error)")
	ErrFilesystemDeleteDirectory     = New("can't delete directory, please use DeleteDirectory")

	ErrGrpcEmptyClientHost         = New("client's host can't be empty")
	ErrGrpcEmptyClientPost         = New("client's port can't be empty")
	ErrGrpcInvalidInterceptorsType = New("the type of clients.%s.interceptors must be []string")
	ErrGrpcEmptyServerHost         = New("host can't be empty")
	ErrGrpcEmptyServerPort         = New("port can't be empty")

	ErrLogDriverNotSupported      = New("invalid driver: %s, only support stack, single, daily, custom").SetModule(ModuleLog)
	ErrLogDriverCircularReference = New("%s driver can't include self channel").SetModule(ModuleLog)
	ErrLogEmptyLogFilePath        = New("empty log file path").SetModule(ModuleLog)

	ErrLangFileNotExist = New("translation file does not exist")

	ErrValidationDuplicateFilter = New("duplicate filter name: %s")
	ErrValidationDuplicateRule   = New("duplicate rule name: %s")
	ErrValidationDataInvalidType = New("data must be map[string]any or map[string][]string or struct")
	ErrValidationEmptyData       = New("data can't be empty")
	ErrValidationEmptyRules      = New("rules can't be empty")
)
