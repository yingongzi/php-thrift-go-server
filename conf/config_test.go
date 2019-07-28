package conf

import "testing"

//此处涉及到conf文件目录位置问题，main方法调用和test方法调用路径是不一致的，首先 保全main包
func TestLoadConfigFile(t *testing.T) {
	LoadConfigFile()
}
