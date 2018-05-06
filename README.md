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
It is made up of 2 parts, Server and Client, all communcation between server and client is encrypted, it is based on the client will push their information to the server that will collect all the clients' information and give two simple tables as username/password protected web page where you can see hashrate of each node, how many peers it has and when it reported in last.

## Server
The server takes all the messages sent from the clients and put them into a very simple sqlite database, it will only keep data for 2 hours, this is configured like this to remain fast and light. "Node Down" notifications are sent 5 minutes after a nodes last message, between 1-3 messages will be sent in the windows of 5-8 minutes, then also another 2-3 messages will be sent 1h after it has been down after that no more notifications will be sent concerning that node. Any time a node reports 0 peers you will get a message every 100 seconds. (this can get very spammy and is on my list of todos). 

the server package will have the following content:
* server.exe
* server-config.json
* db.db

### db.db
this is a simple sqlite database that comes empty with some predefined views and tables.

###server-config.conf 
has the following content:
<pre>
        "MPort": "3333",
        "WEBPORT":"8081",
        "EncryptionKey": "blabla1223123456",
        "WEBUsername": "username",
        "WEBPassword":  "password1",
        "UseTelegramBot":"YES",
        "TelegramBotAPIKey":"Full_botID",
        "TelegramChannelID":"your_personal_channelID"
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
Your full telegram API botkey

#### TelegramChannelID
your private channel ID with your notification bot.

Once you have configured the above parameters you can just double click start the server.exe

## Client
The client is something you have to run on each node and configure for each specific node. It comes with the following files:
* client.exe
* config.json
* cpuHash.bat
* peer.bat

The client can track either CPU or GPU hashing, and it also tracks peer count. By default in CPU mode it checks the CPU hashrate by calling the cpuHash.bat about 6 times per minute and then doing an average hash rate and then send it to the server every minute, when running GPU it will sample every every second and then calculate the minute average and report that to the server, together with the peer count, which it gets by running the peer.bat.

### cpuHash.bat
verify paths and RPC ports, if using different then default

### peer.bat
verify paths and RPC ports, if using different then default

### config.json
<pre>
        "Id": 1,
        "Name":"comp1",
        "Server": "serverIP:3333",
        "EncryptionKey": "blabla1223123456",
        "Frequency": 60,
        "Mode":"GPU",
        "CpuMinerWalletLocation":"C:\\Program Files\\WTC",
        "GpuMinerLocation":"C:\\Program Files\\WTC\\Walton-GPU-64\\GPUMing_v0.2"
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

#### CpuMinerWalletLocation
This should be the path to where your "walton.exe" lives

#### GpuMinerLocation
this only really matters if you are doing GPU mining. This is the path to where your ming_run.exe lives, and where your "0202001" file lives.

### client.exe
once server is running, and local mining has been started and the config has been filed in just double click it. it will take 60 seconds before you see the first line writing out your hash and peer count, and at that point you will also see a line on the server. Yes of course you can run client.exe on the same machine as you run server.exe.


# Checking things while its running
* In a web browser just go to: http://" the IP of the machine you are running the server on ":" web port number" e.g. http://1.1.1.1:8081
* supply the username/password you put in the server-config.json
* Everytime you reload the page the data is updated as can be seen from the "Current UTC time" changes
You should start seeing 2 tables: 
* The left most will contain latest data from all nodes, their hash rate, and peer count and when last they checked in. 
* The right hand table will show your aggregated hashing power for each minute. (node count shows how how many nodes reported in that minute, it is possible it shows more then your total node count, it is a known bug but mostly it will show correctly)

# Using this monitor for Multi-GPU rigs
This should be possible, I dont have one so have not tried it. You will have to run 1 client per GPU and make sure the paths are setup correctly in config.json per client, and that the peer.bat paths are also correct.

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
