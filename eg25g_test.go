package mihari

import "testing"

func TestGetQuecCellRAT(t *testing.T) {
	type getQuecCellRATTestCase struct {
		ATCommandResp string
		ExpectedRAT   string
		ExpectedErr   error
	}

	var getQuecCellRATTestCases = []getQuecCellRATTestCase{
		{
			ATCommandResp: `+QENG: "servingcell","SEARCH"`,
			ExpectedRAT:   "",
			ExpectedErr:   ErrModemNotAttached,
		},
		{
			ATCommandResp: `+QENG: "servingcell","CONNECT","WCDMA",440,10,75,6FE0090,10736,58,0,-78,-3,-,-,-,-,-`,
			ExpectedRAT:   RATTypeWCDMA,
			ExpectedErr:   nil,
		},
		{
			ATCommandResp: `+QENG: "servingcell","CONNECT","LTE","FDD",440,10,2734811,235,6100,19,3,3,1684,-81,-10,-54,19,50`,
			ExpectedRAT:   RATTypeLTE,
			ExpectedErr:   nil,
		},
	}

	for _, testCase := range getQuecCellRATTestCases {
		actualRAT, err := getQuecCellRAT(testCase.ATCommandResp)
		if err != testCase.ExpectedErr {
			t.Errorf("got unexpected result, expected: %v, but got %v", testCase.ExpectedErr, err)
		}
		if actualRAT != testCase.ExpectedRAT {
			t.Errorf("got unexpected result, expected: %v, but got %v", testCase.ExpectedRAT, actualRAT)
		}
	}

}
