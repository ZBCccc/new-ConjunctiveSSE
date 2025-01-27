package sdssecqClient

import (
	"testing"

	util "github.com/ZBCccc/Aura/Util"
)

func TestClient(t *testing.T) {
	client := NewClient()
	// insert
	// client.Update(utils.Add, "test", "1")
	// client.Update(utils.Add, "test", "2")
	client.Update(util.Insert, "hello", "1")
	client.Update(util.Insert, "hello", "2")
	// client.Update(util.Insert, "hello", "3")

	client.Update(util.Insert, "world", "1")
	client.Update(util.Insert, "world", "2")
	client.Update(util.Insert, "world", "3")
	// delete
	// client.Update(util.Delete, "test", "1")	// 2 12 12
	client.Update(util.Delete, "hello", "2") // 1 1 12
	// client.Update(util.Delete, "world", "1") // 2 2 2

	// search
	// res := client.Search([]string{"test", "hello", "world"})
	// t.Log("res", res, "len", len(res))
	res, _, _ := client.Search([]string{"hello", "world"})
	t.Log("res", res, "len", len(res))
	res, _, _ = client.Search([]string{"hello"})
	t.Log("res", res, "len", len(res))
}
