package testing

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/goravel/framework/contracts/testing/docker"
)

type Fake interface {
	// Instance returns the underlying *gofakeit.Faker instance.
	Instance() *gofakeit.Faker

	// Person & Identity
	Person() *gofakeit.PersonInfo
	Name() string
	NamePrefix() string
	NameSuffix() string
	FirstName() string
	MiddleName() string
	LastName() string
	Gender() string
	Age() int
	SSN() string
	EIN() string

	// Contact & Auth
	Contact() *gofakeit.ContactInfo
	Email() string
	Phone() string
	PhoneFormatted() string
	Username() string
	Password(lower, upper, numeric, special, space bool, num int) string

	// Address & Location
	Address() *gofakeit.AddressInfo
	Street() string
	StreetName() string
	StreetNumber() string
	StreetPrefix() string
	StreetSuffix() string
	City() string
	State() string
	StateAbr() string
	Zip() string
	Country() string
	CountryAbr() string
	Latitude() float64
	LatitudeInRange(min, max float64) (float64, error)
	Longitude() float64
	LongitudeInRange(min, max float64) (float64, error)

	// Company, Job & Finance
	Company() string
	CompanySuffix() string
	BuzzWord() string
	BS() string
	Job() *gofakeit.JobInfo
	JobTitle() string
	JobDescriptor() string
	JobLevel() string
	Currency() *gofakeit.CurrencyInfo
	CurrencyShort() string
	CurrencyLong() string
	Price(min, max float64) float64
	CreditCard() *gofakeit.CreditCardInfo
	CreditCardNumber(cco *gofakeit.CreditCardOptions) string
	CreditCardCvv() string
	CreditCardExp() string
	CreditCardType() string
	BankName() string
	BankType() string
	AchAccount() string
	AchRouting() string
	BitcoinAddress() string

	// Internet, Web & Files
	URL() string
	UrlSlug(words int) string
	DomainName() string
	DomainSuffix() string
	IPv4Address() string
	IPv6Address() string
	MacAddress() string
	UserAgent() string
	HTTPMethod() string
	HTTPStatusCode() int
	HTTPStatusCodeSimple() int
	Color() string
	HexColor() string
	SafeColor() string
	RGBColor() []int
	FileExtension() string
	FileMimeType() string

	// Identifiers & Strings
	UUID() string
	ID() string
	Letter() string
	LetterN(n uint) string
	Vowel() string
	Digit() string
	DigitN(n uint) string
	Numerify(str string) string
	Lexify(str string) string
	Regex(regexStr string) string

	// Words, Sentences & Text
	Word() string
	Sentence(wordCount ...int) string
	Paragraph(paragraphCount ...any) string
	LoremIpsumWord() string
	LoremIpsumSentence(wordCount int) string
	LoremIpsumParagraph(paragraphCount, sentenceCount, wordCount int, separator string) string
	Question() string
	Quote() string
	Phrase() string
	Comment() string
	Slogan() string
	Blurb() string

	// Primitives & Numbers
	Bool() bool
	Number(min, max int) int
	Int() int
	Int8() int8
	Int16() int16
	Int32() int32
	Int64() int64
	IntN(n int) int
	IntRange(min, max int) int
	Uint() uint
	Uint8() uint8
	Uint16() uint16
	Uint32() uint32
	Uint64() uint64
	UintN(n uint) uint
	UintRange(min, max uint) uint
	Float32() float32
	Float32Range(min, max float32) float32
	Float64() float64
	Float64Range(min, max float64) float64
	HexUint(bitSize int) string

	// Dates & Time
	Date() time.Time
	DateRange(start, end time.Time) time.Time
	PastDate() time.Time
	FutureDate() time.Time
	Year() int
	Month() int
	MonthString() string
	WeekDay() string
	Day() int
	Hour() int
	Minute() int
	Second() int
	NanoSecond() int
	TimeZone() string
	TimeZoneAbv() string
	TimeZoneFull() string
	TimeZoneOffset() float32
	TimeZoneRegion() string

	// Random Selection & Shuffling
	RandomString(a []string) string
	RandomInt(i []int) int
	RandomUint(u []uint) uint
	RandomMapKey(mapI any) any
	ShuffleStrings(a []string)
	ShuffleInts(a []int)
	ShuffleAnySlice(v any)
	Weighted(options []any, weights []float32) (any, error)

	// Structured Data & Utilities
	Struct(v any) error
	Slice(v any)
	Map() map[string]any
	Generate(dataVal string) (string, error)
	Template(template string, co *gofakeit.TemplateOptions) (string, error)
	JSON(jo *gofakeit.JSONOptions) ([]byte, error)
	CSV(co *gofakeit.CSVOptions) ([]byte, error)
	XML(xo *gofakeit.XMLOptions) ([]byte, error)
	Seed(args ...any) error
}

type TestingT interface {
	Errorf(format string, args ...any)
	FailNow()
}

type Testing interface {
	// Docker get the Docker instance.
	Docker() Docker
}

type Docker interface {
	// Cache gets a cache connection instance.
	Cache(store ...string) (docker.CacheDriver, error)
	// Database gets a database connection instance.
	Database(connection ...string) (docker.Database, error)
	// Image gets a image instance.
	Image(image docker.Image) docker.ImageDriver
}
