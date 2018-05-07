package main

import "net"
import "fmt"

import "bufio"
import "math"

import "os"
import "wtc-miner-monitor/wtcPayload"

import "wtc-miner-monitor/aesEncryption"
import "time"
import "encoding/json"
import "github.com/tkanos/gonfig"
import "strconv"
import "os/exec"
import "strings"

type Configuration struct {
	Id                     int
	Name                   string
	Server                 string
	EncryptionKey          string
	Frequency              int
	Mode                   string
	CpuMinerWalletLocation string
	GpuMinerLocation       string
}

func main() {
	configuration := Configuration{}
	err := gonfig.GetConf("config.json", &configuration)
	if err != nil {
		fmt.Printf("Failed to read the configuration file: %s", err)
		os.Exit(3)
	}
	for {

		createAPackage(configuration)

	}
}

func createAPackage(configuration Configuration) {
	var s wtcPayload.WtcPayload
	s.Hashrate = getHash(configuration)
	conn, err := net.Dial("tcp", configuration.Server)
	if err != nil {
		fmt.Printf("Unable to make a connection object: %s", err)
	}
	key := []byte(configuration.EncryptionKey)
	s.Id = configuration.Id
	s.Name = configuration.Name
	s.Ts = time.Now()
	fmt.Println("HASH: ", s.Hashrate)
	s.Ip = conn.LocalAddr().String()
	s.Peercount = getPeerCount()
	fmt.Println("PEERCOUNT: ", s.Peercount)
	data, err := json.Marshal(s)
	if err != nil {
		fmt.Printf("unable to marshal packat into a json object: %s", err)
	}
	encryptedData, err := aesEncryption.Encrypt(key, string(data))
	if err != nil {
		fmt.Printf("encryption problem: %s", err)
	}

	_, err = conn.Write([]byte(encryptedData))
	if err != nil {
		fmt.Printf("Unable to write data to server %s", err)
	}
}

func getPeerCount() int {
	c, err := exec.Command("peer.bat").Output()

	if err != nil {
		fmt.Println("Error: ", err)
	}
	words := strings.Fields(string(c))
	f, err := strconv.Atoi(words[6])
	if err != nil {
		return 0
	}
	return f
}

func getHash(configuration Configuration) int {

	if configuration.Mode == "CPU" {
		var sum float64 = 0
		for i := 0; i < 6; i++ {
			c, err := exec.Command("cpuHash.bat").Output()
			if err != nil {
				fmt.Println("Error: ", err)
			}
			words := strings.Fields(string(c))
			f, err := strconv.ParseFloat(words[6], 64)
			if err == nil {
				sum += f
				time.Sleep(time.Duration(9) * time.Second)
			} else {

				return 0
			}

		}
		return int(math.RoundToEven(sum / 6))
	}
	if configuration.Mode == "GPU" {
		fmt.Println("Doing GPU")
		var sum float64 = 0
		for i := 0; i < configuration.Frequency; i++ {
			file, err := os.Open(configuration.GpuMinerLocation + "\\0202001")
			if err != nil {
				fmt.Printf("unable to open GPU hashfile %s", err)
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				tex := strings.TrimSpace(scanner.Text())
				if s, err := strconv.ParseFloat(tex, 32); err == nil {
					sum += s
				} else {
					fmt.Println(err)
				}
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
		return int(math.RoundToEven(sum / float64(configuration.Frequency)))
	}

	return 0
}
