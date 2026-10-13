package telosairduetcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"maps"
	"strconv"
)

/* ~~ MK3 Var 0 ~~ */
var DuetTypeMk3Var0 = DuetTypeInfo{
	ExpectedBytes:        110,
	ExpectedStringLen:    14,
	StructInstanceGetter: func() DuetData { return &DuetDataMk3Var0{} },
	TypeAlias:            "Mk3.0",
}

type DuetDataMk3Var0 struct {
	SerialNumber   uint16
	SampleTimeMs   uint32
	UnixSec        uint32
	LastResetUnix  uint32
	SensorStates   uint8
	ConnectionType int
	PiMcuTemp      float32
	piMcuTempSet   bool

	Pt1       Pms5003Measurement
	Pt2       Pms5003Measurement
	PtM       Pms5003Measurement
	Scd       Scd41Measurement
	Htu       Htu21Measurement
	TempRh    CombinedTempRhMeasurements
	Mprls     MprlsMeasurement
	Sgp       Sgp40Measurement
	RadioMeta RadioMetadata

	timeResolved bool
}

func (d *DuetDataMk3Var0) TimeResolved() bool {
	return d.timeResolved
}

func (d *DuetDataMk3Var0) MarkTimeResolved(v bool) {
	d.timeResolved = v
}

func (d *DuetDataMk3Var0) Timestamp() uint32 {
	return d.UnixSec
}

func (d *DuetDataMk3Var0) ResolveTime(t uint32) {
	d.UnixSec = t
}

func (d *DuetDataMk3Var0) SensorMeasurements() []SensorMeasurement {
	return []SensorMeasurement{d.PtM, d.TempRh, d.Scd, d.Mprls, d.Sgp, DuetSensorState{d.SensorStates}}
}

func (d *DuetDataMk3Var0) SetRadioData(v RadioMetadata) {
	d.RadioMeta = v
}

func (d *DuetDataMk3Var0) SetPiMcuTemp(val float32) {
	d.PiMcuTemp = val
	d.piMcuTempSet = true
}

func (d *DuetDataMk3Var0) String() string {
	return fmt.Sprintf("[Duet %d, Type 3.0 | Unix %d | %s | HTU: %s | SCD: %s | MPRLS: %s | SGP: %s | PT1: %s | PT2: %s | PTM: %s | Radio: %s | Errstate %d]",
		d.SerialNumber, d.UnixSec, d.TempRh.String(), d.Htu.String(), d.Scd.String(), d.Mprls.String(), d.Sgp.String(), d.Pt1.String(), d.Pt2.String(), d.PtM.String(),
		d.RadioMeta.String(), d.SensorStates)
}

func (d *DuetDataMk3Var0) GetTypeInfo() DuetTypeInfo {
	return DuetTypeMk3Var0
}

func (d *DuetDataMk3Var0) SetConnectionType(ct int) {
	d.ConnectionType = ct
}

func (d *DuetDataMk3Var0) SetTimeRadio(unixSecRecieved uint32) error {
	if (d.RadioMeta.RadioSentTimeMs < d.SampleTimeMs) || (unixSecRecieved*d.RadioMeta.RadioSentTimeMs*unixSecRecieved == 0) {
		return fmt.Errorf("incompatible timekeeping parameters: unix: %d, radio sent ms: %d, sample ms: %d", unixSecRecieved, d.RadioMeta.RadioSentTimeMs, d.SampleTimeMs)
	}
	d.UnixSec = unixSecRecieved - ((d.RadioMeta.RadioSentTimeMs - d.SampleTimeMs) / 1000)
	return nil
}

func (d *DuetDataMk3Var0) SetTimeSerial(unixSecRecieved uint32) {
	d.UnixSec = unixSecRecieved
}

func (d *DuetDataMk3Var0) RecalculateLastResetUnix() {
	d.LastResetUnix = d.UnixSec - (d.SampleTimeMs / 1000)
}

func (d *DuetDataMk3Var0) doPopulateFromSubStrings(splitStr []string) error {
	sn, err := strconv.ParseUint(splitStr[0], 10, 16)
	if err != nil {
		return fmt.Errorf("failed to convert DuetSerialNumber string, %s, to uint16", splitStr[0])
	}
	d.SerialNumber = uint16(sn)

	st, err := strconv.ParseUint(splitStr[1], 10, 32)
	if err != nil {
		return fmt.Errorf("failed to convert SampleTime string, %s, to uint32", splitStr[1])
	}
	d.SampleTimeMs = uint32(st)

	if err := d.Pt1.FromSerialString(splitStr[2]); err != nil {
		return fmt.Errorf("failed to convert PT1 string, %s, to PlantowerData: %w", splitStr[2], err)
	}
	if err := d.Pt2.FromSerialString(splitStr[3]); err != nil {
		return fmt.Errorf("failed to convert PT2 string, %s, to PlantowerData: %w", splitStr[3], err)
	}

	if temp, err := strconv.ParseFloat(splitStr[4], 32); err != nil {
		return fmt.Errorf("failed to convert HTU temperature string, %s, to float32", splitStr[4])
	} else {
		d.Htu.Temp = float32(temp)
	}
	if temp, err := strconv.ParseFloat(splitStr[5], 32); err != nil {
		return fmt.Errorf("failed to convert SCD temperature string, %s, to float32", splitStr[5])
	} else {
		d.Scd.Temp = float32(temp)
	}
	if hum, err := strconv.ParseFloat(splitStr[6], 32); err != nil {
		return fmt.Errorf("failed to convert HTU humidity string, %s, to float32", splitStr[6])
	} else {
		d.Htu.Hum = float32(hum)
	}
	if hum, err := strconv.ParseFloat(splitStr[7], 32); err != nil {
		return fmt.Errorf("failed to convert SCD humidity string, %s, to float32", splitStr[7])
	} else {
		d.Scd.Hum = float32(hum)
	}
	if pressure, err := strconv.ParseFloat(splitStr[8], 32); err != nil {
		return fmt.Errorf("failed to convert pressure string, %s, to float32", splitStr[8])
	} else {
		d.Mprls.Pressure = float32(pressure)
	}
	if voc, err := strconv.ParseUint(splitStr[9], 10, 32); err != nil {
		return fmt.Errorf("failed to convert VOC index string, %s, to uint32", splitStr[9])
	} else {
		d.Sgp.VocIndex = uint32(voc)
	}
	if co2, err := strconv.ParseUint(splitStr[10], 10, 16); err != nil {
		return fmt.Errorf("failed to convert CO2 string, %s, to uint16", splitStr[10])
	} else {
		d.Scd.Co2 = uint16(co2)
	}
	if sensorStates, err := strconv.ParseUint(splitStr[11], 10, 8); err != nil {
		return fmt.Errorf("failed to convert sensor states string, %s, to uint8", splitStr[11])
	} else {
		d.SensorStates = uint8(sensorStates)
	}

	MergePT(&d.Pt1, &d.Pt2, &d.PtM)
	CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)
	return nil
}

func (d *DuetDataMk3Var0) doPopulateFromBytes(buff []byte) error {
	d.SensorStates = buff[0]
	// buff[1] is the firmware struct's explicit alignment byte.
	d.SerialNumber = binary.LittleEndian.Uint16(buff[2:4])
	d.Scd.Co2 = binary.LittleEndian.Uint16(buff[4:6])
	d.Sgp.VocIndex = binary.LittleEndian.Uint32(buff[6:10])
	d.SampleTimeMs = binary.LittleEndian.Uint32(buff[10:14])

	reader := bytes.NewReader(buff[14:34])
	if err := binary.Read(reader, binary.LittleEndian, &d.Htu.Temp); err != nil {
		return fmt.Errorf("error converting HTU temperature bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Scd.Temp); err != nil {
		return fmt.Errorf("error converting SCD temperature bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Htu.Hum); err != nil {
		return fmt.Errorf("error converting HTU humidity bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Scd.Hum); err != nil {
		return fmt.Errorf("error converting SCD humidity bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Mprls.Pressure); err != nil {
		return fmt.Errorf("error converting pressure bytes to float: %w", err)
	}

	// The Mk3.0 wire struct retains 36 bytes for legacy gas readings and two
	// bytes for a gas-type bitfield. Current firmware does not populate them.
	if err := d.Pt1.PopulateFromBytes(buff[72:90]); err != nil {
		return fmt.Errorf("error parsing bytes for PT1: %w", err)
	}
	if err := d.Pt2.PopulateFromBytes(buff[90:108]); err != nil {
		return fmt.Errorf("error parsing bytes for PT2: %w", err)
	}

	MergePT(&d.Pt1, &d.Pt2, &d.PtM)
	CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)
	return nil
}

func (d *DuetDataMk3Var0) ToMap(gatewaySerial string) map[string]any {
	ret := map[string]any{
		KEY_DEVICE_TYPE:     3.0,
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
	maps.Copy(ret, d.Pt1.ToMap("_t"))
	maps.Copy(ret, d.Pt2.ToMap("_b"))
	maps.Copy(ret, d.PtM.ToMap("_m"))
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
