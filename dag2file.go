package merkledag

// Hash to file

import (
	"encoding/json"
	"strings"
)

// Hash to file
func Hash2File(store KVStore, hash []byte, path string, hp HashPool) []byte {
	// 根据hash和path， 返回对应的文件, hash对应的类型是tree
	//获取hash对应数据
	jsonObj, err := store.Get(hash)
	if err != nil {
		panic("没有对应hash数据")
	}

	//转换json数据
	var obj Object
	json.Unmarshal(jsonObj, &obj)

	//查找数据
	var firstName, last string
	if obj.Links == nil { //是一个file
		return obj.Data
	} else { //是一个tree，依据path查找
		names := strings.Split(path, "/")
		if len(names) > 1 { //是一串路径,name[0]是根目录名
			firstName = names[1]
			last = strings.Join(names[1:], "/")
		} else { //只是一个文件名
			firstName = names[0]
			last = ""
		}
		for _, link := range obj.Links {
			if link.Name == firstName {
				return Hash2File(store, link.Hash, last, hp)
			}
		}
	}
	return nil
}
