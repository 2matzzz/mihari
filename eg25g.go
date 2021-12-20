package mihari

import "regexp"

var (
	eg25gModelRegexp           = regexp.MustCompile(`(?P<manufacture>.*)\r\n(?P<model>.*)\r\nRevision: (?P<firmware_revision>.*)\r\n`)
	eg25gQuecCellModeRegexp    = regexp.MustCompile(`\+QENG: "servingcell","(?P<state>(SEARCH|LIMSRV|NOCONN|CONNECT))","(?P<rat>(GSM|WCDMA|LTE|CDMAHDR|TDSCDMA))"`)
	eg25gQuecCellLTEInfoRegexp = regexp.MustCompile(`\+QENG: "servingcell","(?P<state>(SEARCH|LIMSRV|NOCONN|CONNECT))","(?P<rat>(GSM|WCDMA|LTE|CDMAHDR|TDSCDMA))","(?P<is_tdd>(TDD|FDD))",(?P<mcc>(-|\d{3})),(?P<mnc>(-|\d+)),(?P<cellid>(-|[0-9A-Z]+)),(?P<pcid>(-|\d+)),(?P<earfcn>(-|\d+)),(?P<freq_band_ind>(-|\d+)),(?P<ul_bandwidth>(-|[0-5]{1})),(?P<dl_bandwidth>(-|[0-5]{1})),(?P<tac>(-|\d+)),(?P<rsrp>(-(\d+)?)),(?P<rsrq>(-(\d+)?)),(?P<rssi>(-(\d+)?)),(?P<sinr>(-|\d+)),(?P<srxlev>(-|\d+))`)
	eg25gIMEIATCommand         = "AT+CGSN"
	eg25gIMSIATCommand         = "AT+CIMI"
	eg25gICCIDATCommand        = "AT+QCCID"
	eg25gCellInfoCommand       = "AT+QENG=\"servingcell\""
)
