package spider

type Spiders []*Spider

var SpiderList = Spiders{}

func (Spiders) Init() {
	SpiderList = Spiders{}
}

func (Spiders) Add(sp *Spider) {
	SpiderList = append(SpiderList, sp)
}
