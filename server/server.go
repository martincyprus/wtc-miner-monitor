package main

import (
	subtle "crypto/subtle"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"time"
	"wtc-miner-monitor/aesEncryption"
	"wtc-miner-monitor/wtcPayload"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tkanos/gonfig"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_TYPE = "tcp"
)

type Configuration struct {
	MPort             int
	WEBPORT           int
	EncryptionKey     string
	WEBUsername       string
	WEBPassword       string
	UseTelegramBot    string
	TelegramBotAPIKey string
	TelegramChannelID string
	Debug             string
	KeepLogsHours     int
	UsePostgres       string
}

var Db *sql.DB
var Postgres bool

func main() {

	configuration := Configuration{}
	err := gonfig.GetConf("server-config.json", &configuration)
	if err != nil {
		fmt.Printf("Failed to read the configuration file: %s", err)
		os.Exit(3)
	}

	if strings.ToUpper(configuration.UsePostgres) == "YES" {
		Postgres = true
		fmt.Println("YES")
		connStr := "host=localhost user=postgres dbname=postgres sslmode=disable"
		Db, err = sql.Open("postgres", connStr)
		if err != nil {
			fmt.Println("Error opening SQLITE DB: %s", err.Error())
			os.Exit(1)
		}
		Db = verifyHashDatabaseExists(Db)
	} else {
		Postgres = false
		Db, err = sql.Open("sqlite3", "./db.db")
		if err != nil {
			fmt.Println("Error opening SQLITE DB: %s", err.Error())
			os.Exit(1)
		}
	}

	os.Exit(3)
	validateServerConfig(configuration)

	CONN_PORT := configuration.MPort

	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+strconv.Itoa(CONN_PORT))
	if err != nil {
		fmt.Println("Error starting message listening service: %s", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening for miner connections on Port:" + strconv.Itoa(CONN_PORT))
	fmt.Println("Web page is listening on Port:" + strconv.Itoa(configuration.WEBPORT))

	go func() {
		for {
			time.Sleep(time.Duration(100) * time.Second)
			stoppedNodes := checkForStoppedNodes(Db, Postgres)
			for _, row := range stoppedNodes {
				if strings.ToUpper(configuration.UseTelegramBot) == "YES" {
					nodeIsDownTelegram(row, configuration)
				} else {
					fmt.Println("-----------------------------------------------")
					fmt.Println("---------WAARRNNNINGGGG NODE DOWN--------------")
					fmt.Println("-----------------------------------------------")
					fmt.Printf("NodeID: %v Name: %v Last Seen: %v", row.Nodeid, row.Nodename, row.Ts.UTC().Format("2006-01-02 15:04 UTC\n"))
					fmt.Println("-----------------------------------------------")
					fmt.Println("-----------------------------------------------")

				}
			}
			stoppedNodes = checkForZeroPeers(Db)
			for _, row := range stoppedNodes {
				if strings.ToUpper(configuration.UseTelegramBot) == "YES" {
					zeroPeersTelegram(row, configuration)
				} else {
					fmt.Println("-----------------------------------------------")
					fmt.Println("---------WAARRNNNINGGGG ZERO PEERS-------------")
					fmt.Println("-----------------------------------------------")
					fmt.Printf("NodeID: %v Name: %v Peer count: %v Last Seen: %v", row.Nodeid, row.Nodename, row.Peercount, row.Ts.UTC().Format("2006-01-02 15:04 UTC\n"))
					fmt.Println("-----------------------------------------------")
					fmt.Println("-----------------------------------------------")

				}
			}

			cleanupOldRecords(Db, configuration.KeepLogsHours, Postgres)
		}
	}()

	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				os.Exit(1)
			}
			// Handle connections in a new goroutine.
			go handleRequest(conn, Db, configuration)
		}
	}()

	http.HandleFunc("/", BasicAuth(handle, configuration.WEBUsername, configuration.WEBPassword, "Please enter your username and password for this site"))

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(configuration.WEBPORT), nil))

}

func handle(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, buildHtml())
}

func BasicAuth(handler http.HandlerFunc, username, password, realm string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}

		handler(w, r)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn, Db *sql.DB, configuration Configuration) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(buf)
	if err != nil {
		if err.Error() == "EOF" {
			fmt.Println("Hand shake from client:" + conn.RemoteAddr().String())
		} else {
			fmt.Println("Error reading:")
		}
		conn.Close()
		return
	}
	key := []byte(configuration.EncryptionKey)
	//Decrypt the data
	data, err := aesEncryption.Decrypt(key, string(buf[0:reqLen]))
	if err != nil {
		fmt.Println("Error decrypting:", err.Error())
		conn.Close()
		return

	}

	var s wtcPayload.WtcPayload

	err = json.Unmarshal([]byte(data), &s)
	if err != nil {
		fmt.Println("Error Unmarshalling:", err.Error())
		conn.Close()
		return
	}
	if strings.ToUpper(configuration.Debug) == "YES" {
		fmt.Println(s)
	}
	// Send a response back to person contacting us.
	conn.Write([]byte("Message received."))
	// Close the connection when you're done with it.
	conn.Close()

	stmt, err := Db.Prepare("INSERT INTO hashlog(nodeid, nodename, ts,hashrate,ip,peercount,blocknumber) values(?,?,?,?,?,?,?)")
	if err != nil {
		fmt.Println("Bad values unable to make proper SQL statement error was: ", err)
		return
	}
	stmt.Exec(s.Id, s.Name, s.Ts, s.Hashrate, s.Ip, s.Peercount, s.BlockNumber)
	if err != nil {
		fmt.Println("Unable to insert values to db with following statement: ", err)
		return
	}

}

func validateServerConfig(configuration Configuration) {

	//check MPort
	if configuration.MPort < 1 {
		fmt.Printf("Validation failed on server-config.json: MPort is not a number, it is: %v \n", configuration.MPort)
		os.Exit(3)
	}

	if configuration.WEBPORT < 1 {
		fmt.Printf("Validation failed on server-config.json: WEBPORT is not a number, it is: %v \n", configuration.WEBPORT)
		os.Exit(3)
	}

	//Check that EncryptionKey is at least 16 characters
	if len(configuration.EncryptionKey) < 16 {
		fmt.Printf("Validation failed on server-config.json: EncryptionKey must be at least 16 character it is currently only: %v", len(configuration.EncryptionKey))
		os.Exit(3)
	}

	if len(configuration.WEBUsername) < 4 {
		fmt.Printf("Validation failed on server-config.json: WEBUsername must not be less than 4 character")
		os.Exit(3)
	}

	if len(configuration.WEBPassword) < 4 {
		fmt.Printf("Validation failed on server-config.json: WEBPassword must not be less than 4")
		os.Exit(3)
	}

	if strings.ToUpper(configuration.UseTelegramBot) == "YES" {
		if len(configuration.TelegramBotAPIKey) < 20 {
			fmt.Printf("Validation failed on server-config.json: TelegramBotAPIKey looks too small please check it")
			os.Exit(3)
		}
		if _, err := strconv.ParseInt(configuration.TelegramChannelID, 10, 64); err != nil {
			fmt.Printf("Validation failed on server-config.json: TelegramChannelID is not a number, it is: %v \n", configuration.TelegramChannelID)
			os.Exit(3)

		}
	}
	if configuration.KeepLogsHours < 1 {
		fmt.Printf("Validation failed on server-config.json: KeepLogsHours less then 1, please use a number between 1 and 20, it is: %v \n", configuration.KeepLogsHours)
		os.Exit(3)
	}
}

func verifyHashDatabaseExists(db *sql.DB) *sql.DB {
	fmt.Println("IN IT")
	sqlQ := "SELECT count(*) as dbExists FROM pg_database WHERE datname='hashdb'"
	rows, err := db.Query(sqlQ)
	if err != nil {
		panic(err)
	}
	var dbExists int
	err = rows.Scan(&dbExists)
	if dbExists == 1 {
		fmt.Println("DB exists")
		db.Close()
		connStr := "host=localhost user=postgres dbname=hashdb sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			panic(err)
		}
		return db
	} else {
		fmt.Println("Doesn't exist")
		sqlQ = "create database hashdb;"
		rows, err = db.Query(sqlQ)
		if err != nil {
			panic(err)
		}
		db.Close()
		connStr := "host=localhost user=postgres dbname=hashdb sslmode=disable"
		db, err = sql.Open("postgres", connStr)

		sqlQ = `
CREATE TABLE public.hashlog
(
    id serial,
	nodeid integer NOT NULL,
    nodename text NOT NULL,
    ts timestamp without time zone NOT NULL,
    hashrate integer NOT NULL,
    ip text NOT NULL,
    peercount integer NOT NULL,
    blocknumber integer NOT NULL,
    CONSTRAINT hashlog_pkey PRIMARY KEY (id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.hashlog
    OWNER to postgres;

CREATE INDEX hashlog_name_idx
    ON public.hashlog USING btree
    (nodename COLLATE pg_catalog."default")
    TABLESPACE pg_default;

CREATE INDEX hashlog_nodeid_idx
    ON public.hashlog USING btree
    (nodeid)
    TABLESPACE pg_default;


CREATE INDEX hashlog_ts_idx
    ON public.hashlog USING btree
    (ts)
    TABLESPACE pg_default;
	
CREATE VIEW latest_node_data(nodeid,nodename,ts,hashrate,ip,peercount, blocknumber) 
as SELECT t1.nodeid,t1.nodename,t1.ts,t1.hashrate,t1.ip,t1.peercount, t1.blocknumber 
FROM hashlog t1 
LEFT OUTER JOIN hashlog t2
ON t1.nodeid = t2.nodeid 
AND t1.ts < t2.ts
WHERE t2.nodeid IS NULL
ORDER BY t1.ts;

CREATE or REPLACE VIEW average_hash_by_node(nodeid,nodename,hashrate) as select nodeid, nodename, round(avg(hashrate)) from hashlog group by nodeid,nodename;


`
		_, err = Db.Query(sqlQ)
		if err != nil {
			panic(err)
		}

		return db
	}
}
