package mihari

import (
	"reflect"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	config, err := NewConfig("example/mihari.yml")
	if err != nil {
		t.Errorf("%v", err)
	}

	t.Logf("%#v", config)
}

//TODO: mock
// func TestNewClient(t *testing.T) {
// 	config, err := NewConfig("example/mihari.yml")
// 	if err != nil {
// 		t.Errorf("%v", err)
// 	}
// 	client, err := NewClient(config)
// 	if err != nil {
// 		t.Errorf("%v", err)
// 	}
// 	defer client.Close()
// }

func TestGetQuecCellRATMode(t *testing.T) {
	validCommandResp := []byte(`+QENG: "servingcell","NOCONN","LTE","FDD",440,10,2734811,235,6100,19,3,3,1684,-81,-10,-54,19,50`)
	rat, err := getQuecCellRAT(string(validCommandResp))
	if err != nil {
		t.Errorf("%v", err)
	}
	if rat == "" {
		t.Errorf("queccell command response parse error, expected \"LTE\" but actual \"%s\"", rat)
	}

	invalidStateCommandResp := []byte(`+QENG: "servingcell","INVALID","INVALID","FDD",440,10,2734811,235,6100,19,3,3,1684,-81,-10,-54,19,50`)
	rat, err = getQuecCellRAT(string(invalidStateCommandResp))
	if err == nil || rat != "" {
		t.Errorf("queccell command response parse failed, this test case should be fail but succeed")
	}
}

func TestGetLTECellInfo(t *testing.T) {
	type getLTECellInfoTestCase struct {
		ATCommandResp string
		LTECellInfo
	}
	//TODO: flaky test code, it can fail because the parsed timing and the time to run the test are not exactly the same
	now := time.Now().UTC().UnixMilli()
	var getLTECellInfoTestCases = []getLTECellInfoTestCase{
		{
			ATCommandResp: `+QENG: "servingcell","NOCONN","LTE","FDD",440,10,2734811,235,6100,19,3,3,1684,-81,-10,-54,19,50`,
			LTECellInfo: LTECellInfo{
				Timestamp:   now,
				RAT:         "LTE",
				State:       "NOCONN",
				IsTDD:       "FDD",
				MCC:         440,
				MNC:         10,
				CellID:      "2734811",
				PCID:        235,
				EARFCN:      6100,
				Band:        19,
				ULBandwidth: 3,
				DLBandwidth: 3,
				Tac:         1684,
				RSRP:        -81,
				RSRQ:        -10,
				RSSI:        -54,
				SINR:        19,
				Srxlev:      50,
			},
		},
		{
			ATCommandResp: `+QENG: "servingcell","NOCONN","LTE","FDD",-,-,-,-,-,-,-,-,-,-,-,-,-,-`,
			LTECellInfo: LTECellInfo{
				Timestamp:   now,
				RAT:         "LTE",
				State:       "NOCONN",
				IsTDD:       "FDD",
				MCC:         0,
				MNC:         0,
				CellID:      "",
				PCID:        0,
				EARFCN:      0,
				Band:        0,
				ULBandwidth: 0,
				DLBandwidth: 0,
				Tac:         0,
				RSRP:        0,
				RSRQ:        0,
				RSSI:        0,
				SINR:        0,
				Srxlev:      0,
			},
		},
		{
			ATCommandResp: `+QENG: "servingcell","NOCONN","LTE","FDD",440,10,2C81000,193,1850,3,5,5,1694,-100,-11,-67,11,29`,
			LTECellInfo: LTECellInfo{
				Timestamp:   now,
				RAT:         "LTE",
				State:       "NOCONN",
				IsTDD:       "FDD",
				MCC:         440,
				MNC:         10,
				CellID:      "2C81000",
				PCID:        193,
				EARFCN:      1850,
				Band:        3,
				ULBandwidth: 5,
				DLBandwidth: 5,
				Tac:         1694,
				RSRP:        -100,
				RSRQ:        -11,
				RSSI:        -67,
				SINR:        11,
				Srxlev:      29,
			},
		},
	}
	for _, expect := range getLTECellInfoTestCases {
		actual, err := getLTECellInfo(string(expect.ATCommandResp))
		if err != nil {
			t.Errorf("%v", err)
		}
		if actual != expect.LTECellInfo {
			t.Errorf("got unexpected result, expected: %v, but got %v", expect.LTECellInfo, actual)
		}
	}

}

func TestGetWCDMACellInfo(t *testing.T) {
	type getWCDMACellInfoTestCase struct {
		ATCommandResp string
		WCDMACellInfo
	}
	//TODO: flaky test code, it can fail because the parsed timing and the time to run the test are not exactly the same
	now := time.Now().UTC().UnixMilli()
	var getWCDMACellInfoTestCases = []getWCDMACellInfoTestCase{
		{
			ATCommandResp: `+QENG: "servingcell","NOCONN","WCDMA",440,10,75,6FE0090,10736,58,0,-78,-3,-,-,-,-,-`,
			WCDMACellInfo: WCDMACellInfo{
				Timestamp:  now,
				RAT:        "WCDMA",
				State:      "NOCONN",
				MCC:        440,
				MNC:        10,
				LAC:        "75",
				CellID:     "6FE0090",
				UARFCN:     10736,
				PSC:        58,
				RAC:        0,
				RSCP:       -78,
				ECIO:       -3,
				PhyCh:      0,
				SF:         0,
				Slot:       0,
				SpeechCode: 0,
				ComMod:     0,
			},
		},
		{
			ATCommandResp: `+QENG: "servingcell","NOCONN","WCDMA",440,20,0,ABCDEFF,0,-,-,-,-,-,-,-,-,-`,
			WCDMACellInfo: WCDMACellInfo{
				Timestamp:  now,
				RAT:        "WCDMA",
				State:      "NOCONN",
				MCC:        440,
				MNC:        20,
				LAC:        "0",
				CellID:     "ABCDEFF",
				UARFCN:     0,
				PSC:        0,
				RAC:        0,
				RSCP:       0,
				ECIO:       0,
				PhyCh:      0,
				SF:         0,
				Slot:       0,
				SpeechCode: 0,
				ComMod:     0,
			},
		},
		{
			ATCommandResp: `+QENG: "servingcell","NOCONN","WCDMA",440,20,-,-,-,-,-,-,-,-,-,-,-,-`,
			WCDMACellInfo: WCDMACellInfo{
				Timestamp:  now,
				RAT:        "WCDMA",
				State:      "NOCONN",
				MCC:        440,
				MNC:        20,
				LAC:        "-",
				CellID:     "-",
				UARFCN:     0,
				PSC:        0,
				RAC:        0,
				RSCP:       0,
				ECIO:       0,
				PhyCh:      0,
				SF:         0,
				Slot:       0,
				SpeechCode: 0,
				ComMod:     0,
			},
		},
	}
	for _, expect := range getWCDMACellInfoTestCases {
		actual, err := getWCDMACellInfo(string(expect.ATCommandResp))
		if err != nil {
			t.Errorf("%v", err)
		}
		if actual != expect.WCDMACellInfo {
			t.Errorf("got unexpected result, expected: %v, but got %v", expect.WCDMACellInfo, actual)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	validConfigPath := "test/mihari.yml"
	expectedValidConfig := &Config{
		Name:        "eg25g",
		Path:        "/dev/ttyUSB3",
		Interval:    60,
		NewLineCode: "crlf",
		Parity:      "none",
		Stopbits:    1,
		Baurdrate:   115200,
		Databits:    8,
		ReadTimeout: 3,
		Forwarder:   "harvest",
	}
	actualConfig, err := loadConfig(validConfigPath)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedValidConfig, actualConfig) {
		t.Errorf("there is invalid config, expected %#v, got %#v", expectedValidConfig, actualConfig)
	}
}
