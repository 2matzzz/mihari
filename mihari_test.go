package mihari

import (
	"reflect"
	"testing"
)

func TestNewConfig(t *testing.T) {
	config, err := NewConfig("example/mihari.yml")
	if err != nil {
		t.Errorf("%v", err)
	}
	client, err := NewClient(config)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer client.Close()

	t.Logf("%#v", client)
}

func TestPort_GetIMEI(t *testing.T) {

}

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

type GetLTECellInfoTestCase struct {
	ATCommandResp string
	LTECellInfo
}

func TestGetLTECellInfo(t *testing.T) {
	var getLTECellInfoTestCases = []GetLTECellInfoTestCase{
		{
			ATCommandResp: `+QENG: "servingcell","NOCONN","LTE","FDD",440,10,2734811,235,6100,19,3,3,1684,-81,-10,-54,19,50`,
			LTECellInfo: LTECellInfo{
				State:          "NOCONN",
				IsTDD:          "FDD",
				MCC:            440,
				MNC:            10,
				CellID:         "2734811",
				PhysicalCellID: 235,
				EARFCN:         6100,
				Band:           19,
				ULBandwidth:    3,
				DLBandwidth:    3,
				Tac:            1684,
				RSRP:           -81,
				RSRQ:           -10,
				RSSI:           -54,
				SINR:           19,
				Srxlev:         50,
			},
		},
		{
			ATCommandResp: `+QENG: "servingcell","NOCONN","LTE","FDD",-,-,-,-,-,-,-,-,-,-,-,-,-,-`,
			LTECellInfo: LTECellInfo{
				State:          "NOCONN",
				IsTDD:          "FDD",
				MCC:            0,
				MNC:            0,
				CellID:         "",
				PhysicalCellID: 0,
				EARFCN:         0,
				Band:           0,
				ULBandwidth:    0,
				DLBandwidth:    0,
				Tac:            0,
				RSRP:           0,
				RSRQ:           0,
				RSSI:           0,
				SINR:           0,
				Srxlev:         0,
			},
		},
		{
			ATCommandResp: `+QENG: "servingcell","NOCONN","LTE","FDD",440,10,2C81851,193,1850,3,5,5,1694,-100,-11,-67,11,29`,
			LTECellInfo: LTECellInfo{
				State:          "NOCONN",
				IsTDD:          "FDD",
				MCC:            440,
				MNC:            10,
				CellID:         "2C81851",
				PhysicalCellID: 193,
				EARFCN:         1850,
				Band:           3,
				ULBandwidth:    5,
				DLBandwidth:    5,
				Tac:            1694,
				RSRP:           -100,
				RSRQ:           -11,
				RSSI:           -67,
				SINR:           11,
				Srxlev:         29,
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

func TestLoadConfig(t *testing.T) {
	validConfigPath := "example/mihari.yml"
	expectedValidConfig := &Config{
		Name:        "eg25g",
		Path:        "/dev/ttyUSB3",
		Interval:    10,
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
