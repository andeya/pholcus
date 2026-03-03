package teleport

// ReturnData 构建 API 响应。OpAndToAndFrom[0] 为空则沿用对端操作；[1] 为空则对端为接收方；[2] 为空则自身为发送方。
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

// ReturnError returns an error response; receive should be the original *NetData.
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
