package mysql

import (
	"fmt"
	"github.com/go-xorm/xorm"
	"strconv"
	"time"
)

const (
	GTIDModeOn  = "ON"
	GTIDModeOff = "OFF"
)
const changeMasterToWithAuto = `CHANGE MASTER TO 
    MASTER_HOST = "%s", MASTER_PORT = %d, 
    MASTER_USER = "%s", MASTER_PASSWORD = "%s", 
    MASTER_AUTO_POSITION = 1`

type Variable struct {
	VariableName string `json:"Variable_name"`
	Value        string `json:"Value"`
}

type Executor struct {
	eng *xorm.Engine
}

type MasterStatus struct {
	File            string `json:"File"`
	Position        int    `json:"Position"`
	BinlogDoDB      string `json:"Binlog_Do_DB"`
	BinlogIgnoreDB  string `json:"Binlog_Ignore_DB"`
	ExecutedGTIDSet string `json:"Executed_Gtid_Set"`
}

// SlaveHost  SHOW SLAVE HOSTS result
type SlaveHost struct {
	ServerID  int       `xorm:"Server_id" json:"server_id"`
	Host      string    `xorm:"Host" json:"host"`
	Port      int       `xorm:"Port" json:"port"`
	MasterID  int       `xorm:"Master_id" json:"master_id"`
	SlaveUUID string    `xorm:"Slave_UUID" json:"slave_uuid"`
	LastSeen  time.Time `xorm:"Last_Seen" json:"last_seen"`
}

type ProcessList struct {
	Id      int    `xorm:"Id" json:"Id"`
	User    string `xorm:"User" json:"User"`
	Host    string `xorm:"Host" json:"Host"`
	Db      string `xorm:"db" json:"db"`
	Command string `xorm:"Command" json:"Command"`
	Time    int    `xorm:"Time" json:"Time"`
	State   string `xorm:"State" json:"State"`
	Info    string `xorm:"Info" json:"Info"`
}

type SlaveStatus struct {
	SlaveIORunning      string `json:"Slave_IO_Running"`
	SlaveSQLRunning     string `json:"Slave_SQL_Running"`
	LastIOError         string `json:"Last_IO_Error"`
	LastSQLError        string `json:"Last_SQL_Error"`
	MasterHost          string `json:"Master_Host"`
	MasterUser          string `json:"Master_User"`
	MasterPort          int    `json:"Master_Port"`
	RelayMasterLogFile  string `json:"Relay_Master_Log_File"`
	ExecMasterLogPos    int    `json:"Exec_Master_Log_Pos"`
	RelayLogFile        string `json:"Relay_Log_File"`
	RelayLogPos         int    `json:"Relay_Log_Pos"`
	SecondsBehindMaster int    `json:"Seconds_Behind_Master"`
}

func NewExecutorByEngine(eng *xorm.Engine) *Executor {
	return &Executor{eng: eng}
}

func (e *Executor) MakeBinLogOffBySession() error {
	_, err := e.eng.Exec("SET SQL_LOG_BIN=0")
	return err
}

func (e *Executor) MakeBinLogOnBySession() error {
	_, err := e.eng.Exec("SET SQL_LOG_BIN=1")
	return err
}

func (e *Executor) ShowSlaveStatus() (SlaveStatus, error) {
	var status SlaveStatus
	_, err := e.eng.SQL("SHOW SLAVE STATUS").Get(&status)
	return status, err
}

func (e *Executor) ShowMasterStatus() (MasterStatus, error) {
	var status MasterStatus
	_, err := e.eng.SQL("SHOW MASTER STATUS").Get(&status)
	return status, err
}

func (e *Executor) StartSlave() error {
	_, err := e.eng.Exec("START SLAVE")
	return err
}

func (e *Executor) StopSlave() error {
	_, err := e.eng.Exec("STOP SLAVE")
	return err
}

func (e *Executor) StopSlaveIOThread() error {
	_, err := e.eng.Exec("STOP SLAVE IO_THREAD")
	return err
}

func (e *Executor) ResetSlave() error {
	_, err := e.eng.Exec("RESET SLAVE")
	return err
}

func (e *Executor) ResetSlaveALL() error {
	_, err := e.eng.Exec("RESET SLAVE ALL")
	return err
}

func (e *Executor) ResetMaster() error {
	_, err := e.eng.Exec("RESET MASTER")
	return err
}

func (e *Executor) MysqlGTIDMode() (string, error) {
	var res string = ""
	_, err := e.eng.SQL("SELECT @@gtid_mode").Get(&res)
	if err != nil {
		return "", err
	}
	if res == GTIDModeOn {
		return GTIDModeOn, nil
	}
	return GTIDModeOff, nil
}

func (e *Executor) ServerUUID() (string, error) {
	var res string = ""
	_, err := e.eng.SQL("SELECT @@server_uuid").Get(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (e *Executor) Version() (res string, err error) {
	_, err = e.eng.SQL("SELECT VERSION()").Get(&res)
	return
}

func (e *Executor) ServerID() (id int, err error) {
	var variable Variable
	_, err = e.eng.SQL("SHOW VARIABLES LIKE 'server_id'").Get(&variable)
	if err != nil {
		return
	}
	return strconv.Atoi(variable.Value)
}

func (e *Executor) IsReadOnly() (res bool, err error) {
	var variable Variable
	_, err = e.eng.SQL("SHOW VARIABLES LIKE 'read_only'").Get(&variable)
	if err != nil {
		return false, err
	}
	fmt.Println(variable)
	return variable.Value == "true", nil
}

func (e *Executor) SetReadonlyON() error {
	_, err := e.eng.Exec("SET GLOBAL read_only = ON")
	return err
}

func (e *Executor) SetReadonlyOFF() error {
	_, err := e.eng.Exec("SET GLOBAL read_only = OFF")
	return err
}

func (e *Executor) LockTables() error {
	_, err := e.eng.Exec("FLUSH TABLES WITH READ LOCK")
	return err
}

func (e *Executor) UnlockTables() error {
	_, err := e.eng.Exec("UNLOCK TABLE")
	return err
}

func (e *Executor) MasterPosWait(logFile string, logPos int, timeout int) error {
	sql := fmt.Sprintf("SELECT MASTER_POS_WAIT('%s', %d, %d)", logFile, logPos, timeout)
	_, err := e.eng.Exec(sql)
	return err
}

func (e *Executor) WaitUntilAfterGTIDs(gtids string) error {
	sql := fmt.Sprintf("SELECT WAIT_UNTIL_SQL_THREAD_AFTER_GTIDS('%s')", gtids)
	_, err := e.eng.Exec(sql)
	return err
}

func (e *Executor) ChangeMasterToWithAuto(host string, port int, replUserName, replUserPasswd string) error {
	sql := fmt.Sprintf(changeMasterToWithAuto, host, port, replUserName, replUserPasswd)
	_, err := e.eng.Exec(sql)
	return err
}

func (e *Executor) ShowProcesslist() (processList []ProcessList, err error) {
	err = e.eng.SQL("SHOW PROCESSLIST").Find(&processList)
	return
}

func (e *Executor) ShowSlaveHosts() (res []*SlaveHost, err error) {
	err = e.eng.SQL("SHOW SLAVE HOSTS").Find(&res)
	return
}
