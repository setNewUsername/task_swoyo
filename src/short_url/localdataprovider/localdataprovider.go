package localdataprovider

import (
	"fmt"
	"main/common/randstr"
	"time"
)

type LocalDataProvider struct {
	TokenURLPare      map[string]string
	TokenTimeStampMap map[string]int64
	TokenLifeTime     int64
}

func (LDP LocalDataProvider) SaveTokenURLPare(FullURL string) string {
	var Token string = randstr.CreateRandomString(32)

	LDP.TokenURLPare[Token] = FullURL
	LDP.TokenTimeStampMap[Token] = time.Now().Unix()

	return Token
}

func (LDP LocalDataProvider) CheckTokenTimestamp(Token string) bool {
	fmt.Println("checking " + Token + " token timestamp")
	var Result bool = true

	if time.Now().Unix()-LDP.TokenTimeStampMap[Token] > LDP.TokenLifeTime {
		fmt.Println(Token + " token expired")
		Result = false
	}

	return Result
}

func (LDP LocalDataProvider) CheckFullURL(FullURL string) string {
	fmt.Println("trying to find token-URL pare")
	var Result string

	for key, value := range LDP.TokenURLPare {
		if value == FullURL {
			Result = key
			fmt.Println("found returning token")
			break
		}
	}

	return Result
}

func (LDP LocalDataProvider) DeleteTokenURLPare(Token string) {
	fmt.Println("Deleted token ", Token)
	delete(LDP.TokenURLPare, Token)
	delete(LDP.TokenTimeStampMap, Token)
}

func (LDP LocalDataProvider) GetFullURLbyToken(Token string) string {
	fmt.Println("trying to get full URL by token", Token)

	var Result string = "none"

	Result = LDP.TokenURLPare[Token]

	return Result
}
