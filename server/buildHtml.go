// buildHtml.go
package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"
)

func buildHtml() string {
	db, err := sql.Open("sqlite3", "./db.db")
	if err != nil {
		fmt.Println("Error opening SQLITE DB: %s", err.Error())
		os.Exit(1)
	}
	bla := getLatestNodeData(db)
	nodeIDs := getAllNodeIds(db)
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
	averageHashes := getAverageHash(db)
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
	totalHashes := getLatestTotalHash(db)
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
