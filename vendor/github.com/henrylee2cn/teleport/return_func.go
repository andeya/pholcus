package teleport

// ***********************************************常用函数*************************************************** \\
// API中生成返回结果的方法
// OpAndToAndFrom[0]参数为空时，系统将指定与对端相同的操作符
// OpAndToAndFrom[1]参数为空时，系统将指定与对端为接收者
// OpAndToAndFrom[2]参数为空时，系统将指定自身为发送者
func ReturnData(body interface{}, OpAndToAndFrom ...string) *NetData {
	data := &NetData{
		Status: SUCCESS,
		Body:   body,
	}
	if len(OpAndToAndFrom) > 0 {
		data.Operation = OpAndToAndFrom[0]
	}
	if len(OpAndToAndFrom) > 1 {
		data.To = OpAndToAndFrom[1]
	}
	if len(OpAndToAndFrom) > 2 {
		data.From = OpAndToAndFrom[2]
	}
	return data
}

// 返回错误，receive建议为直接接收到的*NetData
func ReturnError(receive *NetData, status int, msg string, nodeuid ...string) *NetData {
	receive.Status = status
	receive.Body = msg
	receive.From = ""
	if len(nodeuid) > 0 {
		receive.To = nodeuid[0]
	} else {
		receive.To = ""
	}
	return receive
}
