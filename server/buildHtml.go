// buildHtml.go
package main

import (
	"fmt"
	"strconv"
	"time"
)

type StatsData struct {
	PageTitle     string
	CurrentTime   string
	TotalHashes   []TotalHash
	AverageHashes []AverageHash
	LatestLogHash []HashlogItem
	AllNodeIds    []int
}

func (h HashlogItem) FormatTimeStamp() string {
	return h.Ts.In(TimeZoneLocation).Format("2006-01-02 15:04 (MST)")
}

func (h TotalHash) FormatTotalHashesTimeStamp() string {
	fmt.Println(h.Tstamp)
	t, err := time.Parse("2018-05-29 13:04", h.Tstamp)
	if err != nil {
		fmt.Println("ERROR", err.Error())
	}

	fmt.Println(t)
	return t.In(TimeZoneLocation).Format("2006-01-02 15:04 (MST)")
}

func (a AverageHash) FormatAvgHash() string {
	return strconv.FormatFloat(a.Hashrate, 'f', -1, 32)
}

func (a StatsData) TotalNumberOfNodes() string {
	return strconv.Itoa(len(a.AllNodeIds))
}

func (h HashlogItem) HashRateColor() string {
	if h.Hashrate == 0 {
		return "red"
	} else {
		return "green"
	}
}

func (h HashlogItem) PeerCountColor() string {
	if h.Peercount >= 24 {
		return "green"
	}
	if h.Peercount >= 10 {
		return "yellow"
	}
	if h.Peercount >= 3 {
		return "orange"
	}
	return "red"
}

func (h HashlogItem) BlockNumberColor() string {
	diff := getLagestBlockNumber() - h.Blocknumber
	if diff > 4 {
		return "red"
	}
	if diff > 3 {
		return "yellow"
	}
	return "green"
}

func (h HashlogItem) TimeStampColor() string {
	duration := time.Since(h.Ts).Minutes()

	if duration > 5 {
		return "red"
	}
	if duration > 3 {
		return "yellow"
	}
	return "green"
}

func getStatsData() StatsData {
	var stats StatsData
	stats.AllNodeIds = getAllNodeIds(Db)
	stats.TotalHashes = getLatestTotalHash(Db, Postgres)
	stats.AverageHashes = getAverageHash(Db)
	stats.LatestLogHash = getLatestNodeData(Db)
	stats.PageTitle = "Statistics"
	stats.CurrentTime = time.Now().In(TimeZoneLocation).Format("2006-01-02 15:04 (MST)")
	return stats
}

const (
	STATSTEMPLATE = `
<html>
<body>
<h1>Latest Hash By Nodes</h1>
<h2>Time now: {{ .CurrentTime }} </h2>
<table border = 2 cellpadding=2>
<tr>
<th>NodeID</th><th>Name</th><th>h/s</th><th>Peer Count</th><th>Block</th><th>Last datapoint</th>
</tr>
	{{range .LatestLogHash}}
		<tr>
			<td>{{.Nodeid}}</td>
			<td>{{.Nodename}}</td>
			<td style="background-color:{{.HashRateColor}}" >{{.Hashrate}}</td>
			<td style="background-color:{{.PeerCountColor}}" >{{.Peercount}}</td>
			<td style="background-color:{{.BlockNumberColor}}" >{{.Blocknumber}}</td>
			<td style="background-color:{{.TimeStampColor}}" >{{.FormatTimeStamp}}</td>
		</tr>
	{{end}}
</table>

<br><br>
<h1>Average Hashes</h1>
<table border = 2 cellpadding=2>
<tr><th>Nodeid</th><th>Nodename</th><th>Average Hash</th></tr>
	{{range .AverageHashes}}
		<tr>
			<td>{{.Nodeid}}</td>
			<td>{{.Nodename}}</td>
			<td>{{.FormatAvgHash}}</td>
		</tr>
	{{end}}
</table>
<br><br>

<h1>Latest Total Hashes</h1>
	<table border = 2 cellpadding=2>
	<tr>	<th>Timestamp</th><th>Total h/s</th><th>Number of Nodes{{.TotalNumberOfNodes}}</th></tr>
		{{range .TotalHashes}}
		<tr>
			<td>{{.FormatTotalHashesTimeStamp}}</td>
			<td>{{.TotalHash}}</td>
			<td>{{.NumberOfNodes}}</td>
			</tr>
			{{end}}
	</table>
</body>
</html>
`
)
