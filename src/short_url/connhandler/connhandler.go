package connhandler

import (
	"fmt"
	"main/idataprovider"
	"net/http"
	"strconv"
	"strings"
)

type ConnectionHandler struct {
	Protocol     string
	Host         string
	Port         int
	DataProvider idataprovider.IDataProvider
}

func (ConnHand ConnectionHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "POST":
		fmt.Println("POST method used")
		ConnHand.ServePOST(writer, request)
		break

	case "GET":
		fmt.Println("GET method used")
		ConnHand.ServeGET(writer, request)
		break

	default:
		writer.Write([]byte(http.ErrNotSupported.ErrorString))
	}
}

func (ConnHand ConnectionHandler) ServePOST(writer http.ResponseWriter, request *http.Request) {
	UrlBuffer := make([]byte, 2048)
	request.Body.Read(UrlBuffer)

	Token := ConnHand.DataProvider.CheckFullURL(string(UrlBuffer))

	if Token == "" {
		Token = ConnHand.DataProvider.SaveTokenURLPare(string(UrlBuffer))
	}

	writer.Write([]byte(ConnHand.Protocol + "://" + ConnHand.Host + ":" + strconv.Itoa(ConnHand.Port) + "/" + Token))
}

func (ConnHand ConnectionHandler) ServeGET(writer http.ResponseWriter, request *http.Request) {
	Token := strings.ReplaceAll(request.URL.Path, "/", "")

	var FullURL string = ""

	if ConnHand.DataProvider.CheckTokenTimestamp(Token) {
		FullURL = ConnHand.DataProvider.GetFullURLbyToken(Token)
	} else {
		ConnHand.DataProvider.DeleteTokenURLPare(Token)
		writer.Write([]byte("token expired"))
		return
	}

	writer.Write([]byte(FullURL))
}
