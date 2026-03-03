package teleport

const (
	SUCCESS = 0
	FAILURE = -1
	LLLEGAL = -2
)

// NetData is the data transfer structure.
type NetData struct {
	Body      interface{}
	Operation string
	From      string
	To        string
	Status    int
	Flag      string
}

// NewNetData 创建网络数据传输结构。
func NewNetData(from, to, operation string, flag string, body interface{}) *NetData {
	return &NetData{
		From:      from,
		To:        to,
		Body:      body,
		Operation: operation,
		Status:    SUCCESS,
		Flag:      flag,
	}
}
