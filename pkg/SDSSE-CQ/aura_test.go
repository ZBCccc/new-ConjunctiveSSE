package sdssecq

import (
	"testing"

	sseclient "github.com/ZBCccc/Aura/Core/SSEClient"
	util "github.com/ZBCccc/Aura/Util"
)

func TestAura(t *testing.T) {
	sseClient := sseclient.NewSSEClient()

	// insert
	// sseClient.Update(util.Insert, "test", "1")
	// sseClient.Update(util.Insert, "test", "2")
	// sseClient.Update(util.Insert, "hello", "1")
	// sseClient.Update(util.Insert, "hello", "2")
	sseClient.Update(util.Insert, "world", "1")
	sseClient.Update(util.Insert, "world", "2")

	// delete
	// sseClient.Update(util.Delete, "test", "1")
	// sseClient.Update(util.Delete, "test", "2")
	// sseClient.Update(util.Delete, "hello", "1")
	// sseClient.Update(util.Delete, "hello", "2")
	sseClient.Update(util.Delete, "world", "1")
	// sseClient.Update(util.Delete, "world", "2")

	// search
	// res := sseClient.Search("test")
	// t.Log("res", res, "len", len(res))
	// res = sseClient.Search("hello")
	// t.Log("res", res, "len", len(res))
	res := sseClient.Search("world")
	t.Log("res", res, "len", len(res))
}
