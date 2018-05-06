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
	<head>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/chartist.js/latest/chartist.min.css">
		<link rel="stylesheet" href="http://thisisdallas.github.io/Simple-Grid/simpleGrid.css">
    <script src="https://cdn.jsdelivr.net/chartist.js/latest/chartist.min.js"></script>
  </head>
	<body> 
 <script src='https://cdn.jsdelivr.net/chartist.js/latest/chartist.min.js'></script>`
	html += `<div class="grid grid-pad">
	 <div class="col-1-2">
       <div class="content">
		<h1>Latest Hash By Nodes</h1>`
	html += "<h2>Time now: " + time.Now().UTC().Format("2006-01-02 15:04 UTC") + "</h2>"

	html += "<table border = 2 cellpadding=2><tr><th>NodeID</th><th>Name</th><th>h/s</th><th>Peer Count</th><th>Last datapoint</th></tr>"
	for _, row := range bla {
		html += "<tr>" +
			"<td>" + strconv.Itoa(row.Nodeid) + "</td>" +
			"<td>" + row.Nodename + "</td>" +
			"<td>" + strconv.Itoa(row.Hashrate) + "</td>" +
			"<td>" + strconv.Itoa(row.Peercount) + "</td>" +
			"<td>" + row.Ts.UTC().Format("2006-01-02 15:04 UTC") + "</td>" +
			"</tr>"
	}
	html += `</table>
       </div>
   	 </div>
	<div class="col-1-2">
       <div class="content">
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
	html += `</table>
	</div>
		</div>
	</div>`

	html += "</body></html>"
	return html
}
