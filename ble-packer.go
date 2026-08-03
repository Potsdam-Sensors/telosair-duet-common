package telosairduetcommon

import "strings"

// BLEPacker defines the contract any variant struct must satisfy for BLE export.
type BLEPacker interface {
	GetRequiredSensorFiles() []string
	PopulateFromLocalFiles(data map[string]string)
	ToBLEBytes() []byte
}

// GetBLEPacker returns the matching struct instance for a given variant string.
func GetBLEPacker(deviceType string) BLEPacker {
	cleanedType := strings.TrimSpace(deviceType)

	switch cleanedType {
	case "4.7", "Mk4.7":
		return &DuetDataMk4Var7{}
	// As you add new variants, simply register them here:
	// case "4.8", "Mk4.8":
	// 	return &DuetDataMk4Var8{}
	default:
		return nil
	}
}
