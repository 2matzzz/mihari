package mihari

import (
	"testing"
)

func TestPort_Check(t *testing.T) {
	port := &Port{
		Path: "/dev/ttyUSB3",
	}

	if err := port.Check(); err != nil {
		t.Errorf("%s", err)
	}
	t.Logf("%v", port)
}

func TestPort_GetIMEI(t *testing.T) {

}

func TestGetQuecCellRATMode(t *testing.T) {
	validCommandResp := []byte(`+QENG: "servingcell","NOCONN","LTE","FDD",440,10,2734811,235,6100,19,3,3,1684,-81,-10,-54,19,50`)
	state, rat, err := getQuecCellRATMode(string(validCommandResp))
	if err != nil {
		t.Errorf("%v", err)
	}
	if state == "" || rat == "" {
		t.Errorf("queccell command response parse error, expected \"NOCONN\" and \"LTE\" but actual \"%s\" and \"%s\"", state, rat)
	}

	invalidStateCommandResp := []byte(`+QENG: "servingcell","INVALID","INVALID","FDD",440,10,2734811,235,6100,19,3,3,1684,-81,-10,-54,19,50`)
	state, rat, err = getQuecCellRATMode(string(invalidStateCommandResp))
	if err == nil || state != "" || rat != "" {
		t.Errorf("queccell command response parse failed, this test case should be fail but succeed")
	}
}
