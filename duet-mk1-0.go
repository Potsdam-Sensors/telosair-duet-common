package telosairduetcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"maps"
	"strconv"
)

/* ~~ MK1 Var 0 ~~ */
var DuetTypeMk1Var0 = DuetTypeInfo{
	ExpectedBytes:        64,
	ExpectedStringLen:    12,
	StructInstanceGetter: func() DuetData { return &DuetDataMk1Var0{} },
	TypeAlias:            "Mk1.0",
}

/*
	typedef struct __attribute__((packed, aligned(4))) {
	  uint8_t hardwareVersion;
	  uint8_t sensorVariation;
	  uint8_t sensorStates;
	  uint8_t _; //padding to align on 4, TODO: why not just remove `packed`?
	  uint16_t id;
	  uint16_t co2;
	  uint32_t tvoc;
	  unsigned long sample_time;
	  float temperature;
	  float humidity;
	  float pressure;
	  PMS5003::pms5003_data_t pt1;
	  PMS5003::pms5003_data_t pt2;
	} duet_sensor_reading_mk1_t;
*/
type DuetDataMk1Var0 struct {
	SerialNumber   uint16
	SampleTimeMs   uint32
	UnixSec        uint32
	LastResetUnix  uint32
	SensorStates   uint8
	ConnectionType int
	PiMcuTemp      float32
	piMcuTempSet   bool

	Pt1 Pms5003Measurement
	Pt2 Pms5003Measurement
	PtM Pms5003Measurement

	Si        Si7021Measurement
	Mprls     MprlsMeasurement
	Sgp       Sgp40Measurement
	Co2       PlantowerCo2Measurement
	RadioMeta RadioMetadata

	timeResolved bool
}

func (d *DuetDataMk1Var0) TimeResolved() bool {
	return d.timeResolved
}
func (d *DuetDataMk1Var0) MarkTimeResolved(v bool) {
	d.timeResolved = v
}
func (d *DuetDataMk1Var0) Timestamp() uint32 {
	return d.UnixSec
}
func (d *DuetDataMk1Var0) ResolveTime(t uint32) {
	d.UnixSec = t
}

func (d *DuetDataMk1Var0) SensorMeasurements() []SensorMeasurement {
	return []SensorMeasurement{d.PtM, d.Si, d.Co2, d.Mprls, d.Sgp, DuetSensorState{d.SensorStates}}
}

func (d *DuetDataMk1Var0) SetRadioData(v RadioMetadata) {
	d.RadioMeta = v
}
func (d *DuetDataMk1Var0) SetPiMcuTemp(val float32) {
	d.PiMcuTemp = val
	d.piMcuTempSet = true
}
func (d *DuetDataMk1Var0) String() string {
	return fmt.Sprintf("[Duet %d, Type %d.%d | Unix %d | Si7021 %s | PT CO2: %s | MPRLS: %s | SGP: %s | PT1: %s | PT2: %s | PTM: %s | Radio: %s | Errstate %d ]",
		d.SerialNumber, 1, 0, d.UnixSec, d.Si.String(), d.Co2.String(), d.Mprls.String(), d.Sgp.String(), d.Pt1.String(), d.Pt2.String(), d.PtM.String(),
		d.RadioMeta.String(), d.SensorStates)
}
func (d *DuetDataMk1Var0) GetTypeInfo() DuetTypeInfo {
	return DuetTypeMk1Var0
}

func (d *DuetDataMk1Var0) SetConnectionType(ct int) {
	d.ConnectionType = ct
}
func (d *DuetDataMk1Var0) SetTimeRadio(unixSecRecieved uint32) error {
	if (d.RadioMeta.RadioSentTimeMs < d.SampleTimeMs) || (unixSecRecieved*d.RadioMeta.RadioSentTimeMs*unixSecRecieved == 0) {
		return fmt.Errorf("incompatible timekeeping parameters: unix: %d, radio sent ms: %d, sample ms: %d", unixSecRecieved, d.RadioMeta.RadioSentTimeMs, d.SampleTimeMs)
	}
	d.UnixSec = unixSecRecieved - ((d.RadioMeta.RadioSentTimeMs - d.SampleTimeMs) / 1000)
	return nil
}

func (d *DuetDataMk1Var0) SetTimeSerial(unixSecRecieved uint32) {
	d.UnixSec = unixSecRecieved
}

func (d *DuetDataMk1Var0) RecalculateLastResetUnix() {
	d.LastResetUnix = d.UnixSec - (d.SampleTimeMs / 1000)
}

func (d *DuetDataMk1Var0) doPopulateFromSubStrings(splitStr []string) error {
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

	// Temperature
	if temp, err := strconv.ParseFloat(splitStr[4], 32); err != nil {
		return fmt.Errorf("failed to convert htu temp string, %s, to float32", splitStr[5])
	} else {
		d.Si.Temp = float32(temp)
	}

	// Humidity
	if hum, err := strconv.ParseFloat(splitStr[5], 32); err != nil {
		return fmt.Errorf("failed to convert htu hum string, %s, to float32", splitStr[5])
	} else {
		d.Si.Hum = float32(hum)
	}

	// Pressure
	if press, err := strconv.ParseFloat(splitStr[6], 32); err != nil {
		return fmt.Errorf("failed to convert pressure string, %s, to float32", splitStr[5])
	} else {
		d.Mprls.Pressure = float32(press)
	}

	// VOC Index
	if voc, err := strconv.ParseUint(splitStr[7], 10, 32); err != nil {
		return fmt.Errorf("failed to convert voc index string, %s, to uint32", splitStr[5])
	} else {
		d.Sgp.VocIndex = uint32(voc)
	}

	// CO2
	if co2, err := strconv.ParseUint(splitStr[8], 10, 16); err != nil {
		return fmt.Errorf("failed to convert co2 string, %s, to uint32", splitStr[5])
	} else {
		d.Co2.Co2 = uint16(co2)
	}

	// Sensor States
	if sensorStates, err := strconv.ParseUint(splitStr[9], 10, 8); err != nil {
		return fmt.Errorf("failed to convert states string, %s, to uint8", splitStr[14])
	} else {
		d.SensorStates = uint8(sensorStates)
	}
	MergePT(&d.Pt1, &d.Pt2, &d.PtM)

	return nil
}

/*
	typedef struct __attribute__((packed, aligned(4))) {
	  uint8_t hardwareVersion;
	  uint8_t sensorVariation;
	  uint8_t sensorStates;
	  uint8_t _; //padding to align on 4, TODO: why not just remove `packed`?
	  uint16_t id;
	  uint16_t co2;
	  uint32_t tvoc;
	  unsigned long sample_time;
	  float temperature;
	  float humidity;
	  float pressure;
	  PMS5003::pms5003_data_t pt1;
	  PMS5003::pms5003_data_t pt2;
	} duet_sensor_reading_mk1_t;
*/
func (d *DuetDataMk1Var0) doPopulateFromBytes(buff []byte) error {
	d.SensorStates = buff[0]
	// Skip buff[1] as it is padding
	d.SerialNumber = binary.LittleEndian.Uint16(buff[2:4])
	d.Co2.Co2 = binary.LittleEndian.Uint16(buff[4:6])
	d.Sgp.VocIndex = binary.LittleEndian.Uint32(buff[6:10])
	d.SampleTimeMs = binary.LittleEndian.Uint32(buff[10:14])

	reader := bytes.NewReader(buff[14:26])
	if err := binary.Read(reader, binary.LittleEndian, &d.Si.Temp); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Si.Hum); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Mprls.Pressure); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}

	if err := d.Pt1.PopulateFromBytes(buff[26:44]); err != nil {
		return fmt.Errorf("error parsing bytes for PT: %w", err)
	}
	if err := d.Pt2.PopulateFromBytes(buff[44:62]); err != nil {
		return fmt.Errorf("error parsing bytes for PT: %w", err)
	}
	MergePT(&d.Pt1, &d.Pt2, &d.PtM)

	return nil
}
func (d *DuetDataMk1Var0) ToMap(gatewaySerial string) map[string]any {
	ret := map[string]any{
		KEY_DEVICE_TYPE:     1.0,
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
	maps.Copy(ret, d.Si.ToMap())
	maps.Copy(ret, d.Si.TempRh().ToMap())
	maps.Copy(ret, d.Co2.ToMap())
	maps.Copy(ret, d.Mprls.ToMap())
	maps.Copy(ret, d.Sgp.ToMap())
	maps.Copy(ret, d.RadioMeta.ToMap())
	if d.piMcuTempSet {
		ret[KEY_PI_MCU_TEMP] = d.PiMcuTemp
	}

	return ret
}
