package merkledag

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
)

type Link struct {
	Name string
	Hash []byte
	Size int
}

type Object struct {
	Links []Link
	Data  []byte
}

func (o *Object) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, int32(len(o.Links))); err != nil {
		return nil, err
	}

	for _, link := range o.Links {

		if err := binary.Write(&buf, binary.BigEndian, int32(len(link.Name))); err != nil {
			return nil, err
		}

		if _, err := buf.WriteString(link.Name); err != nil {
			return nil, err
		}

		if err := binary.Write(&buf, binary.BigEndian, int32(len(link.Hash))); err != nil {
			return nil, err
		}
		if _, err := buf.Write(link.Hash); err != nil {
			return nil, err
		}

		if err := binary.Write(&buf, binary.BigEndian, int32(link.Size)); err != nil {
			return nil, err
		}
	}

	if err := binary.Write(&buf, binary.BigEndian, int32(len(o.Data))); err != nil {
		return nil, err
	}

	if _, err := buf.Write(o.Data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func hashData(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

func combineHashes(h hash.Hash, hash1, hash2 []byte) []byte {
	h.Reset()
	h.Write(hash1)
	h.Write(hash2)
	return h.Sum(nil)
}

// 构建 Merkle Tree 并返回根哈希值
func buildMerkleTree(hashes [][]byte, h hash.Hash) []byte {
	if len(hashes) == 0 {
		return nil
	}

	// 如果叶子节点数量不是2的幂，添加空哈希值作为占位符
	for len(hashes) > 1 && len(hashes)%2 != 0 {
		hashes = append(hashes, nil)
	}

	for len(hashes) > 1 {
		var newLevel [][]byte
		for i := 0; i < len(hashes); i += 2 {
			var left, right []byte
			if i+1 < len(hashes) {
				right = hashes[i+1]
			}
			combined := combineHashes(h, left, right)
			newLevel = append(newLevel, combined)
		}
		hashes = newLevel
	}

	// 返回 Merkle Root
	return hashes[0]
}

func Add(store KVStore, node Node, h hash.Hash) []byte {
	var nodeHashes [][]byte

	// 为节点及其所有子节点计算哈希值
	rootHash, _ := saveNode(store, node)

	nodeHashes = append(nodeHashes, rootHash)

	// 构建 Merkle Tree 并计算 Merkle Root
	merkleRoot := buildMerkleTree(nodeHashes, h)

	return merkleRoot
}

// 保存节点并获取其哈希值
func saveNode(store KVStore, n Node) ([]byte, error) {
	switch node := n.(type) {
	case File:
		return saveFile(store, node)
	case Dir:
		return saveDir(store, node)
	default:
		return nil, fmt.Errorf("unknown node type: %T", n)
	}
}

// saveFile 保存文件节点并返回其哈希值
func saveFile(store KVStore, file File) ([]byte, error) {
	data := file.Data()
	hash := hashData(data)
	err := store.Put(hash, data)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// saveDir 保存目录节点并返回其哈希值
func saveDir(store KVStore, dir Dir) ([]byte, error) {
	var links []Link
	iter := dir.It()
	defer iter.Close()

	for iter.Next() {
		child := iter.Node()
		childHash, err := saveNode(store, child)
		if err != nil {
			return nil, err
		}
		links = append(links, Link{Name: child.Name(), Hash: childHash, Size: len(childHash)})
	}

	obj := Object{Links: links}
	data, err := obj.MarshalBinary()
	if err != nil {
		return nil, err
	}
	hash := hashData(data)
	hashStr := string(hash)
	err = store.Put([]byte(hashStr), data)
	if err != nil {
		return nil, err
	}
	return hash, nil
}
