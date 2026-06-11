package telosairduetcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"maps"
	"strconv"
)

/* ~~ MK4 Var 24 - One OPC-N3 ~~ */
var DuetTypeMk4Var24 = DuetTypeInfo{
	ExpectedBytes:        152,
	ExpectedStringLen:    19,
	StructInstanceGetter: func() DuetData { return &DuetDataMk4Var24{} },
	TypeAlias:            "Mk4.24",
}

type DuetDataMk4Var24 struct {
	SerialNumber   uint16
	SampleTimeMs   uint32
	UnixSec        uint32
	LastResetUnix  uint32
	SensorStates   uint8
	PoeUsbVoltage  uint8
	ConnectionType int
	PiMcuTemp      float32
	piMcuTempSet   bool

	Opc       AlphasenseOpcN3Measurement
	Scd       Scd41Measurement
	Htu       Htu21Measurement
	TempRh    CombinedTempRhMeasurements
	Mprls     MprlsMeasurement
	Sgp       Sgp40Measurement
	RadioMeta RadioMetadata

	timeResolved bool
}

func (d *DuetDataMk4Var24) TimeResolved() bool {
	return d.timeResolved
}
func (d *DuetDataMk4Var24) MarkTimeResolved(v bool) {
	d.timeResolved = v
}
func (d *DuetDataMk4Var24) Timestamp() uint32 {
	return d.UnixSec
}
func (d *DuetDataMk4Var24) ResolveTime(t uint32) {
	d.UnixSec = t
}

func (d *DuetDataMk4Var24) SensorMeasurements() []SensorMeasurement {
	return []SensorMeasurement{d.Opc, d.TempRh, d.Scd, d.Mprls, d.Sgp, DuetSensorState{d.SensorStates}}
}
func (d *DuetDataMk4Var24) SetRadioData(v RadioMetadata) {
	d.RadioMeta = v
}
func (d *DuetDataMk4Var24) SetPiMcuTemp(val float32) {
	d.PiMcuTemp = val
	d.piMcuTempSet = true
}
func (d *DuetDataMk4Var24) String() string {
	return fmt.Sprintf("[Duet %d, Type 4.24 | Unix %d | %s | HTU: %s | SCD: %s | MPRLS: %s | SGP: %s | OPC: %s | Radio: %s | Errstate %d | PoE Voltage %d]",
		d.SerialNumber, d.UnixSec, d.TempRh.String(), d.Htu.String(), d.Scd.String(), d.Mprls.String(), d.Sgp.String(), d.Opc.String(),
		d.RadioMeta.String(), d.SensorStates, d.PoeUsbVoltage)
}
func (d *DuetDataMk4Var24) GetTypeInfo() DuetTypeInfo {
	return DuetTypeMk4Var24
}

func (d *DuetDataMk4Var24) SetConnectionType(ct int) {
	d.ConnectionType = ct
}
func (d *DuetDataMk4Var24) SetTimeRadio(unixSecRecieved uint32) error {
	if (d.RadioMeta.RadioSentTimeMs < d.SampleTimeMs) || (unixSecRecieved*d.RadioMeta.RadioSentTimeMs*unixSecRecieved == 0) {
		return fmt.Errorf("incompatible timekeeping parameters: unix: %d, radio sent ms: %d, sample ms: %d", unixSecRecieved, d.RadioMeta.RadioSentTimeMs, d.SampleTimeMs)
	}
	d.UnixSec = unixSecRecieved - ((d.RadioMeta.RadioSentTimeMs - d.SampleTimeMs) / 1000)
	return nil
}

func (d *DuetDataMk4Var24) SetTimeSerial(unixSecRecieved uint32) {
	d.UnixSec = unixSecRecieved
}

func (d *DuetDataMk4Var24) RecalculateLastResetUnix() {
	d.LastResetUnix = d.UnixSec - (d.SampleTimeMs / 1000)
}

/*
Error converting data point, 4 24 4404 1831248 27.06 24.52 30.38 28.62 33.65 25.89 404.82 100 426 0.27 0.51 0.59 39 0
: failed to populate for type Mk4.24: failed to convert voc index string, 25.89, to uint32
*/
func (d *DuetDataMk4Var24) doPopulateFromSubStrings(splitStr []string) error {
	cur := 0
	// Serial Number
	sn, err := strconv.ParseUint(splitStr[cur], 10, 16)
	if err != nil {
		return fmt.Errorf("failed to convert DuetSerialNumber string, %s, to uint32", splitStr[cur])
	}
	d.SerialNumber = uint16(sn)
	cur++

	// Sample Time
	st, err := strconv.ParseUint(splitStr[cur], 10, 32)
	if err != nil {
		return fmt.Errorf("failed to convert SampleTime string, %s, to uint32", splitStr[cur])
	}
	d.SampleTimeMs = uint32(st)
	cur++

	// Temperatures (1 & 2 & 3)
	if temp, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert htu temp string, %s, to float32", splitStr[cur])
	} else {
		d.Htu.Temp = float32(temp)
	}
	cur++

	if temp, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert scd temp string, %s, to float32", splitStr[cur])
	} else {
		d.Scd.Temp = float32(temp)
	}
	cur++

	if temp, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert opc temp string, %s, to float32", splitStr[cur])
	} else {
		d.Opc.Temp = float32(temp)
	}
	cur++

	// Humidities (1 & 2 & 3)
	if hum, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert htu hum string, %s, to float32", splitStr[cur])
	} else {
		d.Htu.Hum = float32(hum)
	}
	cur++

	if hum, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert scd hum string, %s, to float32", splitStr[cur])
	} else {
		d.Scd.Hum = float32(hum)
	}
	cur++

	if hum, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert opc hum string, %s, to float32", splitStr[cur])
	} else {
		d.Opc.Rh = float32(hum)
	}
	cur++

	// Pressure
	if press, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert pressure string, %s, to float32", splitStr[cur])
	} else {
		d.Mprls.Pressure = float32(press)
	}
	cur++

	// VOC Index
	if voc, err := strconv.ParseUint(splitStr[cur], 10, 32); err != nil {
		return fmt.Errorf("failed to convert voc index string, %s, to uint32", splitStr[cur])
	} else {
		d.Sgp.VocIndex = uint32(voc)
	}
	cur++

	// CO2
	if co2, err := strconv.ParseUint(splitStr[cur], 10, 16); err != nil {
		return fmt.Errorf("failed to convert co2 string, %s, to uint32", splitStr[cur])
	} else {
		d.Scd.Co2 = uint16(co2)
	}
	cur++

	// OPC-N3 PM1, 2.5, 10 & Bins
	if pm1, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert opc pm1 string, %s, to float32", splitStr[cur])
	} else {
		d.Opc.PM1 = float32(pm1)
	}
	cur++

	if pm2p5, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert opc pm2.5 string, %s, to float32", splitStr[cur])
	} else {
		d.Opc.PM2p5 = float32(pm2p5)
	}
	cur++

	if pm10, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert opc pm10 string, %s, to float32", splitStr[cur])
	} else {
		d.Opc.PM10 = float32(pm10)
	}
	cur++

	if err := d.Opc.PopulateBinsFromString(splitStr[cur]); err != nil {
		return fmt.Errorf("failed to convert opcn3 bins: %v", err)
	}
	cur++

	// PoE / USB Voltage
	if voltage, err := strconv.ParseUint(splitStr[cur], 10, 8); err != nil {
		return fmt.Errorf("failed to convert voltage string, %s, to uint8", splitStr[cur])
	} else {
		d.PoeUsbVoltage = uint8(voltage)
	}
	cur++

	// Sensor States
	if sensorStates, err := strconv.ParseUint(splitStr[cur], 10, 8); err != nil {
		return fmt.Errorf("failed to convert states string, %s, to uint8", splitStr[cur])
	} else {
		d.SensorStates = uint8(sensorStates)
	}
	CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)
	cur++

	return nil
}
func (d *DuetDataMk4Var24) doPopulateFromBytes(buff []byte) error {
	d.SensorStates = buff[0]
	d.PoeUsbVoltage = buff[1]
	d.SerialNumber = binary.LittleEndian.Uint16(buff[2:4])
	d.Scd.Co2 = binary.LittleEndian.Uint16(buff[4:6])
	d.Sgp.VocIndex = binary.LittleEndian.Uint32(buff[6:10])
	d.SampleTimeMs = binary.LittleEndian.Uint32(buff[10:14])

	reader := bytes.NewReader(buff[14:34])
	if err := binary.Read(reader, binary.LittleEndian, &d.Htu.Temp); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Scd.Temp); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Htu.Hum); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Scd.Hum); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Mprls.Pressure); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Opc.Temp); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Opc.Rh); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Opc.PM1); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Opc.PM2p5); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Opc.PM10); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}

	CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)

	return nil
}
func (d *DuetDataMk4Var24) ToMap(gatewaySerial string) map[string]any {
	ret := map[string]any{
		KEY_DEVICE_TYPE:     4.24,
		KEY_SERIAL_NUMBER:   d.SerialNumber,
		KEY_DEVICE_ID:       d.SerialNumber,
		KEY_UNIX:            d.UnixSec,
		KEY_ECO2:            0,
		KEY_RAWH2:           0,
		KEY_SENSOR_STATES:   d.SensorStates,
		KEY_CONNECTION_TYPE: d.ConnectionType,
		KEY_LAST_RESET_TIME: d.LastResetUnix,
		KEY_GATEWAY_SERIAL:  gatewaySerial,
		KEY_POE_USB_VOLTAGE: d.PoeUsbVoltage,
	}
	maps.Copy(ret, d.Opc.ToMapPm("_t"))
	maps.Copy(ret, d.Opc.ToMapPm("_b"))
	maps.Copy(ret, d.Opc.ToMapPm("_m"))
	maps.Copy(ret, d.Opc.ToMapBins())
	maps.Copy(ret, d.Opc.ToMapTempRh())
	maps.Copy(ret, d.Htu.ToMap())
	maps.Copy(ret, d.Scd.ToMap())
	maps.Copy(ret, d.TempRh.ToMap())
	maps.Copy(ret, d.Mprls.ToMap())
	maps.Copy(ret, d.Sgp.ToMap())
	maps.Copy(ret, d.RadioMeta.ToMap())
	if d.piMcuTempSet {
		ret[KEY_PI_MCU_TEMP] = d.PiMcuTemp
	}

	return ret
}
