package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func CreateRandomString(StringLength int) string {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyz"

	Result := make([]byte, StringLength)

	for i := range Result {
		Result[i] = letters[rand.Intn(len(letters))]
	}

	return string(Result)
}

func createChannel() (chan os.Signal, func()) {
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	return stopCh, func() {
		close(stopCh)
	}
}

func start(server *http.Server) {
	log.Println("application started")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	} else {
		log.Println("application stopped gracefully")
	}
}

func shutdown(ctx context.Context, server *http.Server) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	} else {
		log.Println("application shutdowned")
	}
}

type IDataProvider interface {
	SaveTokenURLPare(FullURL string) string
	GetFullURLbyToken(Token string) string
	CheckTokenTimestamp(Token string) bool
	CheckFullURL(FullURL string) string
	DeleteTokenURLPare(Token string)
}

type LocalDataProvider struct {
	tokenURLPare      map[string]string
	tokenTimeStampMap map[string]int64
	tokenLifeTime     int64
}

func (LDP LocalDataProvider) SaveTokenURLPare(FullURL string) string {
	var Token string = CreateRandomString(32)

	LDP.tokenURLPare[Token] = FullURL
	LDP.tokenTimeStampMap[Token] = time.Now().Unix()

	return Token
}

func (LDP LocalDataProvider) CheckTokenTimestamp(Token string) bool {
	fmt.Println("checking " + Token + " token timestamp")
	var Result bool = true

	if time.Now().Unix()-LDP.tokenTimeStampMap[Token] > LDP.tokenLifeTime {
		fmt.Println(Token + " token expired")
		Result = false
	}

	return Result
}

func (LDP LocalDataProvider) CheckFullURL(FullURL string) string {
	fmt.Println("trying to find token-URL pare")
	var Result string

	for key, value := range LDP.tokenURLPare {
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
	delete(LDP.tokenURLPare, Token)
	delete(LDP.tokenTimeStampMap, Token)
}

func (LDP LocalDataProvider) GetFullURLbyToken(Token string) string {
	fmt.Println("trying to get full URL by token", Token)

	var Result string = "none"

	Result = LDP.tokenURLPare[Token]

	return Result
}

type ConnectionHandler struct {
	protocol     string
	host         string
	port         int
	dataProvider IDataProvider
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

	Token := ConnHand.dataProvider.CheckFullURL(string(UrlBuffer))

	if Token == "" {
		Token = ConnHand.dataProvider.SaveTokenURLPare(string(UrlBuffer))
	}

	writer.Write([]byte(ConnHand.protocol + "://" + ConnHand.host + ":" + strconv.Itoa(ConnHand.port) + "/" + Token))
}

func (ConnHand ConnectionHandler) ServeGET(writer http.ResponseWriter, request *http.Request) {
	Token := strings.ReplaceAll(request.URL.Path, "/", "")

	var FullURL string = ""

	if ConnHand.dataProvider.CheckTokenTimestamp(Token) {
		FullURL = ConnHand.dataProvider.GetFullURLbyToken(Token)
	} else {
		ConnHand.dataProvider.DeleteTokenURLPare(Token)
		writer.Write([]byte("token expired"))
		return
	}

	writer.Write([]byte(FullURL))
}

func main() {
	var inputOption string = "none"
	var serverPort int = 0

	flag.StringVar(&inputOption, "d", "storage_local", "information save method")
	flag.IntVar(&serverPort, "p", 8000, "server port")
	flag.Parse()

	switch inputOption {
	case "storage_local":
		fmt.Println("selected local storage method")
		break
	case "storage_db":
		fmt.Println("selected database storage method")
		break
	default:
		fmt.Println("wrong storage method selected")
		return
	}

	LDP := LocalDataProvider{
		tokenURLPare:      make(map[string]string),
		tokenTimeStampMap: make(map[string]int64),
		tokenLifeTime:     60,
	}

	ConnHan := ConnectionHandler{
		protocol:     "http",
		host:         "localhost",
		port:         serverPort,
		dataProvider: LDP,
	}

	log.SetFlags(log.Lshortfile)
	s := &http.Server{Addr: ":" + strconv.Itoa(ConnHan.port), Handler: ConnHan}
	go start(s)

	stopCh, closeCh := createChannel()
	defer closeCh()
	log.Println("notified:", <-stopCh)

	shutdown(context.Background(), s)
}
