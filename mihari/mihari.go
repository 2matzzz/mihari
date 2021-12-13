package mihari

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
)

var (
	defaultSerialReadTimeout   = time.Second
	imeiRegexp                 = regexp.MustCompile(`[0-9]{15}`)
	imsiRegexp                 = regexp.MustCompile(`[0-9]{15}`)
	iccidRegexp                = regexp.MustCompile(`([0-9]{19})F`)
	eg25gModelRegexp           = regexp.MustCompile(`(?P<manufacture>.*)\r\n(?P<model>.*)\r\nRevision: (?P<firmware_revision>.*)\r\n`)
	eg25gQuecCellModeRegexp    = regexp.MustCompile(`\+QENG: "servingcell","(?P<state>(SEARCH|LIMSRV|NOCONN|CONNECT))","(?P<rat>(GSM|WCDMA|LTE|CDMAHDR|TDSCDMA))"`)
	eg25gQuecCellLTEInfoRegexp = regexp.MustCompile(`\+QENG: "servingcell","(?P<state>(SEARCH|LIMSRV|NOCONN|CONNECT))","LTE","(?P<is_tdd>(TDD|FDD))",(?P<mcc>(\-|\d{3})),(?P<mnc>(\-|\d+)),(?P<cellid>(\-|\d+)),(?P<pcid>\d+),(?P<earfcn>\d+),(?P<freq_band_ind>\d+),(?P<ul_bandwidth>[0-5]{1}),(?P<dl_bandwidth>[0-5]{1}),(?P<tac>\d+),(?P<rsrp>\-\d+),(?P<rsrq>\-\d+),(?P<rssi>\-\d+),(?P<sinr>\d+),(?P<srxlev>\d+)`) // AT+QENG="servingcell"
	// +QENG: "servingcell","NOCONN","LTE","FDD",440,10,2734811,235,6100,19,3,3,1684,-81,-6,-57,20,50
)

// const defaultConfigPath = "mihari.conf"

type Config struct {
	Path  string
	Ports []Port
}

type Port struct {
	Path string
	serial.Port
	Interval         int
	Manufacture      string
	Model            string
	FirmwareRevision string
	IMEI             string
	IMSI             string
	ICCID            string
	State            string
	NewLineCode      string
	RAT              string
	CellInfo
}

type CellInfo interface {
}

type LTECellInfo struct {
	State          string
	IsTDD          string
	MCC            int
	MNC            int
	CellID         int
	Band           int
	PhysicalCellID int
	EARFCN         int
	ULBandwidth    int
	DLBandwidth    int
	Tac            int
	RSSI           int
	RSRP           int
	RSRQ           int
	SINR           int
	Srxlev         int `json:"srxlev,omitempty"`
}

type WCDMACellInfo struct {
	MCC    int
	MNC    int
	CellID int
	Band   int
}

func NewConfig(path string) (Config, error) {
	var config Config
	var err error

	// TODO
	// Read config file
	// [ec25]
	// path = "/dev/ttyUSB3"
	// interval = 10
	// baudrate = 115200
	// databit = 8
	// parity = none
	// stopbit = 1
	// newline_code = CR/LF/CRLF

	for _, port := range config.Ports {
		if err := port.Check(); err != nil {
			log.Fatalln(err)
		}
	}

	return config, err
}

func (p *Port) Check() error {
	var err error
	p.Port, err = serial.Open(p.Path, &serial.Mode{})
	if err != nil {
		return fmt.Errorf("%v, %v could not open", err, p.Path)
	}
	// defer p.Port.Close()

	mode := &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	if err := p.Port.SetMode(mode); err != nil {
		return err
	}
	if err := p.Port.SetReadTimeout(defaultSerialReadTimeout); err != nil {
		return err
	}
	// if modemStatusBit, err := p.Port.GetModemStatusBits(); err != nil {
	// 	return err
	// } else {
	// 	return fmt.Errorf(fmt.Sprintf("%#v", modemStatusBit))
	// }

	if err := p.GetModel(); err != nil {
		return err
	}
	if err := p.GetIMEI(); err != nil {
		return err
	}
	if err := p.GetIMSI(); err != nil {
		return err
	}
	if err := p.GetICCID(); err != nil {
		return err
	}
	if err := p.GetCellInfo(); err != nil {
		return err
	}

	if err := p.clearPortBuffer(); err != nil {
		return err
	}

	return nil
}

func parseIMEI(buff string) (string, error) {
	match := imeiRegexp.FindAllString(buff, -1)
	if len(match) == 0 {
		return "", fmt.Errorf("IMEI was not responded")
	}
	return match[0], nil
}

func parseIMSI(buff string) (string, error) {
	imsi := imsiRegexp.FindAllString(buff, -1)
	if len(imsi) == 0 {
		return "", fmt.Errorf("IMSI was not responded")
	}
	return imsi[0], nil //TODO improve
}

func parseICCID(buff string) (string, error) {
	iccid := iccidRegexp.FindStringSubmatch(buff)
	if len(iccid) == 0 {
		return "", fmt.Errorf("ICCID was not responded")
	}
	return iccid[1], nil //TODO improve
}

func parseModel(buff string) (string, string, string, error) {
	result := make(map[string]string)
	modelinfo := eg25gModelRegexp.FindStringSubmatch(buff)
	if len(modelinfo) == 0 {
		return "", "", "", fmt.Errorf("model info was not responded, modem responded %s", fmt.Sprint(buff))
	}
	for i, name := range eg25gModelRegexp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = modelinfo[i]
		}
	}

	return result["manufacture"], result["model"], result["firmware_revision"], nil
}

func getQuecCellRATMode(buff string) (string, string, error) {
	result := make(map[string]string)
	modelinfo := eg25gQuecCellModeRegexp.FindStringSubmatch(buff)
	if len(modelinfo) == 0 {
		return "", "", fmt.Errorf("queccell mode info was not responded")
	}
	for i, name := range eg25gQuecCellModeRegexp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = modelinfo[i]
		}
	}

	return result["state"], result["rat"], nil
}

func parseLTECellInfo(buff string) (LTECellInfo, error) {
	// lteInfo := make(map[string]string)
	var lteCellInfo LTECellInfo
	var err error
	match := eg25gQuecCellLTEInfoRegexp.FindStringSubmatch(buff)
	result := make(map[string]string)
	if len(match) == 0 {
		return lteCellInfo, fmt.Errorf("queccell mode info was invalid format, %s", fmt.Sprint(buff))
	}
	for i, name := range eg25gQuecCellLTEInfoRegexp.SubexpNames() {
		if i != 0 && name != "" {
			// state, is_tdd, mcc, mnc, cellid, pcid, earfcn, freq_band_ind, ul_bandwidth, dl_bandwidth, tac, rsrp, rsrq, rssi, sinr, srxlev
			result[name] = match[i]
		}
	}
	lteCellInfo.State = result["state"]
	lteCellInfo.IsTDD = result["is_tdd"]
	lteCellInfo.MCC, err = strconv.Atoi(result["mcc"])
	if err != nil {
		return lteCellInfo, fmt.Errorf("mcc is not numeber, got %s", result["mcc"])
	}
	lteCellInfo.MNC, err = strconv.Atoi(result["mnc"])
	if err != nil {
		return lteCellInfo, fmt.Errorf("mnc is not number, got %s", result["mnc"])
	}
	lteCellInfo.CellID, err = strconv.Atoi(result["cellid"])
	if err != nil {
		return lteCellInfo, fmt.Errorf("cellid is not number, got %s", result["cellid"])
	}
	lteCellInfo.PhysicalCellID, err = strconv.Atoi(result["pcid"])
	if err != nil {
		return lteCellInfo, fmt.Errorf("pcid is not number, got %s", result["pcid"])
	}
	lteCellInfo.EARFCN, err = strconv.Atoi(result["earfcn"])
	if err != nil {
		return lteCellInfo, fmt.Errorf("earfcn is not number, got %s", result["earfcn"])
	}
	lteCellInfo.Band, err = strconv.Atoi(result["freq_band_ind"])
	if err != nil {
		return lteCellInfo, fmt.Errorf("freq_band_ind is not number, got %s", result["freq_band_ind"])
	}
	lteCellInfo.ULBandwidth, err = strconv.Atoi(result["ul_bandwidth"])
	if err != nil {
		return lteCellInfo, fmt.Errorf("ul_bandwidth is not number, got %s", result["ul_bandwidth"])
	}
	lteCellInfo.DLBandwidth, err = strconv.Atoi(result["dl_bandwidth"])
	if err != nil {
		return lteCellInfo, fmt.Errorf("dl_bandwidth is not number, got %s", result["dl_bandwidth"])
	}
	lteCellInfo.Tac, err = strconv.Atoi(result["tac"])
	if err != nil {
		return lteCellInfo, fmt.Errorf("tac is not number, got %s", result["tac"])
	}
	lteCellInfo.RSRP, err = strconv.Atoi(result["rsrp"])
	if err != nil {
		return lteCellInfo, fmt.Errorf("rsrp is not number, got %s", result["rsrp"])
	}
	lteCellInfo.RSRQ, err = strconv.Atoi(result["rsrq"])
	if err != nil {
		return lteCellInfo, fmt.Errorf("rsrq is not number, got %s", result["rsrq"])
	}
	lteCellInfo.RSSI, err = strconv.Atoi(result["rssi"])
	if err != nil {
		return lteCellInfo, fmt.Errorf("rssi is not number, got %s", result["rssi"])
	}
	lteCellInfo.SINR, err = strconv.Atoi(result["sinr"])
	if err != nil {
		return lteCellInfo, fmt.Errorf("sinr is not number, got %s", result["sinr"])
	}
	if result["srxlev"] == "-" {
		lteCellInfo.Srxlev = 0
	} else {
		lteCellInfo.Srxlev, err = strconv.Atoi(result["srxlev"])
		if err != nil {
			return lteCellInfo, fmt.Errorf("srxlev is not number, got %s", result["srxlev"])
		}
	}

	return lteCellInfo, nil
}

func (p *Port) GetIMEI() error {
	_, err := p.Port.Write([]byte("AT+CGSN\r\n"))
	if err != nil {
		return err
	}

	//TODO 100 is enough?
	buff := make([]byte, 100)
	for {
		n, err := p.Port.Read(buff)
		if err != nil || n == 0 {
			return fmt.Errorf("%v is not available", p.Path)
		}
		if strings.Contains(string(buff[:n]), "\n") {
			break
		}
	}
	imei, err := parseIMEI(string(buff))
	if err != nil {
		return err
	}
	p.IMEI = imei

	return nil
}

func (p *Port) GetIMSI() error {
	_, err := p.Port.Write([]byte("AT+CIMI\r\n"))
	if err != nil {
		return err
	}

	//TODO 100 is enough?
	buff := make([]byte, 100)
	for {
		n, err := p.Port.Read(buff)
		if err != nil || n == 0 {
			return fmt.Errorf("%v is not available", p.Path)
		}
		if strings.Contains(string(buff[:n]), "\n") {
			break
		}
	}
	imsi, err := parseIMSI(string(buff))
	if err != nil {
		return err
	}

	p.IMSI = imsi

	return nil
}

func (p *Port) GetICCID() error {
	_, err := p.Port.Write([]byte("AT+QCCID\r\n"))
	if err != nil {
		return err
	}

	//TODO 100 is enough?
	buff := make([]byte, 100)
	for {
		n, err := p.Port.Read(buff)
		if err != nil || n == 0 {
			return fmt.Errorf("%v is not available", p.Path)
		}
		if strings.Contains(string(buff[:n]), "\n") {
			break
		}
	}
	iccid, err := parseICCID(string(buff))
	if err != nil {
		return err
	}

	p.ICCID = iccid

	return nil
}

func (p *Port) GetCellInfo() error {
	_, err := p.Port.Write([]byte("AT+QENG=\"servingcell\"\r\n"))
	if err != nil {
		return err
	}

	//TODO 100 is enough?
	buff := make([]byte, 100)
	for {
		n, err := p.Port.Read(buff)
		if err != nil {
			return fmt.Errorf("%s is something went wrong, %#v, %#v, %d byte", p.Path, err, string(buff), n)
		}
		if strings.Contains(string(buff[:n]), "\n") {
			break
		}
	}
	p.State, p.RAT, err = getQuecCellRATMode(string(buff))
	if err != nil {
		return err
	}

	switch p.RAT {
	case "LTE":
		lteCellInfo, err := parseLTECellInfo(string(buff))
		if err != nil {
			return err
		}
		p.CellInfo = lteCellInfo
	case "WCDMA":
		// wcdmainfo, err := parseWCDMAInfo(string(buff))
	}
	return nil
}

func (p *Port) GetModel() error {
	_, err := p.Port.Write([]byte("ATI\r\n"))
	if err != nil {
		return err
	}

	//TODO 100 is enough?
	buff := make([]byte, 100)
	for {
		n, err := p.Port.Read(buff)
		if err != nil || n == 0 {
			return fmt.Errorf("%v is not available", p.Path)
		}
		if strings.Contains(string(buff[:n]), "\n") {
			break
		}
	}
	p.Manufacture, p.Model, p.FirmwareRevision, err = parseModel(string(buff))
	if err != nil {
		return err
	}
	return nil
}

func (p *Port) clearPortBuffer() error {
	err := p.Port.ResetInputBuffer()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	err = p.Port.ResetOutputBuffer()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

func ListPorts() []string {
	portNames, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(portNames) == 0 {
		log.Fatal("No serial portNames found!")
	}
	return portNames
}
