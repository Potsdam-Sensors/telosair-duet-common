package telosairduetcommon

import (
	"encoding/binary"
	"math"
	"testing"
)

func putMk3Var0Float(buff []byte, offset int, value float32) {
	binary.LittleEndian.PutUint32(buff[offset:offset+4], math.Float32bits(value))
}

func putMk3Var0Pms(buff []byte, offset int, values []uint16) {
	for i, value := range values {
		binary.LittleEndian.PutUint16(buff[offset+i*2:offset+i*2+2], value)
	}
}

func TestMk3Var0FromRadioBytes(t *testing.T) {
	packet := make([]byte, 112)
	packet[0] = 3
	packet[1] = 0
	packet[2] = 0x2a
	binary.LittleEndian.PutUint16(packet[4:6], 1234)
	binary.LittleEndian.PutUint16(packet[6:8], 876)
	binary.LittleEndian.PutUint32(packet[8:12], 42)
	binary.LittleEndian.PutUint32(packet[12:16], 12_000)
	putMk3Var0Float(packet, 16, 20.5)
	putMk3Var0Float(packet, 20, 22.5)
	putMk3Var0Float(packet, 24, 40.0)
	putMk3Var0Float(packet, 28, 44.0)
	putMk3Var0Float(packet, 32, 1001.25)
	putMk3Var0Pms(packet, 74, []uint16{10, 20, 30, 40, 50, 60, 70, 80, 90})
	putMk3Var0Pms(packet, 92, []uint16{20, 40, 60, 80, 100, 120, 140, 160, 180})

	parsed, err := DuetDataFromRadioBytes(packet, 1_700_000_000, true, false)
	if err != nil {
		t.Fatalf("DuetDataFromRadioBytes() error = %v", err)
	}
	d, ok := parsed.(*DuetDataMk3Var0)
	if !ok {
		t.Fatalf("DuetDataFromRadioBytes() type = %T, want *DuetDataMk3Var0", parsed)
	}

	if d.SerialNumber != 1234 || d.SampleTimeMs != 12_000 || d.SensorStates != 0x2a {
		t.Fatalf("decoded identifiers = serial %d, sample time %d, states %d", d.SerialNumber, d.SampleTimeMs, d.SensorStates)
	}
	if d.Scd.Co2 != 876 || d.Sgp.VocIndex != 42 || d.Mprls.Pressure != 1001.25 {
		t.Fatalf("decoded I2C data = SCD %+v, SGP %+v, MPRLS %+v", d.Scd, d.Sgp, d.Mprls)
	}
	if d.TempRh.Temp != 21.5 || d.TempRh.Hum != 42.0 {
		t.Fatalf("combined temperature/RH = %+v", d.TempRh)
	}
	if d.Pt1.PM2p5 != 20 || d.Pt2.PM2p5 != 40 || d.PtM.PM2p5 != 30 {
		t.Fatalf("decoded/merged PMS5003 data = PT1 %+v, PT2 %+v, PTM %+v", d.Pt1, d.Pt2, d.PtM)
	}
	if d.UnixSec != 1_700_000_000 || d.LastResetUnix != 1_699_999_988 {
		t.Fatalf("decoded time = unix %d, last reset %d", d.UnixSec, d.LastResetUnix)
	}
}

func TestMk3Var0FromSerialString(t *testing.T) {
	const sample = "3 0 1234 12000 [10,20,30,40,50,60,70,80,90] [20,40,60,80,100,120,140,160,180] 20.5 22.5 40 44 1001.25 42 876 42"

	parsed, err := DuetDataFromSerialString(sample, 1_700_000_000, true)
	if err != nil {
		t.Fatalf("DuetDataFromSerialString() error = %v", err)
	}
	d, ok := parsed.(*DuetDataMk3Var0)
	if !ok {
		t.Fatalf("DuetDataFromSerialString() type = %T, want *DuetDataMk3Var0", parsed)
	}

	if d.SerialNumber != 1234 || d.Scd.Co2 != 876 || d.Sgp.VocIndex != 42 {
		t.Fatalf("decoded serial sample = %+v", d)
	}
	if d.PtM.PM2p5 != 30 || d.TempRh.Temp != 21.5 || d.TempRh.Hum != 42.0 {
		t.Fatalf("derived measurements = PTM %+v, TempRh %+v", d.PtM, d.TempRh)
	}
	if got := d.ToMap("gateway")[KEY_DEVICE_TYPE]; got != 3.0 {
		t.Fatalf("ToMap()[%q] = %v, want 3.0", KEY_DEVICE_TYPE, got)
	}
}
