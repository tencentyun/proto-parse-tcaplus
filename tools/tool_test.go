package tools

import (
	"testing"
)

func TestSnakeCase(t *testing.T) {

	ret := SnakeCase("AbCd")
	t.Log(ret)
	ret = SnakeCase("OUT_ChatBattles")
	t.Log(ret)
}
