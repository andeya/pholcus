package queue

type Queue struct {
	PoolSize int
	PoolChan chan interface{}
}

func NewQueue(size int) *Queue {
	return &Queue{
		PoolSize: size,
		PoolChan: make(chan interface{}, size),
	}
}

func (this *Queue) Init(size int) *Queue {
	this.PoolSize = size
	this.PoolChan = make(chan interface{}, size)
	return this
}

func (this *Queue) Push(i interface{}) bool {
	if len(this.PoolChan) == this.PoolSize {
		return false
	}
	this.PoolChan <- i
	return true
}

func (this *Queue) PushSlice(s []interface{}) {
	for _, i := range s {
		this.Push(i)
	}
}

func (this *Queue) Pull() interface{} {
	return <-this.PoolChan
}

// 二次使用Queue实例时，根据容量需求进行高效转换
func (this *Queue) Exchange(num int) (add int) {
	last := len(this.PoolChan)

	if last >= num {
		add = int(0)
		return
	}

	if this.PoolSize < num {
		pool := []interface{}{}
		for i := 0; i < last; i++ {
			pool = append(pool, <-this.PoolChan)
		}
		// 重新定义、赋值
		this.Init(num).PushSlice(pool)
	}

	add = num - last
	return
}
