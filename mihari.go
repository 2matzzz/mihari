package mihari

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"go.bug.st/serial"
)

var (
	defaultSerialReadTimeout = time.Second
	httpClient               = http.Client{}
	imeiRegexp               = regexp.MustCompile(`[0-9]{15}`)
	imsiRegexp               = regexp.MustCompile(`[0-9]{15}`)
	iccidRegexp              = regexp.MustCompile(`([0-9]{19})F`)
)

const (
	soracomHarvestHost = "http://uni.soracom.io"
	defaultConfigPath  = "/etc/mihari.conf"
)

type Client struct {
	config *Config
	// logger    *log.Logger //TODO: logging
	modem *Modem
	CellInfo
}

type CellInfo interface {
}

type Modem struct {
	serial.Port
	Manufacture      string
	Model            string
	FirmwareRevision string
	IMEI             string
	IMSI             string
	ICCID            string
	RAT              string
}

// type Modems map[string]interface{}
type Config struct {
	ConfigFilePath string
	Verbose        bool
	Name           string `yaml:"name"`
	Path           string `yaml:"path"`
	Interval       int    `yaml:"interval"`
	NewLineCode    string `yaml:"newline_code"`
	Parity         string `yaml:"parity"`
	Stopbits       int    `yaml:"stopbits"`
	Baurdrate      int    `yaml:"baudrate"`
	Databits       int    `yaml:"databits"`
	ReadTimeout    int    `yaml:"read_timeout"`
	Forwarder      string `yaml:"forwarder"`
}

type LTECellInfo struct {
	Timestamp      int64  `json:"timestamp"` // epoch milli sec
	RAT            string `json:"rat"`
	State          string `json:"state"`
	IsTDD          string `json:"is_tdd"`
	MCC            int    `json:"mcc,omitempty"`
	MNC            int    `json:"mnc,omitempty"`
	CellID         string `json:"cellid,omitempty"`
	PhysicalCellID int    `json:"pcid,omitempty"`
	EARFCN         int    `json:"earfcn,omitempty"`
	Band           int    `json:"freq_band_ind,omitempty"`
	ULBandwidth    int    `json:"ul_bandwidth,omitempty"`
	DLBandwidth    int    `json:"dl_bandwidth,omitempty"`
	Tac            int    `json:"tac,omitempty"`
	RSRP           int    `json:"rsrp,omitempty"`
	RSRQ           int    `json:"rsrq,omitempty"`
	RSSI           int    `json:"rssi,omitempty"`
	SINR           int    `json:"sinr,omitempty"`
	Srxlev         int    `json:"srxlev,omitempty"`
}

//TODO: Support 3G
type WCDMACellInfo struct {
	MCC    int
	MNC    int
	CellID int
	Band   int
}

func NewConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); err != nil {
		path = defaultConfigPath
		log.Printf("provided config file is not exist, mihari use default config %s", defaultConfigPath)
	}
	config, err := loadConfig(path)
	if err != nil {
		return config, err
	}
	config.ConfigFilePath = path

	return config, nil
}

func loadConfig(path string) (*Config, error) {
	configData, err := loadConfigFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err := yaml.Unmarshal(configData, config); err != nil {
		return config, fmt.Errorf("config parse error, %s", err)
	}

	if err := config.Validate(); err != nil {
		return config, fmt.Errorf("config validation error, got %s", err)
	}
	return config, nil
}

//TODO: improve validate
func (config *Config) Validate() error {
	if config.Interval < 60 {
		specifiedInterval := config.Interval
		config.Interval = 60
		return fmt.Errorf("interval should specify 60 or over, got %d", specifiedInterval)
	}
	return nil
}

func loadConfigFile(path string) ([]byte, error) {
	configData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return configData, nil
}

func (client *Client) setPort() error {
	var err error
	client.modem.Port, err = serial.Open(client.config.Path, &serial.Mode{})
	if err != nil {
		return fmt.Errorf("%v could not open, %v", client.config.Path, err)
	}

	if err := client.setPortMode(); err != nil {
		return fmt.Errorf("port mode could not set, %s", err)
	} else if client.setPortReadTimeout(); err != nil {
		return fmt.Errorf("port read timeout could not set, %s", err)
	}

	return nil
}

func (client *Client) setPortMode() error {
	mode := &serial.Mode{
		BaudRate: client.config.Baurdrate,
		Parity:   serial.NoParity, //TODO: fix
		DataBits: client.config.Databits,
		StopBits: serial.OneStopBit, //TODO: fix
	}
	if err := client.modem.Port.SetMode(mode); err != nil {
		return err
	}
	return nil
}

func (client *Client) setPortReadTimeout() error {
	var readTimeout time.Duration
	if client.config.ReadTimeout < 1 {
		log.Print("read_timeout is set 1 (sec)")
		readTimeout = defaultSerialReadTimeout
	}
	if err := client.modem.Port.SetReadTimeout(readTimeout); err != nil {
		return err
	}

	return nil
}

func (client *Client) Run() {
	interval := time.Duration(60 * time.Second)
	ticker := time.NewTicker(interval)
	client.Exec()
	for range ticker.C {
		client.Exec()
	}
}

func (client *Client) Exec() {
	// fetch cell info
	if err := client.fetchCellInfo(); err != nil {
		log.Printf("cell info fetch error, %v", err)
	}

	// forward
	body, err := json.Marshal(client.CellInfo)
	if err != nil {
		log.Printf("json error, %v", err)
	}

	switch client.config.Forwarder {
	case "harvest":
		var timestamp int64
		switch client.CellInfo.(type) {
		case LTECellInfo:
			timestamp = client.CellInfo.(LTECellInfo).Timestamp
		}

		//TODO: retry, expnetioal backoff w/ jitter
		req, _ := http.NewRequest("POST", soracomHarvestHost, bytes.NewBuffer(body))
		req.Header.Set("x-soracom-timestamp", fmt.Sprintf("%d", timestamp))
		req.Header.Set("content-type", "application/json")
		resp, _ := httpClient.Do(req)
		ioutil.ReadAll(resp.Body)
		// log.Println(string(body), err)
	}
}

func (client *Client) Close() {
	client.modem.Port.Close()
}

func NewClient(config *Config) (*Client, error) {
	client, err := NewCientWithConfig(config)
	if err != nil {
		return client, err
	}
	return client, nil
}

func NewCientWithConfig(config *Config) (*Client, error) {
	client := &Client{
		config: config,
		modem:  &Modem{},
	}

	if err := client.Check(); err != nil {
		return client, err
	}
	return client, nil
}

func (client *Client) Check() error {
	if err := client.setPort(); err != nil {
		return err
	}

	if err := client.fetchModel(); err != nil {
		return err
	}
	if err := client.fetchIMEI(); err != nil {
		return err
	}
	if err := client.fetchIMSI(); err != nil {
		return err
	}
	if err := client.fetchICCID(); err != nil {
		return err
	}
	if err := client.fetchCellInfo(); err != nil {
		return err
	}

	if err := client.clearPortBuffer(); err != nil {
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

func getQuecCellRAT(buff string) (string, error) {
	result := make(map[string]string)
	modelinfo := eg25gQuecCellModeRegexp.FindStringSubmatch(buff)
	if len(modelinfo) == 0 {
		return "", fmt.Errorf("queccell mode info was not responded")
	}
	for i, name := range eg25gQuecCellModeRegexp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = modelinfo[i]
		}
	}

	return result["rat"], nil
}

func getLTECellInfo(buff string) (LTECellInfo, error) {
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
	lteCellInfo.Timestamp = time.Now().UTC().UnixMilli()
	lteCellInfo.RAT = result["rat"]
	lteCellInfo.State = result["state"]
	lteCellInfo.IsTDD = result["is_tdd"]
	if result["mcc"] != "-" {
		lteCellInfo.MCC, err = strconv.Atoi(result["mcc"])
		if err != nil {
			return lteCellInfo, fmt.Errorf("mcc is not numeber, got %s", result["mcc"])
		}
	}
	if result["mnc"] != "-" {
		lteCellInfo.MNC, err = strconv.Atoi(result["mnc"])
		if err != nil {
			return lteCellInfo, fmt.Errorf("mnc is not number, got %s", result["mnc"])
		}
	}
	if result["cellid"] != "-" {
		lteCellInfo.CellID = result["cellid"]
	}
	if result["freq_band_ind"] != "-" {
		lteCellInfo.Band, err = strconv.Atoi(result["freq_band_ind"])
		if err != nil {
			return lteCellInfo, fmt.Errorf("freq_band_ind is not number, got %s", result["freq_band_ind"])
		}
	}
	if result["pcid"] != "-" {
		lteCellInfo.PhysicalCellID, err = strconv.Atoi(result["pcid"])
		if err != nil {
			return lteCellInfo, fmt.Errorf("pcid is not number, got %s", result["pcid"])
		}
	}
	if result["earfcn"] != "-" {
		lteCellInfo.EARFCN, err = strconv.Atoi(result["earfcn"])
		if err != nil {
			return lteCellInfo, fmt.Errorf("earfcn is not number, got %s", result["earfcn"])
		}
	}
	if result["ul_bandwidth"] != "-" {
		lteCellInfo.ULBandwidth, err = strconv.Atoi(result["ul_bandwidth"])
		if err != nil {
			return lteCellInfo, fmt.Errorf("ul_bandwidth is not number, got %s", result["ul_bandwidth"])
		}
	}
	if result["dl_bandwidth"] != "-" {
		lteCellInfo.DLBandwidth, err = strconv.Atoi(result["dl_bandwidth"])
		if err != nil {
			return lteCellInfo, fmt.Errorf("dl_bandwidth is not number, got %s", result["dl_bandwidth"])
		}
	}
	if result["tac"] != "-" {
		lteCellInfo.Tac, err = strconv.Atoi(result["tac"])
		if err != nil {
			return lteCellInfo, fmt.Errorf("tac is not number, got %s", result["tac"])
		}
	}
	if result["rssi"] != "-" {
		lteCellInfo.RSSI, err = strconv.Atoi(result["rssi"])
		if err != nil {
			return lteCellInfo, fmt.Errorf("rssi is not number, got %s", result["rssi"])
		}
	}
	if result["rsrp"] != "-" {
		lteCellInfo.RSRP, err = strconv.Atoi(result["rsrp"])
		if err != nil {
			return lteCellInfo, fmt.Errorf("rsrp is not number, got %s", result["rsrp"])
		}
	}
	if result["rsrq"] != "-" {
		lteCellInfo.RSRQ, err = strconv.Atoi(result["rsrq"])
		if err != nil {
			return lteCellInfo, fmt.Errorf("rsrq is not number, got %s", result["rsrq"])
		}
	}
	if result["sinr"] != "-" {
		lteCellInfo.SINR, err = strconv.Atoi(result["sinr"])
		if err != nil {
			return lteCellInfo, fmt.Errorf("sinr is not number, got %s", result["sinr"])
		}
	}
	if result["srxlev"] != "-" {
		lteCellInfo.Srxlev, err = strconv.Atoi(result["srxlev"])
		if err != nil {
			return lteCellInfo, fmt.Errorf("srxlev is not number, got %s", result["srxlev"])
		}
	}

	return lteCellInfo, nil
}

func (client *Client) fetchIMEI() error {
	var atCommand string
	switch model := strings.ToLower(client.modem.Model); model {
	case "eg25":
		atCommand = eg25gIMEIATCommand
	}
	_, err := client.modem.Port.Write([]byte(atCommand + "\r\n"))
	if err != nil {
		return fmt.Errorf("%s command fail, %s", atCommand, err)
	}

	//TODO: 100 is enough?
	buff := make([]byte, 100)
	for {
		n, err := client.modem.Port.Read(buff)
		if err != nil || n == 0 {
			return fmt.Errorf("%v is not available", client.config.Path)
		}
		if strings.Contains(string(buff[:n]), "\n") {
			break
		}
	}
	imei, err := parseIMEI(string(buff))
	if err != nil {
		return fmt.Errorf("IMEI parse error, modem response %s, %v", string(buff), err)
	}
	client.modem.IMEI = imei

	return nil
}

func (client *Client) fetchIMSI() error {
	var atCommand string
	switch model := strings.ToLower(client.modem.Model); model {
	case "eg25":
		atCommand = eg25gIMSIATCommand
	}
	_, err := client.modem.Port.Write([]byte(atCommand + "\r\n"))
	if err != nil {
		return fmt.Errorf("%s command fail, %s", atCommand, err)
	}

	//TODO: 100 is enough?
	buff := make([]byte, 100)
	for {
		n, err := client.modem.Port.Read(buff)
		if err != nil || n == 0 {
			return fmt.Errorf("%v is not available", client.config.Path)
		}
		if strings.Contains(string(buff[:n]), "\n") {
			break
		}
	}
	imsi, err := parseIMSI(string(buff))
	if err != nil {
		return fmt.Errorf("IMSI parse error, modem response %s, %v", string(buff), err)
	}

	client.modem.IMSI = imsi

	return nil
}

func (client *Client) fetchICCID() error {
	var atCommand string
	switch model := strings.ToLower(client.modem.Model); model {
	case "eg25":
		atCommand = eg25gICCIDATCommand
	}
	_, err := client.modem.Port.Write([]byte(atCommand + "\r\n"))
	if err != nil {
		return fmt.Errorf("%s command fail, %s", atCommand, err)
	}

	//TODO: 100 is enough?
	buff := make([]byte, 100)
	for {
		n, err := client.modem.Port.Read(buff)
		if err != nil || n == 0 {
			return fmt.Errorf("%v is not available", client.config.Path)
		}
		if strings.Contains(string(buff[:n]), "\n") {
			break
		}
	}
	iccid, err := parseICCID(string(buff))
	if err != nil {
		return fmt.Errorf("ICCID parse error, modem response %s, %v", string(buff), err)
	}
	client.modem.ICCID = iccid

	return nil
}

func (client *Client) fetchCellInfo() error {
	var atCommand string
	model := strings.ToLower(client.modem.Model)
	switch model {
	case "eg25":
		atCommand = eg25gCellInfoCommand
	}
	_, err := client.modem.Port.Write([]byte(atCommand + "\r\n"))
	if err != nil {
		return fmt.Errorf("%s command fail, %s", atCommand, err)
	}

	//TODO 100 is enough?
	buff := make([]byte, 100)
	for {
		n, err := client.modem.Port.Read(buff)
		if err != nil {
			return fmt.Errorf("%s is something went wrong, %#v, %#v, %d byte", client.config.Path, err, string(buff), n)
		}
		if strings.Contains(string(buff[:n]), "\n") {
			break
		}
	}
	err = client.clearPortBuffer()
	if err != nil {
		return err
	}

	switch model {
	case "eg25":
		client.modem.RAT, err = getQuecCellRAT(string(buff))
		if err != nil {
			return err
		}

		switch client.modem.RAT {
		case "LTE":
			lteCellInfo, err := getLTECellInfo(string(buff))
			if err != nil {
				return err
			}
			client.CellInfo = lteCellInfo
		case "WCDMA":
			// wcdmainfo, err := parseWCDMAInfo(string(buff))
		}
		return nil
	default:
		return fmt.Errorf("model is unsupported, got %s", model)
	}
}

func (client *Client) fetchModel() error {
	atCommand := "ATI"
	_, err := client.modem.Port.Write([]byte(atCommand + "\r\n"))
	if err != nil {
		return fmt.Errorf("%s command fail, %s", atCommand, err)
	}

	//TODO: 100 is enough?
	buff := make([]byte, 100)
	for {
		n, err := client.modem.Port.Read(buff)
		if err != nil || n == 0 {
			return fmt.Errorf("%v is not available", client.config.Path)
		}
		if strings.Contains(string(buff[:n]), "\n") {
			break
		}
	}
	client.modem.Manufacture, client.modem.Model, client.modem.FirmwareRevision, err = parseModel(string(buff))
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) clearPortBuffer() error {
	modem := client.modem
	err := modem.Port.ResetInputBuffer()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	err = modem.Port.ResetOutputBuffer()
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
