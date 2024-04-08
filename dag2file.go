package merkledag

import (
	"encoding/json"
	"strings"
)
// Hash to file
func Hash2File(store KVStore, hash []byte, path string, hp HashPool) []byte {
	// ����hash��path�� ���ض�Ӧ���ļ�, hash��Ӧ��������tree
	flag,_ := store.Has(hash); //����has�����鿴�洢���Ƿ����ָ���Ķ���
	if flag {
		objBinary,_ := store.Get(hash); //��ȡ����������
		obj := binaryToObj(objBinary); //��������������
        pathArr:=strings.Split(path,"\\");//��·���ַ������շָ����ָ������pathArr
		cur := 1
		//��������õ��Ķ���·�����顢��ǰ·�������ʹ洢����,�ú��������ָ���ļ��������ļ�����
		return getFileByDir(obj,pathArr,cur,store)
	}
	return nil
}

func getFileByDir(obj* Object,pathArr []string,cur int,store KVStore) []byte {
	//�жϵ�ǰ������·�������Ƿ񳬳���·������ĳ��ȣ���˵���Ѿ�������·����ֱ�ӷ��� nil
	if cur >= len(path) {
		return nil
	}
	index := 0
	//������ǰĿ¼�����е���������
	for i := range obj.Links {
		//�� obj.Data �л�ȡ��ǰ���ӵ����ͣ�������ת��Ϊ�ַ�����index �� index+STEP �ǵ�ǰ���������� obj.Data �е�������Χ
		objType := string(obj.Data[index : index+STEP])
		index += STEP
		//��ȡ��ǰ���ӵ���Ϣ
		objInfo := obj.Links[i]
		//�����ǰ���ӵ�������·�������е�ǰ������Ӧ��·����ƥ�䣬��������ǰ���ӣ�������һ�ε���
		if objInfo.Name != pathArr[cur] {
			continue
		}
		switch objType {
		case TREE:
			objDirBinary, _ := store.Get(objInfo.Hash)
			objDir := binaryToObj(objDirBinary)
			ans := getFileByDir(objDir, pathArr, cur+1, store)
			if ans != nil {
				return ans
			}
		//�ļ�����
		case BLOB:
			ans, _ := store.Get(objInfo.Hash)
			return ans
		//�б�����
		case LIST:
			objLinkBinary, _ := store.Get(objInfo.Hash)
			objList := binaryToObj(objLinkBinary)
			ans := getFileByList(objList, store)
			return ans
		}
	}
	return nil
}
//�б������еݹ�����ļ������������ҵ����ļ�����ƴ�ӳ�һ�����[]byte ��Ƭ����
func getFileByList(obj *Object, store KVStore) []byte {
	ans := make([]byte, 0)
	index := 0
	for i := range obj.Links {
		curObjType := string(obj.Data[index : index+STEP])
		index += STEP
		curObjLink := obj.Links[i]
		curObjBinary, _ := store.Get(curObjLink.Hash)
		curObj := binaryToObj(curObjBinary)
		if curObjType == BLOB {
			ans = append(ans, curObjBinary...)
		} else { //List
			tmp := getFileByList(curObj, store)
			ans = append(ans, tmp...)
		}
	}
	return ans
}
//�����������ݽ����ɶ���
func binaryToObj(objBinary []byte) *Object {
	var res Object
	//ʹ�� json.Unmarshal ���������������� objBinary ������ Object �ṹ�壬���洢�� res ������
	json.Unmarshal(objBinary, &res)
	return &res  //���ص�ַ
}

