package mihari

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var (
	eg25gModelRegexp             = regexp.MustCompile(`(?P<manufacture>.*)\r\n(?P<model>.*)\r\nRevision: (?P<firmware_revision>.*)\r\n`)
	eg25gQuecCellModeRegexp      = regexp.MustCompile(`(SEARCH|LIMSRV|NOCONN|CONNECT)`)
	eg25gQuecCellRATRegexp       = regexp.MustCompile(`\+QENG: "servingcell","(?P<state>(SEARCH|LIMSRV|NOCONN|CONNECT))","(?P<rat>(GSM|WCDMA|LTE|CDMAHDR|TDSCDMA))"`)
	eg25gQuecCellLTEInfoRegexp   = regexp.MustCompile(`\+QENG: "servingcell","(?P<state>(SEARCH|LIMSRV|NOCONN|CONNECT))","(?P<rat>LTE)","(?P<is_tdd>(TDD|FDD))",(?P<mcc>(-|\d{3})),(?P<mnc>(-|\d+)),(?P<cellid>(-|[0-9A-F]+)),(?P<pcid>(-|\d+)),(?P<earfcn>(-|\d+)),(?P<freq_band_ind>(-|\d+)),(?P<ul_bandwidth>(-|[0-5]{1})),(?P<dl_bandwidth>(-|[0-5]{1})),(?P<tac>(-|\d+)),(?P<rsrp>(-(\d+)?)),(?P<rsrq>(-(\d+)?)),(?P<rssi>(-(\d+)?)),(?P<sinr>(-|\d+)),(?P<srxlev>(-|\d+))`)
	eg25gQuecCellWCDMAInfoRegexp = regexp.MustCompile(`\+QENG: "servingcell","(?P<state>(SEARCH|LIMSRV|NOCONN|CONNECT))","(?P<rat>WCDMA)",(?P<mcc>(-|\d{3})),(?P<mnc>(-|\d+)),(?P<lac>(-|[0-9A-F]+)),(?P<cellid>(-|[0-9A-F]+)),(?P<uarfcn>(-|\d+)),(?P<psc>(-|\d+)),(?P<rac>(-|\d+)),(?P<rscp>(-(\d+)?)),(?P<ecio>(-(\d+)?)),(?P<phych>(-|[0-1]{1})),(?P<sf>(-|[0-8]{1})),(?P<slot>(-|\d+)),(?P<speech_code>(-|\d+)),(?P<com_mod>(-|[0-1]{1}))`)
	eg25gIMEIATCommand           = "AT+CGSN"
	eg25gIMSIATCommand           = "AT+CIMI"
	eg25gICCIDATCommand          = "AT+QCCID"
	eg25gCellInfoCommand         = "AT+QENG=\"servingcell\""
)

var (
	ErrModemNotAttached    = errors.New("modem is not attached")
	ErrModemNoModeReponded = errors.New("queccell mode info was not responded")
	ErrModemNoRATResponded = errors.New("queccell rat info was not responded")
)

func parseModel(buff string) (string, string, string, error) {
	// TODO: supports different types of models
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
	mode := eg25gQuecCellModeRegexp.FindString(buff)
	if len(mode) == 0 {
		return "", ErrModemNoModeReponded
	}
	if mode == QuectelModeSearch {
		return "", ErrModemNotAttached
	}
	rat := eg25gQuecCellRATRegexp.FindStringSubmatch(buff)
	if len(rat) == 0 {
		return "", ErrModemNoRATResponded
	}
	for i, name := range eg25gQuecCellRATRegexp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = rat[i]
		}
	}

	return result["rat"], nil
}

func getWCDMACellInfo(buff string) (WCDMACellInfo, error) {
	var wcdmaCellInfo WCDMACellInfo
	var err error
	match := eg25gQuecCellWCDMAInfoRegexp.FindStringSubmatch(buff)
	result := make(map[string]string)
	if len(match) == 0 {
		return wcdmaCellInfo, fmt.Errorf("queccell mode info was invalid format, %s", fmt.Sprint(buff))
	}
	for i, name := range eg25gQuecCellWCDMAInfoRegexp.SubexpNames() {
		if i != 0 && name != "" {
			// state, rat, mcc, mnc, lac, cellid, uarfcn, psc, rac, rscp, ecio, phych, sf, slot, speech_code, com_mod
			result[name] = match[i]
		}
	}
	wcdmaCellInfo.Timestamp = time.Now().UTC().UnixMilli()
	wcdmaCellInfo.RAT = result["rat"]
	wcdmaCellInfo.State = result["state"]
	if result["mcc"] != "-" {
		wcdmaCellInfo.MCC, err = strconv.Atoi(result["mcc"])
		if err != nil {
			return wcdmaCellInfo, fmt.Errorf("mcc is not numeber, got %s", result["mcc"])
		}
	}
	if result["mnc"] != "-" {
		wcdmaCellInfo.MNC, err = strconv.Atoi(result["mnc"])
		if err != nil {
			return wcdmaCellInfo, fmt.Errorf("mnc is not number, got %s", result["mnc"])
		}
	}
	wcdmaCellInfo.LAC = result["lac"]
	wcdmaCellInfo.CellID = result["cellid"]

	if result["uarfcn"] != "-" {
		wcdmaCellInfo.UARFCN, err = strconv.Atoi(result["uarfcn"])
		if err != nil {
			return wcdmaCellInfo, fmt.Errorf("uarfcn is not number, got %s", result["uarfcn"])
		}
	}
	if result["psc"] != "-" {
		wcdmaCellInfo.PSC, err = strconv.Atoi(result["psc"])
		if err != nil {
			return wcdmaCellInfo, fmt.Errorf("psc is not number, got %s", result["psc"])
		}
	}
	if result["rac"] != "-" {
		wcdmaCellInfo.RAC, err = strconv.Atoi(result["rac"])
		if err != nil {
			return wcdmaCellInfo, fmt.Errorf("rac is not number, got %s", result["rac"])
		}
	}
	if result["rscp"] != "-" {
		wcdmaCellInfo.RSCP, err = strconv.Atoi(result["rscp"])
		if err != nil {
			return wcdmaCellInfo, fmt.Errorf("uarfcn is not number, got %s", result["rscp"])
		}
	}
	if result["ecio"] != "-" {
		wcdmaCellInfo.ECIO, err = strconv.Atoi(result["ecio"])
		if err != nil {
			return wcdmaCellInfo, fmt.Errorf("ecio is not number, got %s", result["ecio"])
		}
	}
	if result["phych"] != "-" {
		wcdmaCellInfo.PhyCh, err = strconv.Atoi(result["phych"])
		if err != nil {
			return wcdmaCellInfo, fmt.Errorf("phych is not number, got %s", result["phych"])
		}
	}
	if result["sf"] != "-" {
		wcdmaCellInfo.SF, err = strconv.Atoi(result["sf"])
		if err != nil {
			return wcdmaCellInfo, fmt.Errorf("sf is not number, got %s", result["sf"])
		}
	}
	if result["slot"] != "-" {
		wcdmaCellInfo.Slot, err = strconv.Atoi(result["slot"])
		if err != nil {
			return wcdmaCellInfo, fmt.Errorf("slot is not number, got %s", result["slot"])
		}
	}
	if result["speech_code"] != "-" {
		wcdmaCellInfo.SpeechCode, err = strconv.Atoi(result["speech_code"])
		if err != nil {
			return wcdmaCellInfo, fmt.Errorf("speech_code is not number, got %s", result["speech_code"])
		}
	}
	if result["com_mod"] != "-" {
		wcdmaCellInfo.ComMod, err = strconv.Atoi(result["com_mod"])
		if err != nil {
			return wcdmaCellInfo, fmt.Errorf("com_mod is not number, got %s", result["com_mod"])
		}
	}

	return wcdmaCellInfo, nil
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
			// state, rat, is_tdd, mcc, mnc, cellid, pcid, earfcn, freq_band_ind, ul_bandwidth, dl_bandwidth, tac, rsrp, rsrq, rssi, sinr, srxlev
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
		lteCellInfo.PCID, err = strconv.Atoi(result["pcid"])
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
