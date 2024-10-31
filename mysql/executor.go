package mysql

import (
	"fmt"
	"github.com/go-xorm/xorm"
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
	_, err := e.eng.SQL("SET SQL_LOG_BIN=0").Exec()
	return err
}

func (e *Executor) MakeBinLogOnBySession() error {
	_, err := e.eng.SQL("SET SQL_LOG_BIN=1").Exec()
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
	_, err := e.eng.SQL("START SLAVE").Exec()
	return err
}

func (e *Executor) StopSlave() error {
	_, err := e.eng.SQL("STOP SLAVE").Exec()
	return err
}

func (e *Executor) StopSlaveIOThread() error {
	_, err := e.eng.SQL("STOP SLAVE IO_THREAD").Exec()
	return err
}

func (e *Executor) ResetSlave() error {
	_, err := e.eng.SQL("RESET SLAVE").Exec()
	return err
}

func (e *Executor) ResetSlaveALL() error {
	_, err := e.eng.SQL("RESET SLAVE ALL").Exec()
	return err
}

func (e *Executor) ResetMaster() error {
	_, err := e.eng.SQL("RESET MASTER").Exec()
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

func (e *Executor) IsReadOnly() (res bool, err error) {
	var variable Variable
	_, err = e.eng.SQL("SHOW VARIABLES LIKE 'read_only' ").Get(&variable)
	if err != nil {
		return false, err
	}
	fmt.Println(variable)
	return variable.Value == "true", nil
}

func (e *Executor) SetReadonlyON() error {
	_, err := e.eng.SQL("SET GLOBAL read_only = ON").Exec()
	return err
}

func (e *Executor) SetReadonlyOFF() error {
	_, err := e.eng.SQL("SET GLOBAL read_only = OFF").Exec()
	return err
}

func (e *Executor) LockTables() error {
	_, err := e.eng.SQL("FLUSH TABLES WITH READ LOCK").Exec()
	return err
}

func (e *Executor) UnlockTables() error {
	_, err := e.eng.SQL("UNLOCK TABLE").Exec()
	return err
}

func (e *Executor) MasterPosWait(logFile string, logPos int, timeout int) error {
	sql := fmt.Sprintf("SELECT MASTER_POS_WAIT('%s', %d, %d)", logFile, logPos, timeout)
	_, err := e.eng.SQL(sql).Exec()
	return err
}

func (e *Executor) WaitUntilAfterGTIDs(gtids string) error {
	sql := fmt.Sprintf("SELECT WAIT_UNTIL_SQL_THREAD_AFTER_GTIDS('%s')", gtids)
	_, err := e.eng.SQL(sql).Exec()
	return err
}

func (e *Executor) ChangeMasterToWithAuto(host string, port int, replUserName, replUserPasswd string) error {
	sql := fmt.Sprintf(changeMasterToWithAuto, host, port, replUserName, replUserPasswd)
	_, err := e.eng.SQL(sql).Exec()
	return err
}