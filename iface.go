package filter

import "context"

// Filter 数据过滤器
type Filter interface {
	Consumer
	ID() string                                                                // 唯一id
	Filter(ctx context.Context, req interface{}, container DataContainer) bool // 过滤逻辑，1.从数据容器中获取依赖数据，2.执行过滤
}

// Loader 数据加载器
type Loader interface {
	Consumer
	Producer
	ID() string                                                         // 唯一id
	Load(ctx context.Context, req interface{}, container DataContainer) // 加载逻辑，1.从容器中获取依赖字段，2.加载数据，3.保存数据至容器
}

type Producer interface {
	ProduceFields() StringSet // 生成的字段
}

type Consumer interface {
	ConsumeFields() StringSet // 消费的字段
}
