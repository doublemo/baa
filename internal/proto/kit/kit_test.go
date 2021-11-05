package kit

import (
	"testing"

	coresproto "github.com/doublemo/baa/cores/proto"
)

func TestMakeErrCode(t *testing.T) {
	t.Log(MakeErrCode(SM, 500))
	for i := -32767; i < 32767; i++ {
		for j := 1; j < 4096; j++ {
			for k := 0; k < 4; k++ {
				code := MakeErrCode(coresproto.Command(i), uint32(j), uint32(k))
				pi, pj, pk := ParseErrCode(code)
				if pi != int32(i) || pj != uint32(j) || pk != uint32(k) {
					t.Fatalf("code:%d, i = %d, j = %d, k = %d  != pi =%d pj = %d pk != %d", code, i, j, k, pi, pj, pk)
				}
			}
		}
	}

}
