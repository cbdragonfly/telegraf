package mcis

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf/cloudbarista/usage"
	cbutility "github.com/influxdata/telegraf/cloudbarista/utility"
	"github.com/labstack/echo"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type MCISAgent struct{}

//var Content = usage.CBMCISMetric{}

//Sysbench DB 생성 및 초기화
func (mc *MCISAgent) InitDB(c echo.Context) (*usage.CBMCISMetric, error) {
	if err := checkInit(Sysbench); err != nil {
		return &usage.CBMCISMetric{}, err
	}
	start := time.Now()

	// Init fileio
	cmdStr := "sysbench fileio --file-total-size=50M prepare"
	outputStr, err := cbutility.SysCall(cmdStr)
	if err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("Error in excuting the benchmark: Init fileio " + err.Error()))
	}

	var grepStr = regexp.MustCompile(`([0-9]+) files, ([0-9]+)([a-zA-Z]+) each, ([0-9]+)([a-zA-Z]+) total`)
	parseStr := grepStr.FindStringSubmatch(outputStr)
	if len(parseStr) > 0 {
		parseStr1 := strings.TrimSpace(parseStr[0])
		outputStr = parseStr1
	}
	// Init DB
	cmdStr = "sysbench /usr/share/sysbench/oltp_read_write.lua --db-driver=mysql --table-size=100000 --mysql-db=sysbench --mysql-user=sysbench --mysql-password=psetri1234ak prepare"
	outputStr2, err := cbutility.SysCall(cmdStr)
	if err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("Error in excuting the benchmark: Init DB " + err.Error()))
	}

	grepStr = regexp.MustCompile(` ([0-9]+) records into .([a-zA-Z]+).`)
	parseStr = grepStr.FindStringSubmatch(outputStr2)
	if len(parseStr) > 0 {
		parseStr1 := strings.TrimSpace(parseStr[0])
		outputStr2 = parseStr1
	}

	elapsed := time.Since(start)
	elapsedStr := strconv.FormatFloat(elapsed.Seconds(), 'f', 6, 64)

	outputStr += ", "
	outputStr += outputStr2

	//result = "The init is complete: "
	usage.Content.Result = "The init is complete"
	usage.Content.Elapsed = elapsedStr
	usage.Content.Desc = outputStr + " are created"

	return &usage.Content, nil
}

//Sysbenc DB 삭제
func (mc *MCISAgent) ResetDB(c echo.Context) (*usage.CBMCISMetric, error) {
	if err := checkInit(Sysbench); err != nil {
		return &usage.CBMCISMetric{}, err
	}
	start := time.Now()

	// Clean fileio
	cmdStr := "sysbench fileio --file-total-size=50M cleanup"
	result, err := cbutility.SysCall(cmdStr)
	if err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("message: Error in excuting the benchmark: Clean fileio " + err.Error()))
	}

	// Clean DB
	cmdStr = "sysbench /usr/share/sysbench/oltp_read_write.lua --db-driver=mysql --table-size=100000 --mysql-db=sysbench --mysql-user=sysbench --mysql-password=psetri1234ak cleanup"
	result2, err := cbutility.SysCall(cmdStr)
	if err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("message: Error in excuting the benchmark: Clean DB " + err.Error()))
	}

	elapsed := time.Since(start)
	elapsedStr := strconv.FormatFloat(elapsed.Seconds(), 'f', 6, 64)

	result += result2
	result = "The cleaning is complete"
	usage.Content.Result = result
	usage.Content.Elapsed = elapsedStr
	usage.Content.Desc = "The benchmark files and tables are removed"
	return &usage.Content, nil
}

//CpuM ...
func (mc *MCISAgent) CpuM(c echo.Context) (*usage.CBMCISMetric, error) {
	if err := checkInit(Sysbench); err != nil {
		return &usage.CBMCISMetric{}, err
	}
	cores := strconv.Itoa(cbutility.GetNumCPU())
	start := time.Now()
	cmdStr := "sysbench cpu --cpu-max-prime=100000 --threads=" + cores + " run"
	result, err := cbutility.SysCall(cmdStr)

	elapsed := time.Since(start)
	elapsedStr := strconv.FormatFloat(elapsed.Seconds(), 'f', 6, 64)
	if err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("Error in excuting the benchmark: CPU"))
	}

	var grepStr = regexp.MustCompile(`events per second:(\s+[+-]?([0-9]*[.])?[0-9]+)`)
	//for excution time:`execution time \(avg/stddev\):(\s+[+-]?([0-9]*[.])?[0-9]+)/`

	parseStr := grepStr.FindStringSubmatch(result)
	if len(parseStr) > 0 {
		parseStr1 := strings.TrimSpace(parseStr[1])
		result = parseStr1
	}

	usage.Content.Result = result
	usage.Content.Elapsed = elapsedStr
	usage.Content.Desc = "Repeat the calculation (excution) for prime numbers in 100000 using " + cores + "cores"
	usage.Content.Unit = "Executions/sec"

	return &usage.Content, nil
}

//CpuS ...
func (mc *MCISAgent) CpuS(c echo.Context) (*usage.CBMCISMetric, error) {
	if err := checkInit(Sysbench); err != nil {
		return &usage.CBMCISMetric{}, err
	}
	cores := strconv.Itoa(1)
	start := time.Now()

	cmdStr := "sysbench cpu --cpu-max-prime=100000 --threads=" + cores + " run"
	result, err := cbutility.SysCall(cmdStr)
	elapsed := time.Since(start)
	elapsedStr := strconv.FormatFloat(elapsed.Seconds(), 'f', 6, 64)

	if err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("Error in excuting the benchmark: CPU"))
	}

	var grepStr = regexp.MustCompile(`events per second:(\s+[+-]?([0-9]*[.])?[0-9]+)`)
	//for excution time:`execution time \(avg/stddev\):(\s+[+-]?([0-9]*[.])?[0-9]+)/`

	parseStr := grepStr.FindStringSubmatch(result)
	if len(parseStr) > 0 {
		parseStr1 := strings.TrimSpace(parseStr[1])
		result = parseStr1
	}

	usage.Content.Result = result
	usage.Content.Elapsed = elapsedStr
	usage.Content.Desc = "Repeat the calculation (excution) for prime numbers in 100000 using " + cores + "cores"
	usage.Content.Unit = "Executions/sec"

	return &usage.Content, nil
}

//MemR ...
func (mc *MCISAgent) MemR(c echo.Context) (*usage.CBMCISMetric, error) {
	if err := checkInit(Sysbench); err != nil {
		return &usage.CBMCISMetric{}, err
	}
	start := time.Now()

	cmdStr := "sysbench memory --memory-block-size=1K --memory-scope=global --memory-total-size=10G --memory-oper=read run"
	result, err := cbutility.SysCall(cmdStr)

	elapsed := time.Since(start)
	elapsedStr := strconv.FormatFloat(elapsed.Seconds(), 'f', 6, 64)
	if err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("Error in excuting the benchmark: MEMR"))
	}

	var grepStr = regexp.MustCompile(` transferred .([+-]?([0-9]*[.])?[0-9]+) `)
	parseStr := grepStr.FindStringSubmatch(result)
	if len(parseStr) > 0 {
		parseStr1 := strings.TrimSpace(parseStr[1])
		result = parseStr1
	}

	usage.Content.Result = result
	usage.Content.Elapsed = elapsedStr
	usage.Content.Desc = "Allocate 10G memory buffer and read (repeat reading a pointer)"
	usage.Content.Unit = "MiB/sec"

	return &usage.Content, nil
}

//MemW ...
func (mc *MCISAgent) MemW(c echo.Context) (*usage.CBMCISMetric, error) {
	if err := checkInit(Sysbench); err != nil {
		return &usage.CBMCISMetric{}, err
	}
	start := time.Now()

	cmdStr := "sysbench memory --memory-block-size=1K --memory-scope=global --memory-total-size=10G --memory-oper=write run"
	result, err := cbutility.SysCall(cmdStr)

	elapsed := time.Since(start)
	elapsedStr := strconv.FormatFloat(elapsed.Seconds(), 'f', 6, 64)
	if err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("Error in excuting the benchmark: MEMW"))
	}

	var grepStr = regexp.MustCompile(` transferred .([+-]?([0-9]*[.])?[0-9]+) `)
	parseStr := grepStr.FindStringSubmatch(result)
	if len(parseStr) > 0 {
		parseStr1 := strings.TrimSpace(parseStr[1])
		result = parseStr1
	}

	usage.Content.Result = result
	usage.Content.Elapsed = elapsedStr
	usage.Content.Desc = "Allocate 10G memory buffer and write (repeat writing a pointer)"
	usage.Content.Unit = "MiB/sec"

	return &usage.Content, nil
}

//FioR ...
func (mc *MCISAgent) FioR(c echo.Context) (*usage.CBMCISMetric, error) {
	if err := checkInit(Sysbench); err != nil {
		return &usage.CBMCISMetric{}, err
	}
	start := time.Now()

	cmdStr := "sysbench fileio --file-total-size=50M --file-test-mode=rndrd --max-time=30 --max-requests=0 run"
	result, err := cbutility.SysCall(cmdStr)

	elapsed := time.Since(start)
	elapsedStr := strconv.FormatFloat(elapsed.Seconds(), 'f', 6, 64)
	if err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("Error in excuting the benchmark: FIOR"))
	}

	var grepStr = regexp.MustCompile(`read, MiB/s:(\s+[+-]?([0-9]*[.])?[0-9]+)`)
	parseStr := grepStr.FindStringSubmatch(result)
	if len(parseStr) > 0 {
		parseStr1 := strings.TrimSpace(parseStr[1])
		result = parseStr1
	}

	usage.Content.Result = result
	usage.Content.Elapsed = elapsedStr
	usage.Content.Desc = "Check read throughput by excuting random reads for files in 50MiB for 30s"
	usage.Content.Unit = "MiB/sec"

	return &usage.Content, nil
}

//FioW ...
func (mc *MCISAgent) FioW(c echo.Context) (*usage.CBMCISMetric, error) {
	if err := checkInit(Sysbench); err != nil {
		return &usage.CBMCISMetric{}, err
	}
	start := time.Now()

	cmdStr := "sysbench fileio --file-total-size=50M --file-test-mode=rndwr --max-time=30 --max-requests=0 run"
	result, err := cbutility.SysCall(cmdStr)

	elapsed := time.Since(start)
	elapsedStr := strconv.FormatFloat(elapsed.Seconds(), 'f', 6, 64)
	if err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("Error in excuting the benchmark: FIOW"))
	}

	var grepStr = regexp.MustCompile(`written, MiB/s:(\s+[+-]?([0-9]*[.])?[0-9]+)`)
	parseStr := grepStr.FindStringSubmatch(result)
	if len(parseStr) > 0 {
		parseStr1 := strings.TrimSpace(parseStr[1])
		result = parseStr1
	}

	usage.Content.Result = result
	usage.Content.Elapsed = elapsedStr
	usage.Content.Desc = "Check write throughput by excuting random writes for files in 50MiB for 30s"
	usage.Content.Unit = "MiB/sec"

	return &usage.Content, nil
}

//DBR ...
func (mc *MCISAgent) DBR(c echo.Context) (*usage.CBMCISMetric, error) {
	if err := checkInit(Sysbench); err != nil {
		return &usage.CBMCISMetric{}, err
	}
	start := time.Now()

	cmdStr := "sysbench /usr/share/sysbench/oltp_read_only.lua --db-driver=mysql --table-size=100000 --mysql-db=sysbench --mysql-user=sysbench --mysql-password=psetri1234ak run"
	result, err := cbutility.SysCall(cmdStr)

	elapsed := time.Since(start)
	elapsedStr := strconv.FormatFloat(elapsed.Seconds(), 'f', 6, 64)
	if err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("Error in excuting the benchmark: DBR"))
	}

	var grepStr = regexp.MustCompile(`transactions:(\s+([0-9]*)(\s+)\([+-]?([0-9]*[.])?[0-9]+)`)
	parseStr := grepStr.FindStringSubmatch(result)
	if len(parseStr) > 0 {

		parseStr1 := strings.Split(parseStr[1], "(")
		result = parseStr1[1]
	}

	usage.Content.Result = result
	usage.Content.Elapsed = elapsedStr

	usage.Content.Desc = "Read transactions by simulating transaction loads (OLTP) in DB for 100000 records"
	usage.Content.Unit = "Transactions/s"

	return &usage.Content, nil
}

//DBW ...
func (mc *MCISAgent) DBW(c echo.Context) (*usage.CBMCISMetric, error) {
	if err := checkInit(Sysbench); err != nil {
		return &usage.CBMCISMetric{}, err
	}
	start := time.Now()

	cmdStr := "sysbench /usr/share/sysbench/oltp_write_only.lua --db-driver=mysql --table-size=100000 --mysql-db=sysbench --mysql-user=sysbench --mysql-password=psetri1234ak run"
	result, err := cbutility.SysCall(cmdStr)

	elapsed := time.Since(start)
	elapsedStr := strconv.FormatFloat(elapsed.Seconds(), 'f', 6, 64)
	if err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("Error in excuting the benchmark: DBW"))
	}

	var grepStr = regexp.MustCompile(`transactions:(\s+([0-9]*)(\s+)\([+-]?([0-9]*[.])?[0-9]+)`)
	parseStr := grepStr.FindStringSubmatch(result)
	if len(parseStr) > 0 {

		parseStr1 := strings.Split(parseStr[1], "(")
		result = parseStr1[1]
	}

	usage.Content.Result = result
	usage.Content.Elapsed = elapsedStr
	usage.Content.Desc = "Write transactions by simulating transaction loads (OLTP) in DB for 100000 records"
	usage.Content.Unit = "Transactions/s"

	return &usage.Content, nil
}

//Rtt ...
func (mc *MCISAgent) Rtt(c echo.Context) (*usage.CBMCISMetric, error) {
	if err := checkInit(Ping); err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("Error in excuting ping: install iputils-ping"))
	}
	Req := usage.Request{}
	start := time.Now()
	if err := c.Bind(&Req); err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("Error in request binding " + err.Error()))
	}
	pingHost := Req.Host

	// system call for ping
	cmdStr := "ping -c 10 " + pingHost
	outputStr, err := cbutility.SysCall(cmdStr)
	if err != nil {
		return &usage.CBMCISMetric{}, errors.New(fmt.Sprintf("Error in excuting the benchmark: Ping " + err.Error()))
	}

	var grepStr = regexp.MustCompile(`(\d+.\d+)/(\d+.\d+)/(\d+.\d+)/(\d+.\d+)`)
	parseStr := grepStr.FindAllStringSubmatch(outputStr, -1)
	if len(parseStr) > 0 {
		vals := parseStr[0]
		outputStr = vals[2]
	}
	elapsed := time.Since(start)
	elapsedStr := strconv.FormatFloat(elapsed.Seconds(), 'f', 6, 64)

	usage.Content.Result = outputStr
	usage.Content.Elapsed = elapsedStr
	usage.Content.Desc = "Average RTT to " + pingHost
	usage.Content.Unit = "ms"
	return &usage.Content, nil
}

//Mrtt ...
func (server *MCISAgent) Mrtt(c echo.Context) (*usage.MultiInfo, error) {
	contentArray := usage.MultiInfo{}
	if err := checkInit(Ping); err != nil {
		return &usage.MultiInfo{}, errors.New(fmt.Sprintf("Error in excuting ping: install iputils-ping"))
	}
	mReq := usage.MRequest{}
	start := time.Now()
	if err := c.Bind(&mReq); err != nil {
		return &usage.MultiInfo{}, errors.New(fmt.Sprintf("Error in request binding " + err.Error()))
	}

	hostList := mReq.Multihost
	for _, v := range hostList {

		pingHost := v.Host
		// system call for ping
		cmdStr := "ping -c 10 " + pingHost
		outputStr, err := cbutility.SysCall(cmdStr)
		if err != nil {
			return &usage.MultiInfo{}, errors.New(fmt.Sprintf("Error in excuting the benchmark: Ping " + err.Error()))
		}

		var grepStr = regexp.MustCompile(`(\d+.\d+)/(\d+.\d+)/(\d+.\d+)/(\d+.\d+)`)
		parseStr := grepStr.FindAllStringSubmatch(outputStr, -1)
		if len(parseStr) > 0 {
			vals := parseStr[0]
			outputStr = vals[2]
		}

		elapsed := time.Since(start)
		elapsedStr := strconv.FormatFloat(elapsed.Seconds(), 'f', 6, 64)

		usage.Content.Result = outputStr
		usage.Content.Elapsed = elapsedStr
		usage.Content.Desc = "Average RTT to " + pingHost
		usage.Content.Unit = "ms"
		usage.Content.SpecId = v.Spec

		contentArray.ResultArray = append(contentArray.ResultArray, usage.Content)

	}

	return &contentArray, nil
}

func checkInit(target string) error {
	_, err := cbutility.SysLookPath(target)
	if err != nil {
		return errors.New("Error in excuting the benchmark: not initialized")
	}
	return nil
}

func CleanMCISMetric() {
	usage.Content.Unit = ""
	usage.Content.SpecId = ""
	usage.Content.Elapsed = ""
	usage.Content.Desc = ""
	usage.Content.Result = ""
}

func InitializeMetricList(mc map[string]string) {
	mc[INITDB] = INITDB
	mc[RESETDB] = RESETDB
	mc[CPUS] = CPUS
	mc[CPUM] = CPUM
	mc[MEMR] = MEMR
	mc[MEMW] = MEMW
	mc[FIOR] = FIOR
	mc[FIOW] = FIOW
	mc[DBR] = DBR
	mc[DBW] = DBW
	mc[RTT] = RTT
	mc[MRTT] = MRTT
}
