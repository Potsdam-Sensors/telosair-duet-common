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
	ExpectedBytes:        108, // Haven't verified
	ExpectedStringLen:    16,
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

	Sps1      Sps30Measurement
	Sps2      Sps30Measurement
	SpsM      Sps30Measurement
	Scd       Scd41Measurement
	Htu       Htu21Measurement
	TempRh    CombinedTempRhMeasurements
	Mprls     MprlsMeasurement
	Sgp       Sgp40Measurement
	RadioMeta RadioMetadata

	Gas         GasSensorsMeasurement
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
	return []SensorMeasurement{d.SpsM, d.TempRh, d.Scd, d.Mprls, d.Sgp, &d.Gas, DuetSensorState{d.SensorStates}}
}
func (d *DuetDataMk4Var5) SetRadioData(v RadioMetadata) {
	d.RadioMeta = v
}
func (d *DuetDataMk4Var5) SetPiMcuTemp(val float32) {
	d.PiMcuTemp = val
	d.piMcuTempSet = true
}
func (d *DuetDataMk4Var5) String() string {
	return fmt.Sprintf("[Duet %d, Type 4.5 | Unix %d | Co %.2f, O3: %.2f, CH4: %.2f | %s | HTU: %s | SCD: %s | MPRLS: %s | SGP: %s | SPS30s: 1[%s], 2[%s|, M[%s] | Radio: %s | Errstate %d | PoE Voltage %d]",
		d.SerialNumber, d.UnixSec, d.Co, d.O3, d.No2, d.TempRh.String(), d.Htu.String(), d.Scd.String(), d.Mprls.String(), d.Sgp.String(), d.Sps1.String(), d.Sps2.String(), d.SpsM.String(),
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

	// SPS30s (as PMS5003)
	if err := d.Sps1.FromSerialString(splitStr[2]); err != nil {
		return fmt.Errorf("failed to convert sps30 string, %s, to PlantowerData", splitStr[3])
	}
	d.Sps2 = d.Sps1
	d.SpsM = d.Sps1

	// Temperatures (1 & 2)
	if temp, err := strconv.ParseFloat(splitStr[3], 32); err != nil {
		return fmt.Errorf("failed to convert htu temp string, %s, to float32", splitStr[5])
	} else {
		d.Htu.Temp = float32(temp)
	}

	if temp, err := strconv.ParseFloat(splitStr[4], 32); err != nil {
		return fmt.Errorf("failed to convert scd temp string, %s, to float32", splitStr[5])
	} else {
		d.Scd.Temp = float32(temp)
	}

	// Humidities (1 & 2)
	if hum, err := strconv.ParseFloat(splitStr[5], 32); err != nil {
		return fmt.Errorf("failed to convert htu hum string, %s, to float32", splitStr[5])
	} else {
		d.Htu.Hum = float32(hum)
	}

	if hum, err := strconv.ParseFloat(splitStr[6], 32); err != nil {
		return fmt.Errorf("failed to convert scd hum string, %s, to float32", splitStr[5])
	} else {
		d.Scd.Hum = float32(hum)
	}

	// Pressure
	if press, err := strconv.ParseFloat(splitStr[7], 32); err != nil {
		return fmt.Errorf("failed to convert pressure string, %s, to float32", splitStr[5])
	} else {
		d.Mprls.Pressure = float32(press)
	}

	// VOC Index
	if voc, err := strconv.ParseUint(splitStr[8], 10, 32); err != nil {
		return fmt.Errorf("failed to convert voc index string, %s, to uint32", splitStr[5])
	} else {
		d.Sgp.VocIndex = uint32(voc)
	}

	// CO2
	if co2, err := strconv.ParseUint(splitStr[9], 10, 16); err != nil {
		return fmt.Errorf("failed to convert co2 string, %s, to uint32", splitStr[5])
	} else {
		d.Scd.Co2 = uint16(co2)
	}

	// PoE / USB Voltage
	if voltage, err := strconv.ParseUint(splitStr[10], 10, 8); err != nil {
		return fmt.Errorf("failed to convert voltage string, %s, to uint8", splitStr[13])
	} else {
		d.PoeUsbVoltage = uint8(voltage)
	}

	// Sensor States
	if sensorStates, err := strconv.ParseUint(splitStr[11], 10, 8); err != nil {
		return fmt.Errorf("failed to convert states string, %s, to uint8", splitStr[14])
	} else {
		d.SensorStates = uint8(sensorStates)
	}

	// Gas Sensors Enabled
	if bitfield, err := strconv.ParseUint(splitStr[12], 10, 16); err != nil {
		return fmt.Errorf("failed to interperet substring, %s,  as uint16 for gas sensors enabled: %w", splitStr[14], err)
	} else {
		d.Gas.SensorBitField = uint16(bitfield)
	}

	if err := d.Gas.PopulateFromString(splitStr[13]); err != nil {
		return fmt.Errorf("failed to convert string to gas sensors: %w", err)
	}

	CombineTempRhMeasurements(d.Htu, d.Scd, &(d.TempRh))
	d.Co = d.Gas.Co
	d.O3 = d.Gas.O3
	d.No2 = d.Gas.No2

	return nil
}

func (d *DuetDataMk4Var5) doPopulateFromBytes(buff []byte) error {
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

	d.Gas.SensorBitField = binary.LittleEndian.Uint16(buff[70:72])
	if err := d.Gas.PopulateFromBytes(buff[34:70]); err != nil {
		return fmt.Errorf("error populating gas sensors from bytes: %w", err)
	}
	d.Co = d.Gas.Co
	d.O3 = d.Gas.O3
	d.No2 = d.Gas.No2
	if err := d.Sps1.PopulateFromBytes(buff[72:90]); err != nil {
		return fmt.Errorf("error parsing bytes for sps1: %w", err)
	}
	if err := d.Sps2.PopulateFromBytes(buff[90:108]); err != nil {
		return fmt.Errorf("error parsing bytes for sps2: %w", err)
	}
	if err := MergePT(&d.Sps1, &d.Sps2, &d.SpsM); err != nil {
		return fmt.Errorf("failed to merge sps: %w", err)
	}
	CombineTempRhMeasurements(d.Htu, d.Scd, &(d.TempRh))
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
	maps.Copy(ret, d.Sps1.ToMap("_t"))
	maps.Copy(ret, d.Sps2.ToMap("_b"))
	maps.Copy(ret, d.SpsM.ToMap("_m"))
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
