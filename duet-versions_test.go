package telosairduetcommon

import (
	"fmt"
	"testing"
)

/*
This test just checks that all the DuetData types implement the DuetData interface.
This does not even really need to run as the compiler will check this at compile time.
*/
func TestDuetsImplementDuetData(t *testing.T) {
	// Compile-time checks:
	for _, _ = range []DuetData{
		&DuetDataMk1Var0{}, &DuetDataMk1Var2{},
		&DuetDataMk3Var1{},
		&DuetDataMk4Var0{}, &DuetDataMk4Var1{}, &DuetDataMk4Var2{}, &DuetDataMk4Var3{},
		&DuetDataMk4Var4{}, &DuetDataMk4Var5{}, &DuetDataMk4Var6{}, &DuetDataMk4Var7{},
		&DuetDataMk4Var8{}, &DuetDataMk4Var9{}, &DuetDataMk4Var10{}, &DuetDataMk4Var12{},
	} {
	}

}

/*
This test checks that the `getTypeInfo` function returns the correct DuetTypeInfo
for the given hardware version and sensor variation.
It also checks that it returns `nil` for invalid combinations.
*/
func TestGetTypeInfo(t *testing.T) {
	// Test cases that should all return an actual Duet Type

	/* MK4 */
	for varNum, duetTypeInstance := range []*DuetTypeInfo{
		&DuetTypeMk4Var0, &DuetTypeMk4Var1, &DuetTypeMk4Var2, &DuetTypeMk4Var3,
		&DuetTypeMk4Var4, &DuetTypeMk4Var5, &DuetTypeMk4Var6, &DuetTypeMk4Var7,
		&DuetTypeMk4Var8,
	} {
		resultDuetType := getTypeInfo(4, uint8(varNum))
		if resultDuetType == nil {
			t.Errorf("nil for `getTypeInfo(4, %d)", uint8(varNum))
		}

		if resultDuetType != duetTypeInstance {
			t.Errorf("expected to get duet type `%s` for Mk4.%d, got `%s` instead", duetTypeInstance.TypeAlias, uint8(varNum), resultDuetType.TypeAlias)
		}
	}

	/* MK1 */
	for varNum, duetTypeInstance := range []*DuetTypeInfo{
		&DuetTypeMk1Var0, nil, &DuetTypeMk1Var2,
	} {
		if duetTypeInstance == nil { // Just because Mk1.1 is not implemented, we skip it
			continue
		}
		resultDuetType := getTypeInfo(1, uint8(varNum))
		if resultDuetType == nil {
			t.Errorf("nil for `getTypeInfo(4, %d)", uint8(varNum))
		}

		if resultDuetType != duetTypeInstance {
			t.Errorf("expected to get duet type `%s` for Mk4.%d, got `%s` instead", duetTypeInstance.TypeAlias, uint8(varNum), resultDuetType.TypeAlias)
		}
	}

	/* MK3 */
	for varNum, duetTypeInstatnce := range []*DuetTypeInfo{
		nil, &DuetTypeMk3Var1,
	} {
		if duetTypeInstatnce == nil { // Just because Mk3.0 is not implemented, we skip it
			continue
		}
		resultDuetType := getTypeInfo(3, uint8(varNum))
		if resultDuetType == nil {
			t.Errorf("nil for `getTypeInfo(3, %d)", uint8(varNum))
		}
		if resultDuetType != duetTypeInstatnce {
			t.Errorf("expected to get duet type `%s` for Mk3.%d, got `%s` instead", duetTypeInstatnce.TypeAlias, uint8(varNum), resultDuetType.TypeAlias)
		}
	}

	// Test cases that should be invalid and return nil
	for _, testData := range [][2]uint8{
		{0, 1},
		{100, 12},
		{0, 0},
		{4, 100},
		{3, 7},
	} {
		if duetTypeResult := getTypeInfo(testData[0], testData[1]); duetTypeResult != nil {
			t.Errorf("expected `nil` for type Mk%d.%d, got `%s` instead", testData[0], testData[1], duetTypeResult.TypeAlias)
		}
	}
}

func compareTypes(a, b interface{}) bool {
	return fmt.Sprintf("%T", a) == fmt.Sprintf("%T", b)
}
func TestDuetTypeMethods(t *testing.T) {
	type testData struct {
		t *DuetTypeInfo
		a string
		d DuetData
	}
	testDuetType := func(td testData) error {
		duetTypeInstance := td.t
		expectedTypeAlias := td.a
		expectedDuetDataTypeInstance := td.d

		// Check that the TypeAlias function works as expected
		if resultingTypeAlias := duetTypeInstance.TypeAlias; resultingTypeAlias != expectedTypeAlias {
			return fmt.Errorf("expected type alias given to be `%s`, but got `%s`", expectedTypeAlias, resultingTypeAlias)
		}

		// Check that the struct instance getter works as expected
		if resultingDataStructInstance := duetTypeInstance.StructInstanceGetter(); !compareTypes(expectedDuetDataTypeInstance, resultingDataStructInstance) {
			return fmt.Errorf("expected data type from getter to be '%T', but got '%T' instead", expectedDuetDataTypeInstance, resultingDataStructInstance)
		}
		return nil
	}

	for _, testData := range []testData{
		{&DuetTypeMk1Var0, "Mk1.0", &DuetDataMk1Var0{}},
		{&DuetTypeMk1Var2, "Mk1.2", &DuetDataMk1Var2{}},
		{&DuetTypeMk3Var1, "Mk3.1", &DuetDataMk3Var1{}},
		{&DuetTypeMk4Var0, "Mk4.0", &DuetDataMk4Var0{}},
		{&DuetTypeMk4Var1, "Mk4.1", &DuetDataMk4Var1{}},
		{&DuetTypeMk4Var2, "Mk4.2", &DuetDataMk4Var2{}},
		{&DuetTypeMk4Var3, "Mk4.3", &DuetDataMk4Var3{}},
		{&DuetTypeMk4Var4, "Mk4.4", &DuetDataMk4Var4{}},
		{&DuetTypeMk4Var5, "Mk4.5", &DuetDataMk4Var5{}},
		{&DuetTypeMk4Var6, "Mk4.6", &DuetDataMk4Var6{}},
		{&DuetTypeMk4Var7, "Mk4.7", &DuetDataMk4Var7{}},
		{&DuetTypeMk4Var8, "Mk4.8", &DuetDataMk4Var8{}},
		{&DuetTypeMk4Var9, "Mk4.9", &DuetDataMk4Var9{}},
		{&DuetTypeMk4Var10, "Mk4.10", &DuetDataMk4Var10{}},
		{&DuetTypeMk4Var12, "Mk4.12", &DuetDataMk4Var12{}},
	} {
		if err := testDuetType(testData); err != nil {
			t.Error(err)
		}
	}
}

func TestDuetMethodsSimple(t *testing.T) {
	type TestData struct {
		duetDataInstance      DuetData
		expectedDuetTypeFloat float32
	}

	for _, testData := range []TestData{
		{&DuetDataMk1Var0{}, 1.0},
		{&DuetDataMk1Var2{}, 1.2},
		{&DuetDataMk3Var1{}, 3.1},
		{&DuetDataMk4Var0{}, 4.0},
		{&DuetDataMk4Var1{}, 4.1},
		{&DuetDataMk4Var2{}, 4.2},
		{&DuetDataMk4Var3{}, 4.3},
		{&DuetDataMk4Var4{}, 4.4},
		{&DuetDataMk4Var5{}, 4.5},
		{&DuetDataMk4Var6{}, 4.6},
		{&DuetDataMk4Var7{}, 4.7},
		{&DuetDataMk4Var8{}, 4.8},
	} {
		data := testData.duetDataInstance

		/* ToMap() Type Integrity */
		expectedType := testData.expectedDuetTypeFloat
		duetDataMap := data.ToMap("abc")

		// Make sure key/value pair exists
		duetTypeUntyped, ok := duetDataMap[KEY_DEVICE_TYPE]
		if !ok {
			t.Errorf("duet type %.1f's map from `ToMap()` has no field '%s'", expectedType, KEY_DEVICE_TYPE)
		}

		// Make sure duet type is a float (tolerant of size for ease of testing)
		var duetTypeFloat float32
		switch duetType := duetTypeUntyped.(type) {
		case float32:
			duetTypeFloat = duetType
		case float64:
			duetTypeFloat = float32(duetType)
		default:
			t.Errorf("duet type %.1f from map, using key '%s', gives a variable of type '%s', not the expected float",
				expectedType, KEY_DEVICE_TYPE, duetType)
		}

		// Check that duet type matches
		if duetTypeFloat != expectedType {
			t.Errorf("duet type was expected to be %.1f, got %.1f instead", expectedType, duetTypeFloat)
		}
	}
}
