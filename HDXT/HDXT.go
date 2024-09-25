package HDXT

import (
	"ConjunctiveSSE/utils"
	"crypto/rand"
	"database/sql"
	"log"
)

type Mitra struct {
	Key     []byte
	FileCnt map[string]int
	EDB     *sql.DB
}

type Auhme struct {
	Keys  [3][]byte
	Cnt   int
	S     []string
	T     map[string]int
	Delta int
}

type HDXT struct {
	Mitra
	Auhme
}

func (hdxt *HDXT) Setup(dbName string, randomKey bool) error {
	// 初始化私钥
	if randomKey {
		// 生成4个16字节长度的随机私钥
		keyLen := 16
		hdxt.Mitra.Key = make([]byte, keyLen)
		if _, err := rand.Read(hdxt.Mitra.Key); err != nil {
			log.Println("Error generating random key:", err)
			return err
		}
		for i := 0; i < 3; i++ {
			key := make([]byte, keyLen)
			if _, err := rand.Read(key); err != nil {
				log.Println("Error generating random key:", err)
				return err
			}
			hdxt.Auhme.Keys[i] = key
		}
	} else {
		// 读取私钥
		var err error
		hdxt.Mitra.Key, hdxt.Auhme.Keys, err = utils.HdxtReadKeys("./benchmark/HDXT/keys.txt")
		if err != nil {
			log.Println("Error reading keys:", err)
			return err
		}
	}

	return nil
}
