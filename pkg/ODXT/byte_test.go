package ODXT

import (
	"os"
	"testing"
)

func TestReadKeys(t *testing.T) {
	filename := "../../cmd/ODXT/configs/keys.txt"

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatalf("keys.txt file does not exist at path: %s", filename)
	}

	// 将文件内容打印出来
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read keys.txt file: %s", err)
	}
	t.Logf("File content: %s", string(content))

	keys := ReadKeys(filename)
	for i := 0; i < 4; i++ {
		t.Logf("Key %d: %x", i, keys[i])
		t.Logf("Length of key %d: %d", i, len(keys[i]))
	}
}
