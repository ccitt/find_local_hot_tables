/**
 * Find localhost mysql hotspot update table
 * version: 0.1a
 * Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
 * Author: zhangli <ccitt@tom.com>
 *
 */
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	//检查运行环境和权限
	if runtime.GOOS != "linux" && os.Getuid() != 0 {
		log.Fatal("You need to run the current tool under root on the Linux system.")
	}

	log.SetOutput(os.Stdout)
}

func main() {
	var (
		dbHost    string
		dbPort    uint
		dbUser    string
		dbPwd     string
		startTime string
		stopTime  string
	)

	const STANDARD_TIME_FORMAT = "2006-01-02 15:04:05"

	//获取命令行传过来的数据库连接相关信息
	flag.StringVar(&dbHost, "h", "127.0.0.1", "Connect to database host.")
	flag.UintVar(&dbPort, "P", 3306, "Connect to database port number.")
	flag.StringVar(&dbUser, "u", "", "Username to use when connecting to database server.")
	flag.StringVar(&dbPwd, "p", "", "Password to use when connecting to database server.")

	//获取命令行传过来的统计范围相关参数信息
	flag.StringVar(&startTime, "start-datetime", time.Now().Add(-time.Minute*10).Format(STANDARD_TIME_FORMAT), "Specify start analysis time(Optional parameter), Default format \"Y-m-d H:i:s\"")
	flag.StringVar(&stopTime, "stop-datetime", time.Now().Format(STANDARD_TIME_FORMAT), "Specify stop analysis time(Optional parameter), Default format \"Y-m-d H:i:s\"")
	//解析命令行参数
	flag.Parse()

	//连接数据库必要参数检查
	if dbHost != "127.0.0.1" {
		log.Fatal("The ip address entered to connect to the database server must be 127.0.0.1.")
	}

	if dbPort < 1 || dbPort > 65535 {
		log.Fatal("Please enter the mysql port to be analyzed between 1-65535!")
	}

	if len(dbUser) == 0 {
		log.Fatal("Please input to connecting database server the Username.")
	}

	if len(dbPwd) == 0 {
		log.Fatal("Please input to connecting database server the Password.")
	}

	FormatedStartTime, _ := time.ParseInLocation(STANDARD_TIME_FORMAT, startTime, time.Local)
	if FormatedStartTime.Format(STANDARD_TIME_FORMAT) != startTime {
		log.Fatal("Incorrect start-datetime parameter or format! Please enter in the following format \"Y-m-d H:i:s\".")
	}

	FormatedStopTime, _ := time.ParseInLocation(STANDARD_TIME_FORMAT, stopTime, time.Local)
	if FormatedStopTime.Format(STANDARD_TIME_FORMAT) != stopTime {
		log.Fatal("Incorrect end-datetime parameter or format! Please enter in the following format \"Y-m-d H:i:s\".")
	}

	//时间逻辑检查,开始时间必须小于结束时间
	if startTime >= stopTime {
		log.Fatal("Please enter a parameter that start-datetime must be less than end-datetime.")
	}

	//为了避免解析文件过多等待时间过长,分析时间范围限制在24小时之内
	if FormatedStopTime.Sub(FormatedStartTime).Hours() > 24 {
		log.Fatal("Please enter start-datetime to end-datetime time range must be less than 24 hours.")
	}

	//链接数据库实例
	dbPortStr := strconv.Itoa(int(dbPort))
	db, err := sql.Open("mysql", dbUser+":"+dbPwd+"@tcp("+dbHost+":"+dbPortStr+")/?charset=utf8mb4,utf8")
	checkErr(err)

	//获取需要分析实例binlog目录路径和前缀名
	var Variable_name, Value string
	err = db.QueryRow("SHOW GLOBAL VARIABLES LIKE 'log_bin_basename'").Scan(&Variable_name, &Value)
	checkErr(err)

	var log_bin_dir, log_bin_basename string
	log_bin_dir, log_bin_basename = path.Split(Value)

	//获取符合分析条件binlog文件列表
	var cmdString, binlogData string
	cmdString = "find " + log_bin_dir + " -name \"" + log_bin_basename + ".*[0-9]\" -type f -newermt \"" + startTime + "\" ! -newermt \"" + stopTime + "\" | sort | tr \"\n\" \" \""
	binlogData = ExecLinuxCommand(cmdString)

	//指定的时间区间内未获取到binlog文件,取stopTime之后最近的一个binlog文件
	if len(binlogData) == 0 {
		cmdString = "find " + log_bin_dir + " -name \"" + log_bin_basename + ".*[0-9]\" -type f -newermt \"" + stopTime + "\" | sort | head -n 1 | tr \"\n\" \" \""
		binlogData = ExecLinuxCommand(cmdString)
	}

	if len(binlogData) == 0 {
		log.Fatal("No matching binlog file was found for the specified time range!")
	}

	fmt.Println("Analyzing localhost mysql " + dbPortStr + " port " + startTime + " to " + stopTime + " binlog file:" + binlogData)
	fmt.Println("Top 20 hot spot tables:")
	cmdString = "mysqlbinlog --no-defaults -v -v --base64-output=DECODE-ROWS --skip-gtids --start-datetime=\"" + startTime + "\" --stop-datetime=\"" + stopTime + "\" " + binlogData + " | grep -Ew '### INSERT|### UPDATE|### DELETE' | awk '/###/ {if($0~/INSERT|UPDATE|DELETE/)count[$2\" \"$NF]++}END{for(i in count) print i,\"\t\",count[i]}' | column -t | sort -k3 -nr | head -n 20"
	analyzedData := ExecLinuxCommand(cmdString)
	fmt.Println(analyzedData)
}

//错误输出检测函数
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//执行外部linux命令函数
func ExecLinuxCommand(strCommand string) string {
	cmd := exec.Command("/bin/bash", "-c", strCommand)

	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		fmt.Println("Execute failed when Start:" + err.Error())
		return ""
	}

	out_bytes, _ := ioutil.ReadAll(stdout)
	stdout.Close()

	if err := cmd.Wait(); err != nil {
		fmt.Println("Execute failed when Wait:" + err.Error())
		return ""
	}
	return string(out_bytes)
}
