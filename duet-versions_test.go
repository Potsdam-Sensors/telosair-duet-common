package telosairduetcommon

import (
	"testing"
)

func TestDuetsImplementDuetData(t *testing.T) {
	// Compile-time checks:
	for _, _ = range []DuetData{
		&DuetDataMk4Var0{}, &DuetDataMk4Var1{}, &DuetDataMk4Var2{}, &DuetDataMk4Var3{},
		&DuetDataMk4Var4{}, &DuetDataMk4Var5{}, &DuetDataMk4Var6{}, &DuetDataMk4Var7{},
	} {
	}

}
