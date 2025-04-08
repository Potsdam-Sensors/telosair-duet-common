package telosairduetcommon

import (
	"fmt"
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
	String() string
	SetPiMcuTemp(val float32)
	SetRadioData(v RadioMetadata)
	SensorMeasurements() []SensorMeasurement
}

func getVersionFromBuffer(b []byte) (*DuetTypeInfo, error) {
	if len(b) < 2 {
		return nil, fmt.Errorf("values are less than 2in length")
	}
	hwVer := b[0]
	snVar := b[1]
	typeInfo := getTypeInfo(hwVer, snVar)
	if typeInfo == nil {
		return nil, fmt.Errorf("failed to match recieved duet type: %d.%d", hwVer, snVar)
	}
	return typeInfo, nil
}

func getVersionFromString(s string) (splitStr []string, typeInfo *DuetTypeInfo, err error) {
	splitStr = strings.Split(strings.TrimSpace(s), " ")
	var hwVer, snsVar uint8
	if len(splitStr) < 2 {
		err = fmt.Errorf("values are less than 2 in length")
		return
	}
	if hwVer32, cErr := strconv.ParseUint(splitStr[0], 10, 8); cErr != nil {
		err = fmt.Errorf("error coverting to uint8: %w", cErr)
		return
	} else {
		hwVer = uint8(hwVer32)
	}
	if snsVar32, cErr := strconv.ParseUint(splitStr[1], 10, 8); cErr != nil {
		err = fmt.Errorf("error coverting to uint8: %w", cErr)
		return
	} else {
		snsVar = uint8(snsVar32)
	}
	typeInfo = getTypeInfo(hwVer, snsVar)
	if typeInfo == nil {
		err = fmt.Errorf("failed to match recieved duet type: Mk%d.%d", hwVer, snsVar)
	}
	return
}

func getTypeInfo(hwVer, snsVar uint8) (ret *DuetTypeInfo) {
	switch hwVer {
	case 1:
		switch snsVar {
		case 0:
			ret = &DuetTypeMk1Var0
		}
	case 4:
		switch snsVar {
		case 0:
			ret = &DuetTypeMk4Var0
		case 1:
			ret = &DuetTypeMk4Var1
		case 2:
			ret = &DuetTypeMk4Var2
		case 3:
			ret = &DuetTypeMk4Var3
		case 4:
			ret = &DuetTypeMk4Var4
		case 5:
			ret = &DuetTypeMk4Var5
		case 6:
			ret = &DuetTypeMk4Var6
		case 7:
			ret = &DuetTypeMk4Var7
		case 8:
			ret = &DuetTypeMk4Var8
		case 9:
			ret = &DuetTypeMk4Var9
		case 10:
			ret = &DuetTypeMk4Var10
		}
	}
	return
}

func DuetDataFromRadioBytes(buff []byte, recievedUnixSec uint32, isRadio bool) (DuetData, error) {
	typeInfo, err := getVersionFromBuffer(buff)
	if err != nil {
		return nil, fmt.Errorf("failed to get duet type info: %w", err)
	}
	if err := typeInfo.checkByteLen(len(buff[2:])); err != nil {
		return nil, err
	}
	d := typeInfo.StructInstanceGetter()

	if err := d.doPopulateFromBytes(buff[2:]); err != nil {
		return nil, fmt.Errorf("failed to populate for type %s: %w", typeInfo.TypeAlias, err)
	}

	if isRadio {
		d.SetConnectionType(CONNECTION_TYPE_LORA_GATEWAY)
		d.SetTimeRadio(recievedUnixSec)
	} else {
		d.SetConnectionType(CONNECTION_TYPE_USB_SERIAL)
		d.SetTimeSerial(recievedUnixSec)
	}
	d.RecalculateLastResetUnix()

	return d, nil
}

func DuetDataFromSerialString(s string, recievedUnixSec uint32) (DuetData, error) {
	/* Validate Arguments */
	splitStr, typeInfo, err := getVersionFromString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to get duet type info: %w", err)
	}

	if err := typeInfo.checkSubstringLen(len(splitStr)); err != nil {
		return nil, fmt.Errorf("failed to parse for type %s: %w", typeInfo.TypeAlias, err)
	}

	d := typeInfo.StructInstanceGetter()

	/* Field Population */
	// Use the string split up by separator to populate data sample fields
	if err := d.doPopulateFromSubStrings(splitStr[2:]); err != nil {
		return nil, fmt.Errorf("failed to populate for type %s: %w", typeInfo.TypeAlias, err)
	}

	// USB specific stuff
	d.SetConnectionType(CONNECTION_TYPE_USB_SERIAL)
	d.SetTimeSerial(recievedUnixSec)

	d.RecalculateLastResetUnix()

	return d, nil
}

type DuetTypeInfo struct {
	ExpectedBytes        int
	ExpectedStringLen    int
	StructInstanceGetter func() DuetData
	TypeAlias            string
}

func (typeInfo DuetTypeInfo) checkByteLen(byteLen int) error {
	if byteLen < typeInfo.ExpectedBytes {
		return fmt.Errorf("exepcted at least %d bytes for sample, only got %d", typeInfo.ExpectedBytes, byteLen)
	}
	return nil
}
func (typeInfo DuetTypeInfo) checkSubstringLen(n int) error {
	if n != typeInfo.ExpectedStringLen {
		return fmt.Errorf("expected a list of values at least %d in length, only got %d", typeInfo.ExpectedStringLen, n)
	}
	return nil
}

func WriteDuetDataToDir(d DuetData, dir string) error {
	for _, m := range d.SensorMeasurements() {
		if err := StoreSensorData(m, dir); err != nil {
			return err
		}
	}
	return nil
}
