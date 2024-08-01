package duetcommongo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"maps"
	"strconv"
	"strings"
)

type DuetData interface {
	doPopulateFromBytes(buff []byte) error
	SetTimeRadio(unixSecRecieved uint32) error
	SetTimeSerial(unixSecRecieved uint32)
	doPopulateFromSubStrings(splitStr []string) error
	SetConnectionType(n int)
	ToMap(gatewaySerial string) map[string]any
	RecalculateLastResetUnix()
	GetTypeInfo() DuetTypeInfo
	// String() string
}

func PopulateFromSerialString(d DuetData, s string, recievedUnixSec uint32) error {
	typeInfo := d.GetTypeInfo()

	/* Validate Arguments */
	splitStr := strings.Split(strings.TrimSpace(s), " ")
	if err := typeInfo.CheckSubstringLen(len(splitStr)); err != nil {
		return err
	}

	/* Field Population */
	// Use the string split up by separator to populate data sample fields
	if err := d.doPopulateFromSubStrings(splitStr); err != nil {
		return err
	}

	// USB specific stuff
	d.SetConnectionType(CONNECTION_TYPE_USB_SERIAL)
	d.SetTimeSerial(recievedUnixSec)

	d.RecalculateLastResetUnix()

	return nil

}
func PopulateFromRadioBytes(d DuetData, buff []byte, recievedUnixSec uint32) error {
	typeInfo := d.GetTypeInfo()

	/* Validate Arguments */
	if err := typeInfo.CheckByteLen(len(buff)); err != nil {
		return err
	}

	/* Field Population */
	// Use the buffer to populate data sample fields
	if err := d.doPopulateFromBytes(buff); err != nil {
		return fmt.Errorf("error populating fields from bytes: %w", err)
	}

	// Set Radio Specific Stuff
	if err := d.SetTimeRadio(recievedUnixSec); err != nil {
		return err
	}
	d.SetConnectionType(CONNECTION_TYPE_LORA_GATEWAY)

	d.RecalculateLastResetUnix()

	return nil
}

type DuetTypeInfo struct {
	ExpectedBytes        int
	ExpectedStringLen    int
	StructInstanceGetter func() DuetData
	TypeAlias            string
}

func (typeInfo DuetTypeInfo) CheckByteLen(byteLen int) error {
	if byteLen != typeInfo.ExpectedBytes {
		return fmt.Errorf("exepcted at least %d bytes for sample, only got %d", typeInfo.ExpectedBytes, byteLen)
	}
	return nil
}
func (typeInfo DuetTypeInfo) CheckSubstringLen(n int) error {
	if n != typeInfo.ExpectedStringLen {
		return fmt.Errorf("expected a list of values at least %d in length, only got %d", typeInfo.ExpectedStringLen, n)
	}
	return nil
}

/* =========== MK4 =========== */

/* ~~ MK4 Var 0 ~~ */
var DuetTypeMk4Var0 = DuetTypeInfo{
	ExpectedBytes:        70,
	ExpectedStringLen:    15,
	StructInstanceGetter: func() DuetData { return &DuetDataMk4Var0{} },
}

type DuetDataMk4Var0 struct {
	SerialNumber   uint16
	SampleTimeMs   uint32
	UnixSec        uint32
	LastResetUnix  uint32
	SensorStates   uint8
	PoeUsbVoltage  uint8
	ConnectionType int

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

/* ~~ MK4 Var 3 (Indoor, 1 SPS30, CO, NO2, CH4) ~~ */
var DuetTypeMk4Var3 = DuetTypeInfo{
	ExpectedBytes:        90,
	ExpectedStringLen:    16,
	StructInstanceGetter: func() DuetData { return &DuetDataMk4Var3{} },
}

type DuetDataMk4Var3 struct {
	SerialNumber   uint16
	SampleTimeMs   uint32
	UnixSec        uint32
	LastResetUnix  uint32
	SensorStates   uint8
	PoeUsbVoltage  uint8
	ConnectionType int

	Sps       Sps30Measurement
	Scd       Scd41Measurement
	Htu       Htu21Measurement
	TempRh    CombinedTempRhMeasurements
	Mprls     MprlsMeasurement
	Sgp       Sgp40Measurement
	RadioMeta RadioMetadata

	Co, No2, Ch4 float32
}

func (d *DuetDataMk4Var3) GetTypeInfo() DuetTypeInfo {
	return DuetTypeMk4Var3
}

func (d *DuetDataMk4Var3) SetConnectionType(ct int) {
	d.ConnectionType = ct
}
func (d *DuetDataMk4Var3) SetTimeRadio(unixSecRecieved uint32) error {
	if (d.RadioMeta.RadioSentTimeMs < d.SampleTimeMs) || (unixSecRecieved*d.RadioMeta.RadioSentTimeMs*unixSecRecieved == 0) {
		return fmt.Errorf("incompatible timekeeping parameters: unix: %d, radio sent ms: %d, sample ms: %d", unixSecRecieved, d.RadioMeta.RadioSentTimeMs, d.SampleTimeMs)
	}
	d.UnixSec = unixSecRecieved - ((d.RadioMeta.RadioSentTimeMs - d.SampleTimeMs) / 1000)
	return nil
}

func (d *DuetDataMk4Var3) SetTimeSerial(unixSecRecieved uint32) {
	d.UnixSec = unixSecRecieved
}

func (d *DuetDataMk4Var3) RecalculateLastResetUnix() {
	d.LastResetUnix = d.UnixSec - (d.SampleTimeMs / 1000)
}

func (d *DuetDataMk4Var3) doPopulateFromSubStrings(splitStr []string) error {
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

	// Sensirion SPS30
	if err := d.Sps.FromSerialString(splitStr[2]); err != nil {
		return fmt.Errorf("failed to convert PT0 string, %s, to PlantowerData", splitStr[3])
	}

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
	gasSensors := GasSensorsMeasurement{}

	if bitfield, err := strconv.ParseUint(splitStr[12], 10, 16); err != nil {
		return fmt.Errorf("failed to interperet substring, %s,  as uint16 for gas sensors enabled: %w", splitStr[14], err)
	} else {
		gasSensors.SensorBitField = uint16(bitfield)
	}

	if err := gasSensors.PopulateFromString(splitStr[13]); err != nil {
		return fmt.Errorf("failed to convert string to gas sensors: %w", err)
	}

	d.Co = gasSensors.Co
	d.Ch4 = gasSensors.Ch4
	d.No2 = gasSensors.No2

	return nil
}

func (d *DuetDataMk4Var3) doPopulateFromBytes(buff []byte) error {

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

	gasSensors := GasSensorsMeasurement{
		SensorBitField: binary.LittleEndian.Uint16(buff[70:72]),
	}
	if err := gasSensors.PopulateFromBytes(buff[34:70]); err != nil {
		return fmt.Errorf("error populating gas sensors from bytes: %w", err)
	}

	if err := d.Sps.PopulateFromBytes(buff[72:90]); err != nil {
		return fmt.Errorf("error parsing bytes for sps30: %w", err)
	}
	CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)

	return nil
}
func (d *DuetDataMk4Var3) ToMap(gatewaySerial string) map[string]any {
	ret := map[string]any{
		KEY_DEVICE_TYPE:     4.3,
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
		KEY_GAS_CH4:         d.Ch4,
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

	// return ret
	return map[string]any{}
}
