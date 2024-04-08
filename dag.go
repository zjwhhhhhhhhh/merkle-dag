package merkledag

import (
	"hash"
	"errors"
	"encoding/json"
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

func Add(store KVStore, node Node, h hash.Hash) []byte {
	//将分片写到KVstore中
	//判断数据类型
	switch n := node.(type) {
	case FILE:
		file := node.(File)  //将节点断言为 File 类型
        tmp := StoreFile(store, file, h)  //调用 StoreFile 函数将文件存储到 KVStore 中，得到存储结果 tmp
        jsonMarshal, _ := json.Marshal(tmp) // tmp 转换为 JSON 格式的字节切片
        hash := calculateHash(jsonMarshal, h)  // 计算存储结果的哈希值
        return hash

	case DIR:
		dir := node.(Dir)
        tmp := StoreDir(store, dir, h)
        jsonMarshal, _ := json.Marshal(tmp)
        hash := calculateHash(jsonMarshal, h)
        return hash

	default:
        // 如果节点类型未知，抛出一个错误
        panic("unknown node type")
	}
}


//接收并处理一个字节切片
func calculateHash(data []byte, h hash.Hash) []byte {
    h.Reset()  //将哈希函数对象重置为初始状态
    h.Write(data)  //将数据写入哈希函数
    hash := h.Sum(nil)  //调用 Sum 方法，返回计算得到的哈希值。参数 nil 表示不需要附加任何数据，而是要获取当前状态下的哈希值
    return hash  //返回计算得到的哈希值
}


//存文件的方法
func StoreFile(store KVStore, file File, h hash.Hash) (*Object,error) {
	//1.获取文件数据
	data := file.Bytes() //获取数据
	//如果小于256KB，则存储为blob类型
	if len(data) <= 256*1024{
		blob := Object{Data:data, Links:nil}
		jsonData,err := json.Marshal(blob)  //将 blob 对象序列化为 JSON 格式的字节切片
		if err != nil {
            return nil, err
        }
		hash := calculateHash(jsonData,h)    //计算切片哈希值
		err = store.Put(hash, jsonData) // 假设 Put 方法接收哈希和数据，存储数据
        if err != nil {
            return nil, err
        }
        return &blob, nil   //返回指向blob对象的指针和nil错误表示存储成功
    }
	//如果大于256KB，则存储为Link
	var links []Link  //声明一个名为links的Link类型的切片,存储每个数据块的信息
	for i := 0; i < len(data); i += 256*1024 {  //遍历数据长度，每次增加256KB
		end := i + 256*1024    //数据块结束位置
		if end > len(data) {   //防止越界
			end = len(data)
		}
		chunk := data[i:end]   //结算开始和结束位置，切割出一个数据块
		blob := Object{Data: chunk, Links: nil}  
		jsonData, err := json.Marshal(blob)  //序列化为json字节
		if err != nil {
			return nil, err
		}
		hash := calculateHash(jsonData, h)
		if err := store.Put(hash, jsonData); err != nil {
			return nil, err
		}
		//将当前数据块的相关信息（文件名、哈希值、大小）添加到links切片中。
		links = append(links, Link{Name: file.Name(), Hash: hash, Size: len(chunk)})
	}
	//循环结束后,调用storeLinks函数,将存储的链接传递给它并返回其结果
	return StoreLinks(store,links,h)
}

func StoreDir(store KVStore, dir Dir, h hash.Hash) (*Object,error) {
	iter := dir.It()  //调用了dir目录的It方法，获取目录迭代器
	links := make([]Link, 0)   //声明一个空的 Link 类型切片 links
	//迭代器遍历每个子节点,Next方法会将迭代器指向下一个节点
	for iter.Next() {  
		elem := iter.Node()  //调用迭代器Node方法，获取每个节点信息
		var hash []byte  //声明一个字节切片hash，用于存储当前节点的哈希值
		var err error
		if elem.Type() == FILE {
			hash, err = StoreFile(store, elem.(File), h)  //递归调用 Add 函数存储子节点数据
		} else if elem.Type() == DIR {
			hash, err = StoreDir(store, elem.(Dir), h)  //递归调用 Add 函数存储子节点数据
		} else {
			return nil, errors.New("unknown node type")
		}
		if err != nil {
			return nil, err
		}
		//当前节点的相关信息名称、哈希值、大小,添加到links切片中
		links = append(links, Link{Name: elem.Name(), Hash: hash, Size: int(elem.Size())})  
	}
	return StoreLinks(store, links, h)
}

// StoreLinks 函数用于存储链接节点到 Merkle DAG 中
func StoreLinks(store KVStore, links []Link, h hash.Hash) (*Object,error) {
	obj := Object{Links: links, Data: nil}
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	hash := calculateHash(jsonData, h)
	if err := store.Put(hash, jsonData); err != nil {
		return nil, err
	}
	return &obj, nil
}
