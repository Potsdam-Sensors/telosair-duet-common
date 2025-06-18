package telosairduetcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"maps"
	"strconv"
)

/* ~~ MK3 Var 1 ~~ */
var DuetTypeMk3Var1 = DuetTypeInfo{
	ExpectedBytes:        54,
	ExpectedStringLen:    13,
	StructInstanceGetter: func() DuetData { return &DuetDataMk3Var1{} },
	TypeAlias:            "Mk3.1",
}

type DuetDataMk3Var1 struct {
	SerialNumber   uint16
	SampleTimeMs   uint32
	UnixSec        uint32
	LastResetUnix  uint32
	SensorStates   uint8
	ConnectionType int
	PiMcuTemp      float32
	piMcuTempSet   bool

	Sps Pms5003Measurement

	Htu       Htu21Measurement
	Scd       Scd41Measurement
	TempRh    CombinedTempRhMeasurements
	Mprls     MprlsMeasurement
	Sgp       Sgp40Measurement
	RadioMeta RadioMetadata
}

func (d *DuetDataMk3Var1) SensorMeasurements() []SensorMeasurement {
	return []SensorMeasurement{d.Sps, d.Htu, d.Scd, d.TempRh, d.Mprls, d.Sgp, DuetSensorState{d.SensorStates}}
}

func (d *DuetDataMk3Var1) SetRadioData(v RadioMetadata) {
	d.RadioMeta = v
}
func (d *DuetDataMk3Var1) SetPiMcuTemp(val float32) {
	d.PiMcuTemp = val
	d.piMcuTempSet = true
}
func (d *DuetDataMk3Var1) String() string {
	return fmt.Sprintf("[Duet %d, Type %d.%d | Unix %d | HTU %s | SCD: %s | TempRH: %s | MPRLS: %s | SGP: %s | SPS30: %s | Radio: %s | Errstate %d ]",
		d.SerialNumber, 3, 1, d.UnixSec, d.Htu.String(), d.Scd.String(), d.TempRh, d.Mprls.String(), d.Sgp.String(), d.Sps.String(),
		d.RadioMeta.String(), d.SensorStates)
}
func (d *DuetDataMk3Var1) GetTypeInfo() DuetTypeInfo {
	return DuetTypeMk3Var1
}

func (d *DuetDataMk3Var1) SetConnectionType(ct int) {
	d.ConnectionType = ct
}
func (d *DuetDataMk3Var1) SetTimeRadio(unixSecRecieved uint32) error {
	if (d.RadioMeta.RadioSentTimeMs < d.SampleTimeMs) || (unixSecRecieved*d.RadioMeta.RadioSentTimeMs*unixSecRecieved == 0) {
		return fmt.Errorf("incompatible timekeeping parameters: unix: %d, radio sent ms: %d, sample ms: %d", unixSecRecieved, d.RadioMeta.RadioSentTimeMs, d.SampleTimeMs)
	}
	d.UnixSec = unixSecRecieved - ((d.RadioMeta.RadioSentTimeMs - d.SampleTimeMs) / 1000)
	return nil
}

func (d *DuetDataMk3Var1) SetTimeSerial(unixSecRecieved uint32) {
	d.UnixSec = unixSecRecieved
}

func (d *DuetDataMk3Var1) RecalculateLastResetUnix() {
	d.LastResetUnix = d.UnixSec - (d.SampleTimeMs / 1000)
}

func (d *DuetDataMk3Var1) doPopulateFromSubStrings(splitStr []string) error {
	// Serial Number
	sn, err := strconv.ParseUint(splitStr[0], 10, 16)
	if err != nil {
		return fmt.Errorf("failed to convert DuetSerialNumber string, %s, to uint32", splitStr[0])
	}
	d.SerialNumber = uint16(sn)

	// Sample Time
	st, err := strconv.ParseUint(splitStr[1], 10, 32)
	if err != nil {
		return fmt.Errorf("failed to convert SampleTime string, %s, to uint32", splitStr[1])
	}
	d.SampleTimeMs = uint32(st)

	// Plantower PMS5003s
	if err := d.Sps.FromSerialString(splitStr[2]); err != nil {
		return fmt.Errorf("failed to convert PT0 string, %s, to PlantowerData", splitStr[2])
	}

	// Temperature HTU
	if temp, err := strconv.ParseFloat(splitStr[3], 32); err != nil {
		return fmt.Errorf("failed to convert htu temp string, %s, to float32", splitStr[3])
	} else {
		d.Htu.Temp = float32(temp)
	}

	// Temperature SCD
	if temp, err := strconv.ParseFloat(splitStr[4], 32); err != nil {
		return fmt.Errorf("failed to convert scd temp string, %s, to float32", splitStr[4])
	} else {
		d.Scd.Temp = float32(temp)
	}

	// Humidity HTU
	if hum, err := strconv.ParseFloat(splitStr[5], 32); err != nil {
		return fmt.Errorf("failed to convert htu hum string, %s, to float32", splitStr[5])
	} else {
		d.Htu.Hum = float32(hum)
	}

	// Humidity SCD
	if hum, err := strconv.ParseFloat(splitStr[6], 32); err != nil {
		return fmt.Errorf("failed to convert scd hum string, %s, to float32", splitStr[6])
	} else {
		d.Scd.Hum = float32(hum)
	}

	// Pressure
	if press, err := strconv.ParseFloat(splitStr[7], 32); err != nil {
		return fmt.Errorf("failed to convert pressure string, %s, to float32", splitStr[7])
	} else {
		d.Mprls.Pressure = float32(press)
	}

	// TVOC
	if voc, err := strconv.ParseUint(splitStr[8], 10, 32); err != nil {
		return fmt.Errorf("failed to convert tvoc string, %s, to uint32", splitStr[8])
	} else {
		d.Sgp.VocIndex = uint32(voc)
	}

	// CO2
	if co2, err := strconv.ParseUint(splitStr[9], 10, 16); err != nil {
		return fmt.Errorf("failed to convert co2 string, %s, to uint32", splitStr[9])
	} else {
		d.Scd.Co2 = uint16(co2)
	}

	// Sensor States
	if sensorStates, err := strconv.ParseUint(splitStr[10], 10, 8); err != nil {
		return fmt.Errorf("failed to convert states string, %s, to uint8", splitStr[10])
	} else {
		d.SensorStates = uint8(sensorStates)
	}
	CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)

	return nil
}

func (d *DuetDataMk3Var1) doPopulateFromBytes(buff []byte) error {
	d.SensorStates = buff[0]
	// Skip buff[1] as it is padding
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

	if err := d.Sps.PopulateFromBytes(buff[34:52]); err != nil {
		return fmt.Errorf("error parsing bytes for PT: %w", err)
	}
	CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)
	return nil
}
func (d *DuetDataMk3Var1) ToMap(gatewaySerial string) map[string]any {
	ret := map[string]any{
		KEY_DEVICE_TYPE:     3.1,
		KEY_SERIAL_NUMBER:   d.SerialNumber,
		KEY_DEVICE_ID:       d.SerialNumber,
		KEY_UNIX:            d.UnixSec,
		KEY_ECO2:            0,
		KEY_RAWH2:           0,
		KEY_SENSOR_STATES:   d.SensorStates,
		KEY_CONNECTION_TYPE: d.ConnectionType,
		KEY_LAST_RESET_TIME: d.LastResetUnix,
		KEY_GATEWAY_SERIAL:  gatewaySerial,
	}
	maps.Copy(ret, d.Sps.ToMap("_t"))
	maps.Copy(ret, d.Sps.ToMap("_b"))
	maps.Copy(ret, d.Sps.ToMap("_m"))
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
