package telosairduetcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"maps"
	"strconv"
)

// func RoundFloatTwoDecimals(val float32) float32 {
// 	return float32(math.Round(float64(val*100)) / 100)
// }

/* ~~ MK4 Var 27 ~~ */
var DuetTypeMk4Var27 = DuetTypeInfo{
	ExpectedBytes:        62,
	ExpectedStringLen:    16,
	StructInstanceGetter: func() DuetData { return &DuetDataMk4Var27{} },
	TypeAlias:            "Mk4.27",
}

type DuetDataMk4Var27 struct {
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
	Pid       PidMeasurement
	RadioMeta RadioMetadata

	timeResolved bool
}

func (d *DuetDataMk4Var27) TimeResolved() bool {
	return d.timeResolved
}
func (d *DuetDataMk4Var27) MarkTimeResolved(v bool) {
	d.timeResolved = v
}
func (d *DuetDataMk4Var27) Timestamp() uint32 {
	return d.UnixSec
}
func (d *DuetDataMk4Var27) ResolveTime(t uint32) {
	d.UnixSec = t
}

func (d *DuetDataMk4Var27) SensorMeasurements() []SensorMeasurement {
	return []SensorMeasurement{d.Sps, d.TempRh, d.Scd, d.Mprls, d.Sgp, DuetSensorState{d.SensorStates}}
}

func (d *DuetDataMk4Var27) SetRadioData(v RadioMetadata) {
	d.RadioMeta = v
}
func (d *DuetDataMk4Var27) SetPiMcuTemp(val float32) {
	d.PiMcuTemp = val
	d.piMcuTempSet = true
}
// func (d *DuetDataMk4Var27) String() string {
// 	return fmt.Sprintf("[Duet %d, Type %d.%d | Unix %d | %s | HTU: %s | SCD: %s | MPRLS: %s | SGP: %s | SPS: %s | Radio: %s | Errstate %d | PoE Voltage %d]",
// 		d.SerialNumber, 4, 27, d.UnixSec, d.TempRh.String(), d.Htu.String(), d.Scd.String(), d.Mprls.String(), d.Sgp.String(), d.Sps.String(),
// 		d.RadioMeta.String(), d.SensorStates, d.PoeUsbVoltage)
// }
// func (d *DuetDataMk4Var27) String() string {
//     return fmt.Sprintf("[Duet %d, Type %d.%d | Unix %d | %s | HTU: %s | SCD: %s | MPRLS: %s | SGP: %s | PID: ppb:%.2f raw:%.2fmV | SPS: %s | Radio: %s | Errstate %d | PoE Voltage %d]",
//         d.SerialNumber, 4, 27, d.UnixSec, d.TempRh.String(), d.Htu.String(), d.Scd.String(), d.Mprls.String(), d.Sgp.String(), 
//         d.Pid.Ppb, d.Pid.RawMv, // <-- Added PID values here
//         d.Sps.String(), d.RadioMeta.String(), d.SensorStates, d.PoeUsbVoltage)
// }

func (d *DuetDataMk4Var27) String() string {
    return fmt.Sprintf("[Duet %d, Type %d.%d | Unix %d | %s | HTU: %s | SCD: %s | MPRLS: %s | SGP: %s | PID: PIDX004:%.2fmV ev:%.2fmV | SPS: %s | Radio: %s | Errstate %d | PoE Voltage %d]",
        d.SerialNumber, 4, 27, d.UnixSec, d.TempRh.String(), d.Htu.String(), d.Scd.String(), d.Mprls.String(), d.Sgp.String(), 
        d.Pid.RawMV, d.Pid.EvMV, // Cleaned up field names and ordered logically (Signal first, Ref second)
        d.Sps.String(), d.RadioMeta.String(), d.SensorStates, d.PoeUsbVoltage)
}
func (d *DuetDataMk4Var27) GetTypeInfo() DuetTypeInfo {
	return DuetTypeMk4Var27
}

func (d *DuetDataMk4Var27) SetConnectionType(ct int) {
	d.ConnectionType = ct
}
func (d *DuetDataMk4Var27) SetTimeRadio(unixSecRecieved uint32) error {
	if (d.RadioMeta.RadioSentTimeMs < d.SampleTimeMs) || (unixSecRecieved*d.RadioMeta.RadioSentTimeMs*unixSecRecieved == 0) {
		return fmt.Errorf("incompatible timekeeping parameters: unix: %d, radio sent ms: %d, sample ms: %d", unixSecRecieved, d.RadioMeta.RadioSentTimeMs, d.SampleTimeMs)
	}
	d.UnixSec = unixSecRecieved - ((d.RadioMeta.RadioSentTimeMs - d.SampleTimeMs) / 1000)
	return nil
}

func (d *DuetDataMk4Var27) SetTimeSerial(unixSecRecieved uint32) {
	d.UnixSec = unixSecRecieved
}

func (d *DuetDataMk4Var27) RecalculateLastResetUnix() {
	d.LastResetUnix = d.UnixSec - (d.SampleTimeMs / 1000)
}

func (d *DuetDataMk4Var27) doPopulateFromSubStrings(splitStr []string) error {
    if len(splitStr) < 14 {
        return fmt.Errorf("insufficient data points: got %d, want at least 14", len(splitStr))
    }

    // Serial Number [0]
    sn, err := strconv.ParseUint(splitStr[0], 10, 16)
    if err != nil {
        return fmt.Errorf("failed to convert DuetSerialNumber string, %s, to uint16", splitStr[0])
    }
    d.SerialNumber = uint16(sn)

    // Sample Time [1]
    st, err := strconv.ParseUint(splitStr[1], 10, 32)
    if err != nil {
        return fmt.Errorf("failed to convert SampleTime string, %s, to uint32", splitStr[1])
    }
    d.SampleTimeMs = uint32(st)

    // SPS30 Data [2] - Re-enabled!
    if err := d.Sps.FromSerialString(splitStr[2]); err != nil {
        return fmt.Errorf("failed to convert SPS string, %s, to PlantowerData", splitStr[2])
    }

    // Temperatures (HTU then SCD) [3, 4]
    if temp, err := strconv.ParseFloat(splitStr[3], 32); err != nil {
        return fmt.Errorf("failed to convert htu temp string, %s, to float32", splitStr[3])
    } else {
        d.Htu.Temp = float32(temp)
    }

    if temp, err := strconv.ParseFloat(splitStr[4], 32); err != nil {
        return fmt.Errorf("failed to convert scd temp string, %s, to float32", splitStr[4])
    } else {
        d.Scd.Temp = float32(temp)
    }

    // Humidities (HTU then SCD) [5, 6]
    if hum, err := strconv.ParseFloat(splitStr[5], 32); err != nil {
        return fmt.Errorf("failed to convert htu hum string, %s, to float32", splitStr[5])
    } else {
        d.Htu.Hum = float32(hum)
    }

    if hum, err := strconv.ParseFloat(splitStr[6], 32); err != nil {
        return fmt.Errorf("failed to convert scd hum string, %s, to float32", splitStr[6])
    } else {
        d.Scd.Hum = float32(hum)
    }

    // Pressure [7]
    if press, err := strconv.ParseFloat(splitStr[7], 32); err != nil {
        return fmt.Errorf("failed to convert pressure string, %s, to float32", splitStr[7])
    } else {
        d.Mprls.Pressure = float32(press)
    }

    // SGP40 VOC Index [8]
    if voc, err := strconv.ParseUint(splitStr[8], 10, 32); err != nil {
        return fmt.Errorf("failed to convert voc index string, %s, to uint32", splitStr[8])
    } else {
        d.Sgp.VocIndex = uint32(voc)
    }

    // SCD41 CO2 [9]
    if co2, err := strconv.ParseUint(splitStr[9], 10, 16); err != nil {
        return fmt.Errorf("failed to convert co2 string, %s, to uint16", splitStr[9])
    } else {
        d.Scd.Co2 = uint16(co2)
    }

    // PID Values [10, 11]
    if rawMv, err := strconv.ParseFloat(splitStr[10], 32); err != nil {
        return fmt.Errorf("failed to convert pid RawMv string, %s, to float32", splitStr[10])
    } else {
        d.Pid.RawMV = float32(rawMv) 
    }

    if evMv, err := strconv.ParseFloat(splitStr[11], 32); err != nil {
        return fmt.Errorf("failed to convert pid EvMv string, %s, to float32", splitStr[11])
    } else {
        d.Pid.EvMV = float32(evMv) 
    }

    // PoE / USB Voltage [13]
    if voltage, err := strconv.ParseUint(splitStr[12], 10, 8); err != nil {
        return fmt.Errorf("failed to convert voltage string, %s, to uint8", splitStr[12])
    } else {
        d.PoeUsbVoltage = uint8(voltage)
    }

    // Sensor States [14]
    if sensorStates, err := strconv.ParseUint(splitStr[13], 10, 8); err != nil {
        return fmt.Errorf("failed to convert states string, %s, to uint8", splitStr[13])
    } else {
        d.SensorStates = uint8(sensorStates)
    }

    CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)

    return nil
}

func (d *DuetDataMk4Var27) doPopulateFromBytes(buff []byte) error {
	d.SensorStates = buff[0]
	d.PoeUsbVoltage = buff[1]
	d.SerialNumber = binary.LittleEndian.Uint16(buff[2:4])
	d.Scd.Co2 = binary.LittleEndian.Uint16(buff[4:6])
	d.Sgp.VocIndex = binary.LittleEndian.Uint32(buff[6:10])
	d.SampleTimeMs = binary.LittleEndian.Uint32(buff[10:14])

	reader := bytes.NewReader(buff[14:42])
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
	if err := binary.Read(reader, binary.LittleEndian, &d.Pid.RawMV); err != nil {
		return fmt.Errorf("error converting bytes to float (Pid.RawMV): %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Pid.EvMV); err != nil {
		return fmt.Errorf("error converting bytes to float (Pid.EvMV): %w", err)
	}
	CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)

	return nil
}

func (d *DuetDataMk4Var27) ToMap(gatewaySerial string) map[string]any {
    ret := map[string]any{
        KEY_DEVICE_TYPE:     4.27,
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
    
    maps.Copy(ret, d.Sps.ToMap("_t"))
    maps.Copy(ret, d.Sps.ToMap("_b"))
    maps.Copy(ret, d.Sps.ToMap("_m"))
    maps.Copy(ret, d.Htu.ToMap())
    maps.Copy(ret, d.Scd.ToMap())
    maps.Copy(ret, d.TempRh.ToMap())
    maps.Copy(ret, d.Mprls.ToMap())
    maps.Copy(ret, d.Sgp.ToMap())
    
    if pidMap := d.Pid.ToMap(); pidMap != nil { // Adjust if you do maps.Copy(ret, d.Pid.ToMap()) directly
        maps.Copy(ret, pidMap)
    }
    
    maps.Copy(ret, d.RadioMeta.ToMap())
    
    if d.piMcuTempSet {
        ret[KEY_PI_MCU_TEMP] = d.PiMcuTemp
    }

    return ret
}
