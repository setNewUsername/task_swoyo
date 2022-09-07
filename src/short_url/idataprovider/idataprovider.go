package idataprovider

type IDataProvider interface {
	SaveTokenURLPare(FullURL string) string
	GetFullURLbyToken(Token string) string
	CheckTokenTimestamp(Token string) bool
	CheckFullURL(FullURL string) string
	DeleteTokenURLPare(Token string)
}
