# [WIP]MIHARI
Mihari collects wireless quality of cellular from modem 

## Getting started

```
sudo apt install mihari
```
```
brew install mihari
```

```
mihari configure
```

```
modem device(default:/dev/modem):
baudrate(default:115200):
```

```
mihari
```

## Data format
```
{
    "time":160000,
    "imsi":"",
    "imei":"",
    "csq":10,
    "rssi":-50,
    "rsrp":50,
    "rsrq":50
}
```

## Output integration
- SORACOM Harvest
