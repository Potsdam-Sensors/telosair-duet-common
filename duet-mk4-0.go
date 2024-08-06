package telosairduetcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"maps"
	"math"
	"strconv"
)

func RoundFloatTwoDecimals(val float32) float32 {
	return float32(math.Round(float64(val*100)) / 100)
}

/* ~~ MK4 Var 0 ~~ */
var DuetTypeMk4Var0 = DuetTypeInfo{
	ExpectedBytes:        70,
	ExpectedStringLen:    15,
	StructInstanceGetter: func() DuetData { return &DuetDataMk4Var0{} },
	TypeAlias:            "Mk4.0",
}

type DuetDataMk4Var0 struct {
	SerialNumber   uint16
	SampleTimeMs   uint32
	UnixSec        uint32
	LastResetUnix  uint32
	SensorStates   uint8
	PoeUsbVoltage  uint8
	ConnectionType int
	PiMcuTemp      float32

	Pt1       Pms5003Measurement
	Pt2       Pms5003Measurement
	PtM       Pms5003Measurement
	Scd       Scd41Measurement
	Htu       Htu21Measurement
	TempRh    CombinedTempRhMeasurements
	Mprls     MprlsMeasurement
	Sgp       Sgp40Measurement
	RadioMeta RadioMetadata
}

func (d *DuetDataMk4Var0) SetPiMcuTemp(val float32) {
	d.PiMcuTemp = val
}
func (d *DuetDataMk4Var0) String() string {
	return fmt.Sprintf("[Duet %d, Type %d.%d | Unix %d | %s | HTU: %s | SCD: %s | MPRLS: %s | SGP: %s | PT1: %s | PT2: %s | PTM: %s | Radio: %s | Errstate %d | PoE Voltage %d]",
		d.SerialNumber, 4, 0, d.UnixSec, d.TempRh.String(), d.Htu.String(), d.Scd.String(), d.Mprls.String(), d.Sgp.String(), d.Pt1.String(), d.Pt2.String(), d.PtM.String(),
		d.RadioMeta.String(), d.SensorStates, d.PoeUsbVoltage)
}
func (d *DuetDataMk4Var0) GetTypeInfo() DuetTypeInfo {
	return DuetTypeMk4Var0
}

func (d *DuetDataMk4Var0) SetConnectionType(ct int) {
	d.ConnectionType = ct
}
func (d *DuetDataMk4Var0) SetTimeRadio(unixSecRecieved uint32) error {
	if (d.RadioMeta.RadioSentTimeMs < d.SampleTimeMs) || (unixSecRecieved*d.RadioMeta.RadioSentTimeMs*unixSecRecieved == 0) {
		return fmt.Errorf("incompatible timekeeping parameters: unix: %d, radio sent ms: %d, sample ms: %d", unixSecRecieved, d.RadioMeta.RadioSentTimeMs, d.SampleTimeMs)
	}
	d.UnixSec = unixSecRecieved - ((d.RadioMeta.RadioSentTimeMs - d.SampleTimeMs) / 1000)
	return nil
}

func (d *DuetDataMk4Var0) SetTimeSerial(unixSecRecieved uint32) {
	d.UnixSec = unixSecRecieved
}

func (d *DuetDataMk4Var0) RecalculateLastResetUnix() {
	d.LastResetUnix = d.UnixSec - (d.SampleTimeMs / 1000)
}

func (d *DuetDataMk4Var0) doPopulateFromSubStrings(splitStr []string) error {
	// Serial Number
	sn, err := strconv.ParseUint(splitStr[0], 10, 16)
	if err != nil {
		return fmt.Errorf("failed to convert DuetSerialNumber string, %s, to uint32", splitStr[0])
	}
	d.SerialNumber = uint16(sn)

	// Sample Time
	st, err := strconv.ParseUint(splitStr[1], 10, 32)
	if err != nil {
		return fmt.Errorf("failed to convert SampleTime string, %s, to uint32", splitStr[2])
	}
	d.SampleTimeMs = uint32(st)

	// Plantower PMS5003s
	if err := d.Pt1.FromSerialString(splitStr[2]); err != nil {
		return fmt.Errorf("failed to convert PT0 string, %s, to PlantowerData", splitStr[3])
	}

	if err := d.Pt2.FromSerialString(splitStr[3]); err != nil {
		return fmt.Errorf("failed to convert PT1 string, %s, to PlantowerData", splitStr[4])
	}

	// Temperatures (1 & 2)
	if temp, err := strconv.ParseFloat(splitStr[4], 32); err != nil {
		return fmt.Errorf("failed to convert htu temp string, %s, to float32", splitStr[5])
	} else {
		d.Htu.Temp = float32(temp)
	}

	if temp, err := strconv.ParseFloat(splitStr[5], 32); err != nil {
		return fmt.Errorf("failed to convert scd temp string, %s, to float32", splitStr[5])
	} else {
		d.Scd.Temp = float32(temp)
	}

	// Humidities (1 & 2)
	if hum, err := strconv.ParseFloat(splitStr[6], 32); err != nil {
		return fmt.Errorf("failed to convert htu hum string, %s, to float32", splitStr[5])
	} else {
		d.Htu.Hum = float32(hum)
	}

	if hum, err := strconv.ParseFloat(splitStr[7], 32); err != nil {
		return fmt.Errorf("failed to convert scd hum string, %s, to float32", splitStr[5])
	} else {
		d.Scd.Hum = float32(hum)
	}

	// Pressure
	if press, err := strconv.ParseFloat(splitStr[8], 32); err != nil {
		return fmt.Errorf("failed to convert pressure string, %s, to float32", splitStr[5])
	} else {
		d.Mprls.Pressure = float32(press)
	}

	// VOC Index
	if voc, err := strconv.ParseUint(splitStr[9], 10, 32); err != nil {
		return fmt.Errorf("failed to convert voc index string, %s, to uint32", splitStr[5])
	} else {
		d.Sgp.VocIndex = uint32(voc)
	}

	// CO2
	if co2, err := strconv.ParseUint(splitStr[10], 10, 16); err != nil {
		return fmt.Errorf("failed to convert co2 string, %s, to uint32", splitStr[5])
	} else {
		d.Scd.Co2 = uint16(co2)
	}

	// PoE / USB Voltage
	if voltage, err := strconv.ParseUint(splitStr[11], 10, 8); err != nil {
		return fmt.Errorf("failed to convert voltage string, %s, to uint8", splitStr[13])
	} else {
		d.PoeUsbVoltage = uint8(voltage)
	}

	// Sensor States
	if sensorStates, err := strconv.ParseUint(splitStr[12], 10, 8); err != nil {
		return fmt.Errorf("failed to convert states string, %s, to uint8", splitStr[14])
	} else {
		d.SensorStates = uint8(sensorStates)
	}
	MergePT(&d.Pt1, &d.Pt2, &d.PtM)
	CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)

	return nil
}
func (d *DuetDataMk4Var0) doPopulateFromBytes(buff []byte) error {
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
	if err := d.Pt1.PopulateFromBytes(buff[34:52]); err != nil {
		return fmt.Errorf("error parsing bytes for PT: %w", err)
	}
	if err := d.Pt2.PopulateFromBytes(buff[52:70]); err != nil {
		return fmt.Errorf("error parsing bytes for PT: %w", err)
	}
	MergePT(&d.Pt1, &d.Pt2, &d.PtM)
	CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)

	return nil
}
func (d *DuetDataMk4Var0) ToMap(gatewaySerial string) map[string]any {
	ret := map[string]any{
		KEY_DEVICE_TYPE:     4.0,
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
	maps.Copy(ret, d.Pt1.ToMap("_t"))
	maps.Copy(ret, d.Pt2.ToMap("_b"))
	maps.Copy(ret, d.PtM.ToMap("_m"))
	maps.Copy(ret, d.Htu.ToMap())
	maps.Copy(ret, d.Scd.ToMap())
	maps.Copy(ret, d.TempRh.ToMap())
	maps.Copy(ret, d.Mprls.ToMap())
	maps.Copy(ret, d.Sgp.ToMap())
	maps.Copy(ret, d.RadioMeta.ToMap())

	return ret
}
