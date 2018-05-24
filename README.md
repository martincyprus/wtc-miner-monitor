# wtc-miner-monitor
A Go server client tool for monitoring WTC mining on multiple machines

This tool was written because I needed it for my own usage, I needed to monitor some 10-20 machines mining WTC and it was to cumbersome to login to each machine all the time to verify that they were still mining. I'm happy to add features and fix things if found but my time is limited and it works great for what I wanted it to do. This software is provided as is.

if you find it useful please donate to any of the following crypto addresses:
* WTC (ecr20 token on ethereum for now): 0x4198981c1204B7C26faE857eaBC9fd92fD8D5109 
* BITCOIN: 3PZfyy5Q34dyg56RZvH2RG833VxswEJ4H7
* Litecoin: MWj2FGj4FhSAUaGCvrK9mKz1qTkgYZDJ9y
* Ethereum (or any other coin on the ethereum network): 0x4198981c1204B7C26faE857eaBC9fd92fD8D5109


If you want me to solve a specific problem or expand the feature list a donation to any of the above will go a long way.
I can be found on the walton slack as <b>martin_cy</b> you can also send me messages on reddit also under <b>martin_cy</b> and finally you can find me on telegram with username: <b>@Martin0cy</b> (https://t.me/Martin0cy)

# Features
It is made up of 2 parts, Server and Client, all communcation between server and client is encrypted, it is based on the client will push their information to the server that will collect all the clients' information and give three simple tables as a username/password protected web page where you can see hashrate, latest block number, and number of peers of each node, as of latest report (default every 50-60 seconds).

## Server
The server takes all the messages sent from the clients and put them into a very simple sqlite database, it will keep data according to KeepLogsHours, for the web page to be fast and responsive keep the following in mind, 10 miners, and 10h of data will load in 2 second, double the time or nodes and your time to load will double, its fairly linair. "Node Down" notifications are sent 5 minutes after a nodes last message, between 1-3 messages will be sent in the window of 5-8 minutes, then also another 1-2 messages will be sent 1h after it has been down after that no more notifications will be sent concerning that node. Any time a node reports 0 peers you will get a message every 100 seconds. (this can get very spammy and is on my list of todos). 

the server package will have the following content:
* server.exe
* server-config.json
* db.db

### db.db
this is a simple sqlite database that comes empty with some predefined views and tables.

###server-config.conf 
has the following content:
<pre>
        "MPort": 3333,
        "WEBPORT":8081,
        "EncryptionKey": "blabla1223123456",
        "WEBUsername": "username",
        "WEBPassword":  "password1",
        "UseTelegramBot":"YES",
        "TelegramBotAPIKey":"Full_botID",
        "TelegramChannelID":"your_personal_channelID",
        "Debug":"YES",
		"KeepLogsHours":4
</pre>


#### MPort
this is the port that all clients will communicate to the server, you can pick what ever port you want, just make sure its not used by something else. Also you will need to open up a rule in your firewall to allow TCP inbound traffic on this port on the machine that server will run. There is also integration with TelegramBots so they will ping you if a node goes down or its peer count drops to 0.

#### WEBPORT
This is the port that the server will use to display your statistics

#### EncryptionKey
This key should be your random seed (it has to be same on all clients and server) it need to be at least 16 characters long, change it from the default so you are sure your traffic is kept private.

#### WEBUsername
Username you want to use when you access your webpage

#### WEBPassword
Password to login to the web page with miner statistics

#### UseTelegramBot
If you want to use the telegramBot option for notifications, see bellow for instructions on how to setup your bot. (yes/no). If you do not want to use Telegram the Node Down or 0 Peer count are just written in the server window, also it will show on the web page.

#### TelegramBotAPIKey
Your full telegram API botkey, should look like something like this: bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11 (important include the "bot" in the beginning.)

#### TelegramChannelID
your private channel ID with your notification bot, this can also be a group/channel ids, (group Id's start with "-100" be sure to include the "-" if its a group id)

#### Debug
print out all the data sent to server from clients

#### KeepLogsHours
this configurs when data is cleaned from the database, in hours. Imagine you have 10 nodes, each report to the server every 1 minute, so in 60 minutes you would have 600 entries, in 10h you would have 6000 entries, the statistics page slows down as amount of data goes up as it is a bigger dataset to aggregate and search for the information. I would recommend to keep the total data for a fast responsive page to be around the 4000-6000 entries. then the page should load in sub 2 seconds. Configure according to your needs.

Once you have configured the above parameters you can just double click start the server.exe, if it exists means that some of the parameters might not be configured correctly, please start it manually from a command line and read the output carefully, and correct the problem.

## Client
The client is something you have to run on each node and configure for each specific node. It comes with the following files:
* client.exe
* config.json


The client can track either CPU or GPU hashing, and it also tracks peer count and blockNumber. Running the first time, and assuming it passes the config file validation 3 bat files are created: cpuHash.bat, peer.bat, blockNumber.bat. These are all simple script that will just extract those specific things from your miner and send to the server.

 By default in CPU mode it checks the CPU hashrate by calling the cpuHash.bat about 6 times per minute and then doing an average hash rate and then send it to the server every minute, when running GPU it will sample every second and then calculate the minute average and report that to the server, together with the peer count, and blockNumber.

### config.json
<pre>
        "Id": 1,
        "Name":"comp1",
        "Server": "serverIP:3333",
        "EncryptionKey": "blabla1223123456",
        "Frequency": 60,
        "Mode":"GPU",
        "WaltonClientPath":"C:\\Program Files\\WTC\\walton.exe",
        "GpuMinerLocation":"C:\\Program Files\\WTC\\Walton-GPU-64\\GPUMing_v0.2",
		"RpcPort":8545
</pre>

#### Id
This is the ID of the node, give it positive integer value, and it must be unique among all clients

#### Name
A name for this node, something that make sense to you, maybe the same name you use in your extraData

#### Server
full server ip: e.g. 180.180.180.180:3333 The port must match what you specified in the server-config.json

#### EncryptionKey
This must be identical to the key you specified in the server-config.json

#### Frequency
How often clients should report to server, I recommend to leave it at 60, (have not tested much with other frequencies)

#### Mode
This can be either "CPU" or "GPU" this is case sensitive.

#### WaltonClientPath
This shold be the full path to a walton.exe, make sure to do double "\" e.g. "C:\\Program Files\\WTC\\walton.exe"

#### GpuMinerLocation
this only really matters if you are doing GPU mining. This is the path to where your ming_run.exe lives, and where your "0202001" file lives. when running Multi-GPU its that is just report a static hashrate, this normal mean that this path is pointing to the wrong "0202001" file, please point it to the directoy where the correct 0202001 file is located.

### client.exe
once server is running, and local mining has been started and the config has been filed in just double click it. it will take 60 seconds before you see the first line writing out your hash and peer and blockNumber count, and at that point you will also see a line on the server. Yes of course you can run client.exe on the same machine as you run server.exe.


# Checking things while its running
* In a web browser just go to: http://" the IP of the machine you are running the server on ":" web port number" e.g. http://1.1.1.1:8081
* supply the username/password you put in the server-config.json
* Everytime you reload the page the data is updated as can be seen from the "Current UTC time" changes
You should start seeing 3 tables: 
* The first will contain latest data from all nodes, their hash rate, and peer count, latest blockNumber and when last checkin. 
* The second table will show the average hashrate per node, based on the "KeepLogsHours" number of hours.
* The third table will show your aggregated hashing power for each minute. (node count shows how how many nodes reported in that minute, it is possible it shows more or less then your total node count, it is a known "feature" but mostly it will show correctly)

# Using this monitor for Multi-GPU rigs
It is possible to monitor Multi-GPU setups. Just create a client folder for each GPU, and configure each config.json accordingly, the things to take extra care about is the RpcPort and NodeId/name since it should be different from the other clients, and the RPC port should match what is in the start_gpu.bat for the node you are monitoring.

# Building it yourself
Assuming some basic Go knowledge and a correct setup Go environment.
* go to the aesEncryption folder: go install
* go the wtcPayLoad folder: go install
* go to server folder: 
  * go get 
  * go build
* go to client folder:
  * go get
  * go build

you should now have the executables needed, configure the configs as above.


# Telegram bot
Please start by reading this following intro to telegram bots:
* https://core.telegram.org/bots

next follow this tutorial to create your bot API key and find your channel ID:
* https://www.forsomedefinition.com/automation/creating-telegram-bot-notifications/


# Donations welcome
if you find this software useful please donate crypto to any of the following crypto addresses:
* WTC (ecr20 token on ethereum for now): 0x4198981c1204B7C26faE857eaBC9fd92fD8D5109 
* BITCOIN: 3PZfyy5Q34dyg56RZvH2RG833VxswEJ4H7
* Litecoin: MWj2FGj4FhSAUaGCvrK9mKz1qTkgYZDJ9y
* Ethereum (or any other coin on the ethereum network): 0x4198981c1204B7C26faE857eaBC9fd92fD8D5109


If you want me to solve a specific problem or expand the feature list a donation to any of the above will go a long way.
I can be found on the walton slack as <b>martin_cy</b> you can also send me messages on reddit also under <b>martin_cy</b>. and finally you can find me on telegram with username: <b>@Martin0cy</b> (https://t.me/Martin0cy)
