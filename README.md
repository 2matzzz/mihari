# [WIP]MIHARI
Mihari collects wireless quality of cellular from modem 

## Data format

LTE
```
{
  "timestamp": 1639997713775,
  "rat": "LTE",
  "state": "NOCONN",
  "is_tdd": "FDD",
  "mcc": 440,
  "mnc": 10,
  "cellid": "2734000",
  "pcid": 400,
  "earfcn": 1850,
  "freq_band_ind": 3,
  "ul_bandwidth": 5,
  "dl_bandwidth": 5,
  "tac": 1684,
  "rsrp": -92,
  "rsrq": -8,
  "rssi": -64,
  "sinr": 15,
  "srxlev": 37
}
```

WCDMA
```
{
  "timestamp": 1639997713775,
  "rat": "WCDMA",
  "state": "NOCONN",
  "mcc": 440,
  "mnc": 10,
  "lac": "60",
  "cellid": "6FD8000",
  "uarfcn": 1037,
  "psc": 20,
  "rac": 0,
  "rscp": -64,
  "ecio": -3,
  "phych": 0,
  "sf": 0,
  "slot": 0,
  "speech_code": 0,
  "com_mod": 0
}
```

## Supported modem list
- Quectel EG25

## Output integration
- SORACOM Harvest

## TODO
- CI/CD
- Package manager
- More modem support