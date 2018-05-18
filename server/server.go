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

	_ "github.com/mattn/go-sqlite3"
	"github.com/tkanos/gonfig"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_TYPE = "tcp"
)

type Configuration struct {
	MPort             string
	WEBPORT           string
	EncryptionKey     string
	WEBUsername       string
	WEBPassword       string
	UseTelegramBot    string
	TelegramBotAPIKey string
	TelegramChannelID string
	Debug             string
}

func main() {
	configuration := Configuration{}
	err := gonfig.GetConf("server-config.json", &configuration)
	if err != nil {
		fmt.Printf("Failed to read the configuration file: %s", err)
		os.Exit(3)
	}

	validateServerConfig(configuration)
	CONN_PORT := configuration.MPort

	db, err := sql.Open("sqlite3", "./db.db")
	if err != nil {
		fmt.Println("Error opening SQLITE DB: %s", err.Error())
		os.Exit(1)
	}

	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error starting message listening service: %s", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening for miner connections on Port:" + CONN_PORT)
	fmt.Println("Web page is listening on Port:" + configuration.WEBPORT)

	go func() {
		for {
			time.Sleep(time.Duration(100) * time.Second)
			stoppedNodes := checkForStoppedNodes(db)
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
			stoppedNodes = checkForZeroPeers(db)
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

			cleanupOldRecords(db)
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
			go handleRequest(conn, db, configuration)
		}
	}()

	http.HandleFunc("/", BasicAuth(handle, configuration.WEBUsername, configuration.WEBPassword, "Please enter your username and password for this site"))

	log.Fatal(http.ListenAndServe(":8081", nil))

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
func handleRequest(conn net.Conn, db *sql.DB, configuration Configuration) {
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

	stmt, err := db.Prepare("INSERT INTO hashlog(nodeid, nodename, ts,hashrate,ip,peercount,blocknumber) values(?,?,?,?,?,?,?)")
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
	if _, err := strconv.ParseInt(configuration.MPort, 10, 64); err != nil {
		fmt.Printf("Validation failed on server-config.json: MPort is not a number, it is: %v \n", configuration.MPort)
		os.Exit(3)
	}

	if _, err := strconv.ParseInt(configuration.WEBPORT, 10, 64); err != nil {
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
}
