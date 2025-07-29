package telosairduetcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"maps"
	"strconv"
)

/* ~~ MK4 Var13 ~~ */
var DuetTypeMk4Var13 = DuetTypeInfo{
	ExpectedBytes:        82,
	ExpectedStringLen:    21,
	StructInstanceGetter: func() DuetData { return &DuetDataMk4Var13{} },
	TypeAlias:            "Mk4.13",
}

type DuetDataMk4Var13 struct {
	SerialNumber   uint16
	SampleTimeMs   uint32
	UnixSec        uint32
	LastResetUnix  uint32
	SensorStates   uint8
	PoeUsbVoltage  uint8
	ConnectionType int
	PiMcuTemp      float32
	piMcuTempSet   bool

	Sps       Sps30Measurement
	Scd       Scd41Measurement
	Htu       Htu21Measurement
	TempRh    CombinedTempRhMeasurements
	Mprls     MprlsMeasurement
	Sgp       Sgp40Measurement
	RadioMeta RadioMetadata

	Co, No, No2, Ch2o, H2s float32
	Tgs2611, Tgs2600       float32 // TODO: What are these called and also make a struct for it

	timeResolved bool
}

func (d *DuetDataMk4Var13) TimeResolved() bool {
	return d.timeResolved
}
func (d *DuetDataMk4Var13) MarkTimeResolved(v bool) {
	d.timeResolved = v
}
func (d *DuetDataMk4Var13) Timestamp() uint32 {
	return d.UnixSec
}
func (d *DuetDataMk4Var13) ResolveTime(t uint32) {
	d.UnixSec = t
}

func (d *DuetDataMk4Var13) SensorMeasurements() []SensorMeasurement {
	return []SensorMeasurement{d.Sps, d.TempRh, d.Scd, d.Mprls, d.Sgp, DuetSensorState{d.SensorStates}}
} // TODO: See Tgs and gas

func (d *DuetDataMk4Var13) SetRadioData(v RadioMetadata) {
	d.RadioMeta = v
}
func (d *DuetDataMk4Var13) SetPiMcuTemp(val float32) {
	d.PiMcuTemp = val
	d.piMcuTempSet = true
}
func (d *DuetDataMk4Var13) String() string {
	return fmt.Sprintf("[Duet %d, Type 4.13 | Unix %d | Co %.3f, No: %.3f, No2: %.3f, Ch2o: %.3f, H2s: %.3f | TGS 2611: %.3f, 2600: %.3f | %s | HTU: %s | SCD: %s | MPRLS: %s | SGP: %s | SPS30: %s | Radio: %s | Errstate %d | PoE Voltage %d]",
		d.SerialNumber, d.UnixSec, d.Co, d.No, d.No2, d.Ch2o, d.H2s, d.Tgs2611, d.Tgs2600, d.TempRh.String(), d.Htu.String(), d.Scd.String(), d.Mprls.String(), d.Sgp.String(), d.Sps.String(),
		d.RadioMeta.String(), d.SensorStates, d.PoeUsbVoltage)
}
func (d *DuetDataMk4Var13) GetTypeInfo() DuetTypeInfo {
	return DuetTypeMk4Var13
}

func (d *DuetDataMk4Var13) SetConnectionType(ct int) {
	d.ConnectionType = ct
}
func (d *DuetDataMk4Var13) SetTimeRadio(unixSecRecieved uint32) error {
	if (d.RadioMeta.RadioSentTimeMs < d.SampleTimeMs) || (unixSecRecieved*d.RadioMeta.RadioSentTimeMs*unixSecRecieved == 0) {
		return fmt.Errorf("incompatible timekeeping parameters: unix: %d, radio sent ms: %d, sample ms: %d", unixSecRecieved, d.RadioMeta.RadioSentTimeMs, d.SampleTimeMs)
	}
	d.UnixSec = unixSecRecieved - ((d.RadioMeta.RadioSentTimeMs - d.SampleTimeMs) / 1000)
	return nil
}

func (d *DuetDataMk4Var13) SetTimeSerial(unixSecRecieved uint32) {
	d.UnixSec = unixSecRecieved
}

func (d *DuetDataMk4Var13) RecalculateLastResetUnix() {
	d.LastResetUnix = d.UnixSec - (d.SampleTimeMs / 1000)
}

// TODO: Is there a clean way to check that idx ends up lte to the expected num
// Honestly the more DRY version I want to cook up probably solves this too
func (d *DuetDataMk4Var13) doPopulateFromSubStrings(splitStr []string) error {
	idx := 0

	// Serial Number
	sn, err := strconv.ParseUint(splitStr[idx], 10, 16)
	if err != nil {
		return fmt.Errorf("failed to convert DuetSerialNumber string, %s, to uint32", splitStr[idx])
	}
	d.SerialNumber = uint16(sn)
	idx += 1

	// Sample Time
	st, err := strconv.ParseUint(splitStr[idx], 10, 32)
	if err != nil {
		return fmt.Errorf("failed to convert SampleTime string, %s, to uint32", splitStr[idx])
	}
	d.SampleTimeMs = uint32(st)
	idx += 1

	// SPS30s (as PMS5003)
	if err := d.Sps.FromSerialString(splitStr[idx]); err != nil {
		return fmt.Errorf("failed to convert sps30 string, %s, to PlantowerData", splitStr[idx])
	}
	idx += 1

	// Temperatures (1 & 2)
	if temp, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert htu temp string, %s, to float32", splitStr[idx])
	} else {
		d.Htu.Temp = float32(temp)
	}
	idx += 1

	if temp, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert scd temp string, %s, to float32", splitStr[idx])
	} else {
		d.Scd.Temp = float32(temp)
	}
	idx += 1

	// Humidities (1 & 2)
	if hum, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert htu hum string, %s, to float32", splitStr[idx])
	} else {
		d.Htu.Hum = float32(hum)
	}
	idx += 1

	if hum, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert scd hum string, %s, to float32", splitStr[idx])
	} else {
		d.Scd.Hum = float32(hum)
	}
	idx += 1

	// Pressure
	if press, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert pressure string, %s, to float32", splitStr[idx])
	} else {
		d.Mprls.Pressure = float32(press)
	}
	idx += 1

	// VOC Index
	if voc, err := strconv.ParseUint(splitStr[idx], 10, 32); err != nil {
		return fmt.Errorf("failed to convert voc index string, %s, to uint32", splitStr[idx])
	} else {
		d.Sgp.VocIndex = uint32(voc)
	}
	idx += 1

	// CO2
	if co2, err := strconv.ParseUint(splitStr[idx], 10, 16); err != nil {
		return fmt.Errorf("failed to convert co2 string, %s, to uint32", splitStr[idx])
	} else {
		d.Scd.Co2 = uint16(co2)
	}
	idx += 1

	// PoE / USB Voltage
	if voltage, err := strconv.ParseUint(splitStr[idx], 10, 8); err != nil {
		return fmt.Errorf("failed to convert voltage string, %s, to uint8", splitStr[idx])
	} else {
		d.PoeUsbVoltage = uint8(voltage)
	}
	idx += 1

	// CO
	if co, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert CO string, %s, to float32", splitStr[idx])
	} else {
		d.Co = float32(co)
	}
	idx += 1

	// NO
	if no, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert NO string, %s, to float32", splitStr[idx])
	} else {
		d.No = float32(no)
	}
	idx += 1

	// NO2
	if no2, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert NO2 string, %s, to float32", splitStr[idx])
	} else {
		d.No2 = float32(no2)
	}
	idx += 1

	// CH2O
	if ch20, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert CH2O string, %s, to float32", splitStr[idx])
	} else {
		d.Ch2o = float32(ch20)
	}
	idx += 1

	// H2S
	if h2s, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert H2S string, %s, to float32", splitStr[idx])
	} else {
		d.H2s = float32(h2s)
	}
	idx += 1

	// TGS2611
	if tgs, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert TGS2611 string, %s, to float32", splitStr[idx])
	} else {
		d.Tgs2611 = float32(tgs)
	}
	idx += 1

	// TGS2600
	if tgs, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert TGS2600 string, %s, to float32", splitStr[idx])
	} else {
		d.Tgs2600 = float32(tgs)
	}
	idx += 1

	// Sensor States
	if sensorStates, err := strconv.ParseUint(splitStr[idx], 10, 8); err != nil {
		return fmt.Errorf("failed to convert states string, %s, to uint8", splitStr[idx])
	} else {
		d.SensorStates = uint8(sensorStates)
	}
	idx += 1

	CombineTempRhMeasurements(d.Htu, d.Scd, &(d.TempRh))
	return nil
}

// TODO: See TODO from populate substring func (wait why aren't we using an iterator with binary.read with a reader, that makes way more sense)
func (d *DuetDataMk4Var13) doPopulateFromBytes(buff []byte) error {
	idx := 0

	d.SensorStates = buff[idx]
	idx += 1

	d.PoeUsbVoltage = buff[idx]
	idx += 1

	d.SerialNumber = binary.LittleEndian.Uint16(buff[idx : idx+2])
	idx += 2

	d.Scd.Co2 = binary.LittleEndian.Uint16(buff[idx : idx+2])
	idx += 2

	d.Sgp.VocIndex = binary.LittleEndian.Uint32(buff[idx : idx+4])
	idx += 4

	d.SampleTimeMs = binary.LittleEndian.Uint32(buff[idx : idx+4])
	idx += 4

	reader := bytes.NewReader(buff[idx : idx+(12*4)])
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
	if err := binary.Read(reader, binary.LittleEndian, &d.Co); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.No); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.No2); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Ch2o); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.H2s); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Tgs2611); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Tgs2600); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	idx += (12 * 4)

	if err := d.Sps.PopulateFromBytes(buff[idx : idx+18]); err != nil {
		return fmt.Errorf("error parsing bytes for sps1: %w", err)
	}
	idx += 18

	CombineTempRhMeasurements(d.Htu, d.Scd, &(d.TempRh))
	return nil
}
func (d *DuetDataMk4Var13) ToMap(gatewaySerial string) map[string]any {
	ret := map[string]any{
		KEY_DEVICE_TYPE:     4.13,
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
		KEY_GAS_CO:          d.Co,
		KEY_GAS_NO:          d.No,
		KEY_GAS_NO2:         d.No2,
		KEY_GAS_CH2O:        d.Ch2o,
		KEY_GAS_H2S:         d.H2s,
		KEY_TGS2611_RS:      d.Tgs2611,
		KEY_TGS2600_RS:      d.Tgs2600,
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
