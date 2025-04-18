package util

import (
	"runtime"
	"strings"
)

func GetCurrentFuncName() string {
	pc, _, _, _ := runtime.Caller(1)             // 获取调用者的信息
	fullFuncName := runtime.FuncForPC(pc).Name() // 获取完整的函数名
	// 提取方法名（去掉包名）
	funcName := fullFuncName[strings.LastIndex(fullFuncName, ".")+1:]
	return funcName
}
