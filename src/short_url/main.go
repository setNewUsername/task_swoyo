package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

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

type Token struct {
	token     string
	timeStamp int
}

type LocalDataProvider struct {
	tokenURLPare map[Token]string
}

type IDataProvider interface {
	SaveTokenURLPare()
}

func (LDP LocalDataProvider) SaveTokenURLPare() {

}

type ConnectionHandler struct {
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

func (ConnHand ConnectionHandler) ServeGET(writer http.ResponseWriter, request *http.Request) {

}

func (ConnHand ConnectionHandler) ServePOST(writer http.ResponseWriter, request *http.Request) {

}

func main() {
	var inputOption string = "none"
	var serverPort int = 0

	ConnHan := ConnectionHandler{}

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

	log.SetFlags(log.Lshortfile)
	s := &http.Server{Addr: ":8079", Handler: ConnHan}
	go start(s)

	stopCh, closeCh := createChannel()
	defer closeCh()
	log.Println("notified:", <-stopCh)

	shutdown(context.Background(), s)
}
