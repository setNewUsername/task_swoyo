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
	"strconv"
	"syscall"
	"time"
	"main/localdataprovider"
	"main/connhandler"
	"main/remotedataprovider"
	"main/idataprovider"
	"math/rand"
)

//server functions

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

func CreateLocalDataProviderInstance() localdataprovider.LocalDataProvider {

	Result := localdataprovider.LocalDataProvider{
		TokenURLPare:      make(map[string]string),
		TokenTimeStampMap: make(map[string]int64),
		TokenLifeTime:     60,
	}

	return Result
}

func CreateRemoteDataProviderInstance() remotedataprovider.RemoteDataProvider {
	Result := remotedataprovider.RemoteDataProvider{TokenLifeTime: 60}
	
	DBC := remotedataprovider.DataBaseConnector{
		User:     "test_user1",
		Password: "one",
		DBname:   "short_url_db",
		Sslmode:  "disable",
	}
	
	Result.DBConn = DBC.Connect()

	return Result
}

//server functions

func main() {
	rand.Seed(time.Now().UnixNano())

	var DataProvider idataprovider.IDataProvider

	var inputOption string = "none"
	var serverPort int = 0

	flag.StringVar(&inputOption, "d", "storage_local", "information save method")
	flag.IntVar(&serverPort, "p", 8080, "server port")
	flag.Parse()

	switch inputOption {
	case "storage_local":
		fmt.Println("selected local storage method")
		DataProvider = CreateLocalDataProviderInstance()
		break
	case "storage_db":
		fmt.Println("selected database storage method")
		DataProvider = CreateRemoteDataProviderInstance()
		break
	default:
		fmt.Println("wrong storage method selected")
		return
	}

	ConnHan := connhandler.ConnectionHandler{
		Protocol:     "http",
		Host:         "localhost",
		Port:         serverPort,
		DataProvider: DataProvider,
	}

	fmt.Println("starting " + ConnHan.Protocol + "://" + ConnHan.Host + " server at port " + strconv.Itoa(ConnHan.Port))

	log.SetFlags(log.Lshortfile)
	s := &http.Server{Addr: ":" + strconv.Itoa(ConnHan.Port), Handler: ConnHan}
	go start(s)

	stopCh, closeCh := createChannel()
	defer closeCh()
	log.Println("notified:", <-stopCh)

	shutdown(context.Background(), s)
}
