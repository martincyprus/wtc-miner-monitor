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
	MinerAddress           string
}

func main() {
	configuration := Configuration{}
	err := gonfig.GetConf("config.json", &configuration)
	if err != nil {
		fmt.Printf("Failed to read the configuration file: %s", err)
		os.Exit(3)
	}

	validateClientConfig(configuration)

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

func validateClientConfig(configuration Configuration) {

	// check server connection
	i := strings.Index(configuration.Server, ":")
	if i > -1 {
		serverIP := configuration.Server[:i]
		Port := configuration.Server[i+1:]
		address := net.ParseIP(serverIP)
		if address == nil {
			fmt.Printf("Validation failed on config.json: ServerIP is not an IP number, it is: %v \n", serverIP)
			os.Exit(3)
		}
		if _, err := strconv.ParseInt(Port, 10, 64); err != nil {
			fmt.Printf("Validation failed on config.json: Server Port is not a number, it is: %v \n", Port)
			os.Exit(3)
		}
		conn, err := net.Dial("tcp", configuration.Server)
		if err != nil {
			fmt.Println("Validation failed on config.json: Unable to make a connection to server on IP: " + serverIP + " on Port: " + Port)
			fmt.Println("Verify that the server is running, also verify that this machine can connect to it, does the server allow inbound connections on port:" + Port + "?")
			conn.Close()
			os.Exit(3)
		}
		conn.Close()
	} else {
		fmt.Printf("Validation failed on config.json: Server is not a a valid IP:PORT value, it is: %v \n", configuration.Server)
		os.Exit(3)
	}

	//Check that EncryptionKey is at least 16 characters
	if len(configuration.EncryptionKey) < 16 {
		fmt.Printf("Validation failed on config.json: EncryptionKey must be at least 16 character it is currently only: %v", len(configuration.EncryptionKey))
		os.Exit(3)
	}

	//Check Mode GPU or CPU
	if configuration.Mode == "GPU" {
		fmt.Println("GPU")
		location := configuration.GpuMinerLocation + `\\ming_run.exe`
		if _, err := os.Stat(location); os.IsNotExist(err) {
			fmt.Println(location)
			fmt.Printf(`Validation failed on config.json: ming.exe does not appear to exist at the GpuMinerLocation, please provide a path to the location of ming_run.exe and use double \\ for \ e.g. C:\\Program Files\\WTC\\Walton-GPU-64\\GPUMing_v0.2` + "\n")
			os.Exit(3)
		}
	} else if configuration.Mode == "CPU" {
		fmt.Println("CPU")
		if _, err := os.Stat(configuration.CpuMinerWalletLocation + `\walton.exe`); os.IsNotExist(err) {
			fmt.Printf(`Validation failed on config.json: Walton.exe does not appear to exist at the CpuMinerWalletLocation, please provide a path to the location of walton.exe and use double \\ for \ e.g. C:\\Program Files\\WTC` + "\n")
			os.Exit(3)
		}
	} else {
		fmt.Println("Other")
		fmt.Printf("Validation failed on config.json: Mode must be either GPU or CPU")
		os.Exit(3)
	}

}
