package remotedataprovider

import (
	"database/sql"
	"fmt"
	"time"
	"main/common/randstr"
	//"reflect"
	//"strconv"
	//"strings"
	_ "github.com/lib/pq"
)

type DataBaseConnector struct {
	User     string
	Password string
	DBname   string
	Sslmode  string
}

func (DBC DataBaseConnector) Connect() *sql.DB {
	connstr := "user=" + DBC.User + " password=" + DBC.Password + " dbname=" + DBC.DBname + " sslmode=" + DBC.Sslmode

	var DBDesc = new(sql.DB)
	var err error

	DBDesc, err = sql.Open("postgres", connstr)
	if err != nil {
		panic(err)
	}
	//defer DBDesc.Close()

	return DBDesc
}

type RemoteDataProvider struct {
	DBConn *sql.DB
	TokenLifeTime int64
}

type Pare struct {
	id int
	fullURL string
	token string
	timestamp int
}

func (RDP RemoteDataProvider) ClearDB() {
	RDP.DBConn.Exec("DELETE from token_url_pare")
}

func (RDP RemoteDataProvider) CloseConnection() {
	RDP.DBConn.Close()
}

func (RDP RemoteDataProvider) SaveTokenURLPare(FullURL string) string {
	Token := randstr.CreateRandomString(32)

	result, err := RDP.DBConn.Exec("INSERT INTO token_url_pare (full_url, token, token_timestamp) VALUES ($1, $2, $3)", FullURL, Token, time.Now().Unix())

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(result.RowsAffected())

	return Token
}

func (RDP RemoteDataProvider) GetFullURLbyToken(Token string) string {
	var Result string

	rows, err := RDP.DBConn.Query("SELECT * FROM token_url_pare WHERE token_url_pare.token = $1", Token)

	if err != nil {
        fmt.Println(err)
    }
    defer rows.Close()

	PareToScan := Pare{}

	for rows.Next() {
		err := rows.Scan(&PareToScan.id, &PareToScan.fullURL, &PareToScan.token, &PareToScan.timestamp)
        if err != nil{
            fmt.Println(err)
            continue
        }
	}

	Result = PareToScan.fullURL

	return Result
}

func (RDP RemoteDataProvider) CheckTokenTimestamp(Token string) bool {
	Result := true

	rows, err := RDP.DBConn.Query("SELECT * FROM token_url_pare WHERE token_url_pare.token = $1", Token)

	if err != nil {
        fmt.Println(err)
    }
    defer rows.Close()

	PareToScan := Pare{}

	for rows.Next() {
		err := rows.Scan(&PareToScan.id, &PareToScan.fullURL, &PareToScan.token, &PareToScan.timestamp)
        if err != nil{
            fmt.Println(err)
            continue
        }
	}

	if time.Now().Unix() - int64(PareToScan.timestamp) > RDP.TokenLifeTime {
		Result = false
	}

	return Result
}

func (RDP RemoteDataProvider) CheckFullURL(FullURL string) string {
	var Result string

	rows, err := RDP.DBConn.Query("SELECT * FROM token_url_pare WHERE token_url_pare.full_url = $1", FullURL)

	if err != nil {
        fmt.Println(err)
    }
    defer rows.Close()

	PareToScan := Pare{}

	for rows.Next() {
		err := rows.Scan(&PareToScan.id, &PareToScan.fullURL, &PareToScan.token, &PareToScan.timestamp)
        if err != nil{
            fmt.Println(err)
            continue
        }
	}

	Result = PareToScan.token

	return Result
}

func (RDP RemoteDataProvider) DeleteTokenURLPare(Token string) {
	result, err := RDP.DBConn.Exec("DELETE FROM token_url_pare WHERE token = $1", Token)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(result.RowsAffected())
}