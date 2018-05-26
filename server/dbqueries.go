// dbqueries.go
package main

import (
	"database/sql"
	"strconv"
	"time"
)

type HashlogItem struct {
	Nodeid      int
	Nodename    string
	Ts          time.Time
	Hashrate    int
	Ip          string
	Peercount   int
	Blocknumber int
}

type BlockInfo struct {
}

func getAllNodeIds(db *sql.DB) []int {
	sql_readall := `SELECT distinct nodeid FROM hashlog`

	rows, err := db.Query(sql_readall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []int
	for rows.Next() {
		var nodeid int
		err2 := rows.Scan(&nodeid)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, nodeid)
	}
	return result
}

type TotalHash struct {
	Tstamp        string
	TotalHash     int
	NumberOfNodes int
}

func getLatestTotalHash(db *sql.DB, postgres bool) []TotalHash {

	var sql_readall string
	if postgres {
		sql_readall = `select to_char(ts, 'YYYY-MM-DD HH24:MI')  as tstamp,sum(hashrate) as totalhash,count(*) as numberOfNodes from hashlog group by to_char(ts, 'YYYY-MM-DD HH24:MI') order by to_char(ts, 'YYYY-MM-DD HH24:MI') desc limit 10;`
	} else {
		sql_readall = `select strftime('%Y-%m-%d %H:%M', ts) as tstamp,sum(hashrate) as totalhash,count(*) as numberOfNodes from hashlog group by strftime('%Y-%m-%d %H:%M', ts) order by strftime('%Y-%m-%d %H:%M', ts) desc limit 10;`
	}
	rows, err := db.Query(sql_readall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []TotalHash
	for rows.Next() {
		item := TotalHash{}
		err2 := rows.Scan(&item.Tstamp, &item.TotalHash, &item.NumberOfNodes)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, item)
	}
	return result
}

type AverageHash struct {
	Nodeid   int
	Nodename string
	Hashrate float64
}

func getAverageHash(db *sql.DB) []AverageHash {

	sql_readall := `select nodeid, nodename, round(avg(hashrate),0) from hashlog group by nodeid, nodename order by nodeid asc;`

	rows, err := db.Query(sql_readall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []AverageHash
	for rows.Next() {
		item := AverageHash{}
		err2 := rows.Scan(&item.Nodeid, &item.Nodename, &item.Hashrate)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, item)
	}
	return result
}

func cleanupOldRecords(db *sql.DB, hours int, postgres bool) {
	var sql_readall string
	if postgres {
		sql_readall = `delete from hashlog where ts < now() - interval '` + strconv.Itoa(hours) + ` hours';`
	} else {
		sql_readall = `delete from hashlog where datetime(ts,'utc') <= datetime('now', '-` + strconv.Itoa(hours) + ` hours');`
	}
	stmt, err := db.Prepare(sql_readall)
	checkErr(err)
	_, err = stmt.Exec()
	checkErr(err)
}

func checkForStoppedNodes(db *sql.DB, postgres bool) []HashlogItem {
	var sql_readall string
	if postgres {
		sql_readall = `select * from latest_node_data where (ts <= now() - interval '4 minutes' and ts >= now() - interval '8 minutes') OR (ts >= now() - interval '59 minutes' and ts <= now() - interval '56 minutes')`
	} else {
		sql_readall = `select * from latest_node_data where (datetime(ts,'utc') <= datetime('now', '-4 minutes') and datetime(ts,'utc') >= datetime('now', '-8 minutes')) OR (datetime(ts,'utc') >= datetime('now', '-59 minutes') and datetime(ts,'utc') <= datetime('now', '-56 minutes'));`
	}

	rows, err := db.Query(sql_readall)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []HashlogItem
	for rows.Next() {
		item := HashlogItem{}
		err2 := rows.Scan(&item.Nodeid, &item.Nodename, &item.Ts, &item.Hashrate, &item.Ip, &item.Peercount, &item.Blocknumber)
		if err2 != nil {
			return nil
		}
		result = append(result, item)
	}
	return result
}

func checkForZeroPeers(db *sql.DB) []HashlogItem {
	sql := `select *  from latest_node_data where peercount = 0;`
	rows, err := db.Query(sql)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []HashlogItem
	for rows.Next() {
		item := HashlogItem{}
		err2 := rows.Scan(&item.Nodeid, &item.Nodename, &item.Ts, &item.Hashrate, &item.Ip, &item.Peercount, &item.Blocknumber)
		if err2 != nil {
			return nil
		}
		result = append(result, item)
	}
	return result
}

func getLatestNodeData(db *sql.DB) []HashlogItem {

	sql_readall := `SELECT nodeid,nodename,ts,hashrate,ip,peercount,blocknumber FROM latest_node_data`

	rows, err := db.Query(sql_readall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []HashlogItem
	for rows.Next() {
		item := HashlogItem{}
		err2 := rows.Scan(&item.Nodeid, &item.Nodename, &item.Ts, &item.Hashrate, &item.Ip, &item.Peercount, &item.Blocknumber)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, item)
	}
	return result
}

func getMachineSeries(db *sql.DB, nodeID int, limit int) []HashlogItem {

	sql_readall := "SELECT ts,hashrate FROM hashlog where nodeid =" + strconv.Itoa(nodeID) + " limit " + strconv.Itoa(limit)

	rows, err := db.Query(sql_readall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []HashlogItem
	for rows.Next() {
		item := HashlogItem{}
		err2 := rows.Scan(&item.Ts, &item.Hashrate)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, item)
	}
	return result
}
