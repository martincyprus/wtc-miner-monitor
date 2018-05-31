// buildHtml.go
package main

import (
	"strconv"
	"time"
)

func buildHtml() string {
	bla := getLatestNodeData(Db)
	nodeIDs := getAllNodeIds(Db)
	var html string
	html = `<html> 

	<body> 
 `
	html += `<h1>Latest Hash By Nodes</h1>`
	html += "<h2>Time now: " + time.Now().UTC().Format("2006-01-02 15:04 UTC") + "</h2>"

	html += "<table border = 2 cellpadding=2><tr><th>NodeID</th><th>Name</th><th>h/s</th><th>Peer Count</th><th>Block</th><th>Last datapoint</th></tr>"
	for _, row := range bla {
		html += "<tr>" +
			"<td>" + strconv.Itoa(row.Nodeid) + "</td>" +
			"<td>" + row.Nodename + "</td>" +
			"<td>" + strconv.Itoa(row.Hashrate) + "</td>" +
			"<td>" + strconv.Itoa(row.Peercount) + "</td>" +
			"<td>" + strconv.Itoa(row.Blocknumber) + "</td>" +
			"<td>" + row.Ts.UTC().Format("2006-01-02 15:04 UTC") + "</td>" +
			"</tr>"
	}
	html += `</table>
	<br><br>
		<h1>Average Hashes</h1>`
	averageHashes := getAverageHash(Db)
	html += "<table border = 2 cellpadding=2><tr><th>Nodeid</th><th>Nodename</th><th>Average Hash</th></tr>"
	for _, row := range averageHashes {
		html += "<tr>" +
			"<td>" + strconv.Itoa(row.Nodeid) + "</td>" +
			"<td>" + row.Nodename + "</td>" +
			"<td>" + strconv.FormatFloat(row.Hashrate, 'f', -1, 32) + "</td>" +
			"</tr>"
	}
	html += `</table>
       <br><br>
		<h1>Latest Total Hashes</h1>`
	totalHashes := getLatestTotalHash(Db, Postgres)
	html += "<table border = 2 cellpadding=2><tr><th>Timestamp</th><th>Total h/s</th><th>Number of Nodes(" + strconv.Itoa(len(nodeIDs)) + ")</th></tr>"
	for _, row := range totalHashes {
		html += "<tr>" +
			"<td>" + row.Tstamp + "</td>" +
			"<td>" + strconv.Itoa(row.TotalHash) + "</td>" +
			"<td>" + strconv.Itoa(row.NumberOfNodes) + "</td>" +
			"</tr>"
	}
	html += `</table>`
	html += "</body></html>"
	return html
}

type StatsData struct {
	PageTitle     string
	CurrentTime   string
	TotalHashes   []TotalHash
	AverageHashes []AverageHash
	LatestLogHash []HashlogItem
	AllNodeIds    []int
}

func (h HashlogItem) FormatTimeStamp() string {
	return h.Ts.UTC().Format("2006-01-02 15:04 UTC")
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
	duration := time.Since(h.Ts)
	minutes := time.Duration(duration) * time.Minute

	if minutes > 5 {
		return "red"
	}
	if minutes > 3 {
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
	stats.CurrentTime = time.Now().UTC().Format("2006-01-02 15:04 UTC")
	return stats
}
