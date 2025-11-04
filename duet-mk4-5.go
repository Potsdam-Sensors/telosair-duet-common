package telosairduetcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"maps"
	"strconv"
)

/* ~~ MK4 Var 5 (Outdoor, 2 SPS30s, CO, O3, NO2) ~~ */
var DuetTypeMk4Var5 = DuetTypeInfo{
	ExpectedBytes:        70, // Haven't verified
	ExpectedStringLen:    17,
	StructInstanceGetter: func() DuetData { return &DuetDataMk4Var5{} },
	TypeAlias:            "Mk4.5",
}

type DuetDataMk4Var5 struct {
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

	Co, O3, No2 float32

	timeResolved bool
}

func (d *DuetDataMk4Var5) TimeResolved() bool {
	return d.timeResolved
}
func (d *DuetDataMk4Var5) MarkTimeResolved(v bool) {
	d.timeResolved = v
}
func (d *DuetDataMk4Var5) Timestamp() uint32 {
	return d.UnixSec
}
func (d *DuetDataMk4Var5) ResolveTime(t uint32) {
	d.UnixSec = t
}

func (d *DuetDataMk4Var5) SensorMeasurements() []SensorMeasurement {
	return []SensorMeasurement{d.Sps, d.TempRh, d.Scd, d.Mprls, d.Sgp, DuetSensorState{d.SensorStates}} // TODO: Gas?
}
func (d *DuetDataMk4Var5) SetRadioData(v RadioMetadata) {
	d.RadioMeta = v
}
func (d *DuetDataMk4Var5) SetPiMcuTemp(val float32) {
	d.PiMcuTemp = val
	d.piMcuTempSet = true
}
func (d *DuetDataMk4Var5) String() string {
	return fmt.Sprintf("[Duet %d, Type 4.5 | Unix %d | Co %.2f, O3: %.2f, CH4: %.2f | %s | HTU: %s | SCD: %s | MPRLS: %s | SGP: %s | SPS30 [%s] | Radio: %s | Errstate %d | PoE Voltage %d]",
		d.SerialNumber, d.UnixSec, d.Co, d.O3, d.No2, d.TempRh.String(), d.Htu.String(), d.Scd.String(), d.Mprls.String(), d.Sgp.String(), d.Sps.String(),
		d.RadioMeta.String(), d.SensorStates, d.PoeUsbVoltage)
}
func (d *DuetDataMk4Var5) GetTypeInfo() DuetTypeInfo {
	return DuetTypeMk4Var5
}

func (d *DuetDataMk4Var5) SetConnectionType(ct int) {
	d.ConnectionType = ct
}
func (d *DuetDataMk4Var5) SetTimeRadio(unixSecRecieved uint32) error {
	if (d.RadioMeta.RadioSentTimeMs < d.SampleTimeMs) || (unixSecRecieved*d.RadioMeta.RadioSentTimeMs*unixSecRecieved == 0) {
		return fmt.Errorf("incompatible timekeeping parameters: unix: %d, radio sent ms: %d, sample ms: %d", unixSecRecieved, d.RadioMeta.RadioSentTimeMs, d.SampleTimeMs)
	}
	d.UnixSec = unixSecRecieved - ((d.RadioMeta.RadioSentTimeMs - d.SampleTimeMs) / 1000)
	return nil
}

func (d *DuetDataMk4Var5) SetTimeSerial(unixSecRecieved uint32) {
	d.UnixSec = unixSecRecieved
}

func (d *DuetDataMk4Var5) RecalculateLastResetUnix() {
	d.LastResetUnix = d.UnixSec - (d.SampleTimeMs / 1000)
}

func (d *DuetDataMk4Var5) doPopulateFromSubStrings(splitStr []string) error {
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

	// SPS
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

	// Gases
	if co, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert co string, %s, to float32", splitStr[idx])
	} else {
		d.Co = float32(co)
	}
	idx += 1

	if o3, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert o3 string, %s, to float32", splitStr[idx])
	} else {
		d.O3 = float32(o3)
	}
	idx += 1

	if no2, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert no2 string, %s, to float32", splitStr[idx])
	} else {
		d.No2 = float32(no2)
	}
	idx += 1

	// PoE / USB Voltage
	if voltage, err := strconv.ParseUint(splitStr[idx], 10, 8); err != nil {
		return fmt.Errorf("failed to convert voltage string, %s, to uint8", splitStr[idx])
	} else {
		d.PoeUsbVoltage = uint8(voltage)
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

func (d *DuetDataMk4Var5) doPopulateFromBytes(buff []byte) error {
	reader := bytes.NewReader(buff)
	pointers := append(
		// 4 + (11*4) + 18 = 70
		[]any{&d.SensorStates, &d.PoeUsbVoltage, &d.SerialNumber, &d.Scd.Co2, &d.Sgp.VocIndex, &d.SampleTimeMs,
			&d.Htu.Temp, &d.Scd.Temp, &d.Htu.Hum, &d.Scd.Hum, &d.Mprls.Pressure, &d.Co, &d.O3, &d.No2},
		d.Sps.PointerIterable()...,
	)

	for idx := range pointers {
		if err := binary.Read(reader, binary.LittleEndian, pointers[idx]); err != nil {
			return fmt.Errorf("error converting bytes at index %d: %w", idx, err)
		}
	}
	CombineTempRhMeasurements(d.Scd, d.Htu, &d.TempRh)
	return nil
}
func (d *DuetDataMk4Var5) ToMap(gatewaySerial string) map[string]any {
	ret := map[string]any{
		KEY_DEVICE_TYPE:     4.5,
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
		KEY_GAS_O3:          d.O3,
		KEY_GAS_NO2:         d.No2,
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
