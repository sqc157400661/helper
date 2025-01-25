package mysql

import (
	"fmt"
	"log"
)

type Discovery struct {
	executor Executor
}

// Fetches master information from a slave node
func (conn *Discovery) GetMasterInfo() (string, error) {
	var masterHost string
	// 使用xorm查询SHOW SLAVE STATUS
	// Use xorm to query SHOW SLAVE STATUS
	results, err := conn.Engine.Query("SHOW SLAVE STATUS")
	if err != nil {
		return "", err
	}
	if len(results) > 0 {
		masterHost = string(results[0]["Master_Host"])
		return masterHost, nil
	}
	return "", fmt.Errorf("No master found")
}

// GetSlavesInfo 获取从节点信息
// Fetches slave information from a master node
func (conn *MySQLConn) GetSlavesInfo() ([]string, error) {
	var slaves []string
	// 使用xorm查询SHOW PROCESSLIST
	// Use xorm to query SHOW PROCESSLIST
	results, err := conn.Engine.Query("SHOW PROCESSLIST")
	if err != nil {
		return nil, err
	}

	for _, row := range results {
		state := string(row["State"])
		if state == "Slave" {
			slaves = append(slaves, string(row["Host"]))
		}
	}
	return slaves, nil
}

// RecursiveDiscover 拓扑递归发现
// Recursively discover the topology starting from the given node
func (conn *MySQLConn) RecursiveDiscover(nodeHost string) {
	// 获取主节点信息
	// Get master information
	masterHost, err := conn.GetMasterInfo()
	if err != nil {
		log.Printf("Error fetching master info for node %s: %v\n", nodeHost, err)
		return
	}
	log.Printf("Node %s's master is: %s\n", nodeHost, masterHost)

	// 获取从节点信息
	// Get slave nodes information
	slaves, err := conn.GetSlavesInfo()
	if err != nil {
		log.Printf("Error fetching slave info for node %s: %v\n", nodeHost, err)
		return
	}
	log.Printf("Slaves of node %s: %v\n", nodeHost, slaves)

	// 递归地查询每个从节点的拓扑
	// Recursively discover the topology for each slave node
	for _, slave := range slaves {
		conn.RecursiveDiscover(slave)
	}
}

// DiscoverMasterAndSlaves 从从节点获取主节点及所有从节点信息
// Discover the master node and all slaves from a given slave node
func DiscoverMasterAndSlaves(config DBConfig) {
	// 连接到MySQL数据库
	// Connect to MySQL database
	conn, err := ConnectMySQL(config)
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
		return
	}
	defer conn.Engine.Close()

	// 开始递归拓扑发现
	// Start recursive topology discovery
	conn.RecursiveDiscover(config.Host)
}
