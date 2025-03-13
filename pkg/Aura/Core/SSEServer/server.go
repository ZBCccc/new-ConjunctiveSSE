package sseserver

import (
	"crypto/aes"
	"encoding/base64"
	"fmt"
	ggmtree "github.com/ZBCccc/Aura/GGM"
	util "github.com/ZBCccc/Aura/Util"
	"log"
	"slices"

	bloom "github.com/ZBCccc/Aura/bloom"
)

type Server struct {
	Tags map[string]string
	Dict map[string][]string
	Keys map[int][]byte
}

var iv = []byte("0123456789123456")

func NewServer() *Server {
	return &Server{
		Tags: make(map[string]string),
		Dict: make(map[string][]string),
		Keys: make(map[int][]byte),
	}
}

// AddEntries adds the entries to the server.
func (s *Server) AddEntries(label string, tag string, ciphertexts []string) {
	s.Tags[label] = tag
	s.Dict[label] = ciphertexts
}

func (s *Server) Search(token []byte, nodeList []ggmtree.GGMNode, level int) []string {
	s.Keys = make(map[int][]byte)
	s.computeLeafKeys(nodeList, level)
	counter := 0
	var resList []string

	for {
		// get label string
		label := util.HmacDigest([]byte(fmt.Sprintf("%d", counter)), token)
		counter++
		labelStr := base64.StdEncoding.EncodeToString(label)

		// terminate if the label is not found
		if _, exists := s.Tags[labelStr]; !exists {
			break
		}

		// get the search position
		bf := bloom.New(util.GGMSize, util.HashSize)
		tagBytes, err := base64.StdEncoding.DecodeString(s.Tags[labelStr])
		if err != nil {
			log.Fatal("Failed to decode tag", err)
		}
		searchPos := bf.GetIndex(tagBytes)
		slices.Sort(searchPos)

		// derive the key from the search position and decrypt the id
		ciphertextList := s.Dict[labelStr]
		for i := 0; i < min(len(searchPos), len(ciphertextList)); i++ {
			if s.Keys[int(searchPos[i])] == nil {
				break
			}

			ciphertext, err := base64.StdEncoding.DecodeString(ciphertextList[i])
			if err != nil {
				log.Fatal("Failed to decode ciphertext", err)
			}
			res, err := util.AesDecrypt(ciphertext, s.Keys[int(searchPos[i])], iv)
			if err != nil {
				log.Fatal("Failed to AesDecrypt", err)
			}
			resList = append(resList, string(res))
		}
	}

	return resList
}

func (s *Server) computeLeafKeys(nodeList []ggmtree.GGMNode, level int) {
	for _, node := range nodeList {
		for i := 0; i < 1<<(level-node.Level); i++ {
			offset := (node.Index << (level - node.Level)) + i
			deriveKey := make([]byte, aes.BlockSize)
			copy(deriveKey, node.Key[:])
			ggmtree.DeriveKeyFromTree(deriveKey, uint(offset), level-node.Level, 0)

			if _, exists := s.Keys[offset]; !exists {
				s.Keys[offset] = make([]byte, aes.BlockSize)
				copy(s.Keys[offset], deriveKey)
			}
		}
	}
}
