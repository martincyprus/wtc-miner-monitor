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
	Id               int
	Name             string
	Server           string
	EncryptionKey    string
	Frequency        int
	Mode             string
	WaltonClientPath string
	GpuMinerLocation string
	RpcPort          int
}

func main() {
	configuration := Configuration{}
	err := gonfig.GetConf("config.json", &configuration)
	if err != nil {
		fmt.Printf("Failed to read the configuration file: %s", err)
		os.Exit(3)
	}

	validateClientConfig(configuration)
	createFiles(configuration)

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
	s.Ip = conn.LocalAddr().String()
	s.Peercount = getPeerCount()
	s.BlockNumber = getBlockNumber()
	fmt.Printf("Block number: %v Peer count: %v Hashrate: %v \n", s.BlockNumber, s.Peercount, s.Hashrate)
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
	p := strings.SplitAfter(string(c), "net.peerCount")[1]
	p = strings.Join(strings.Fields(p), "")
	peerCount, err := strconv.Atoi(p)
	if err != nil {
		return 0
	}
	return peerCount
}

func getBlockNumber() int {
	c, err := exec.Command("blockNumber.bat").Output()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	p := strings.SplitAfter(string(c), "eth.blockNumber")[1]
	p = strings.Join(strings.Fields(p), "")
	blockNumber, err := strconv.Atoi(p)
	if err != nil {
		return 0
	}
	return blockNumber
}

func getHashNumber() int {
	c, err := exec.Command("cpuHash.bat").Output()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	p := strings.SplitAfter(string(c), "eth.hashrate")[1]
	p = strings.Join(strings.Fields(p), "")
	hashrate, err := strconv.Atoi(p)
	if err != nil {
		return 0
	}
	return hashrate
}

func getHash(configuration Configuration) int {

	if configuration.Mode == "CPU" {
		var sum float64 = 0
		for i := 0; i < 6; i++ {
			f := getHashNumber()
			if f != 0 {
				sum += float64(f)
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

	//check that Walton.exe path is valid
	if _, err := os.Stat(configuration.WaltonClientPath); os.IsNotExist(err) {
		fmt.Printf(`Validation failed on config.json: Walton.exe does not appear to exist at the WaltonClientPath, please provide a path to the location of walton.exe and use double \\ for \ e.g. C:\\Program Files\\WTC\\walton.exe` + "\n")
		os.Exit(3)
	}

	//Check Mode GPU or CPU
	if configuration.Mode == "GPU" {
		fmt.Println("GPU")
		location := configuration.GpuMinerLocation
		if _, err := os.Stat(location); os.IsNotExist(err) {
			fmt.Println(location)
			fmt.Printf(`Validation failed on config.json: ming.exe does not appear to exist at the GpuMinerLocation, please provide a path to the location of ming_run.exe and use double \\ for \ e.g. C:\\Program Files\\WTC\\Walton-GPU-64\\GPUMing_v0.2` + "\n")
			os.Exit(3)
		}
	} else if configuration.Mode == "CPU" {
		fmt.Println("CPU")
	} else {
		fmt.Println("Other")
		fmt.Printf("Validation failed on config.json: Mode must be either GPU or CPU")
		os.Exit(3)
	}

	//Check that rpc port is a number and that the rpc server is listening on it
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("", strconv.Itoa(configuration.RpcPort)), time.Duration(1)*time.Second)
	if err != nil {
		fmt.Println("Validation failed on config.json: unable connect to rpc port on local machine", configuration.RpcPort)
		conn.Close()
		os.Exit(3)
	}
	conn.Close()

}

func createFiles(configuration Configuration) {
	createPeerBat(configuration)
	createCpuHashBat(configuration)
	createBlockNumberBat(configuration)
}

func createPeerBat(configuration Configuration) {
	f, err := os.Create("peer.bat")
	if err != nil {
		fmt.Println("ERROR creating files: Unable to create the peer.bat file")
		os.Exit(3)
	}
	w := bufio.NewWriter(f)
	peerBatCommand := `"` + configuration.WaltonClientPath + `" attach http://localhost:` + strconv.Itoa(configuration.RpcPort) + " --exec net.peerCount\n"
	_, err = w.WriteString(peerBatCommand)
	w.Flush()
	f.Close()

}

func createCpuHashBat(configuration Configuration) {
	f, err := os.Create("cpuHash.bat")
	if err != nil {
		fmt.Println("ERROR creating files: Unable to create the cpuHash.bat file")
		os.Exit(3)
	}
	w := bufio.NewWriter(f)
	cpuHashCommand := `"` + configuration.WaltonClientPath + `" attach http://localhost:` + strconv.Itoa(configuration.RpcPort) + " --exec eth.hashrate\n"
	_, err = w.WriteString(cpuHashCommand)
	w.Flush()
	f.Close()
}

func createBlockNumberBat(configuration Configuration) {
	f, err := os.Create("blockNumber.bat")
	if err != nil {
		fmt.Println("ERROR creating files: Unable to create the blockNumber.bat file")
		os.Exit(3)
	}
	w := bufio.NewWriter(f)
	blockNumberCommand := `"` + configuration.WaltonClientPath + `" attach http://localhost:` + strconv.Itoa(configuration.RpcPort) + " --exec eth.blockNumber\n"
	_, err = w.WriteString(blockNumberCommand)
	w.Flush()
	f.Close()

}
