package filter

import (
	"fmt"
	"sync"
)

// DataContainer 用于储存数据的容器
type DataContainer interface {
	// Set 往容器中存储数据，仅Producer声明产出的字段可以存入容器
	Set(l Producer, fieldID FieldID, data interface{}, err error)
	// Get 从容器中获取数据，仅Consumer声明依赖的字段可以从容器中获取
	Get(l Consumer, fieldID FieldID) (interface{}, error)
}

type dataContainer struct {
	data sync.Map
}

func NewDataContainer() DataContainer {
	return &dataContainer{sync.Map{}}
}

func (d *dataContainer) Set(l Producer, fieldID FieldID, data interface{}, err error) {
	// 仅Producer声明产出的字段可以存入容器
	output := false
	for id := range l.ProduceFields() {
		if id == fieldID {
			output = true
			break
		}
	}
	if !output {
		panic(fmt.Errorf("the producer do not output filed:%s", fieldID))
	}
	// 字段索引必须唯一
	_, loaded := d.data.LoadOrStore(fieldID, dataHolder{data, err})
	if loaded {
		panic(fmt.Errorf("filedID:%s is duplicated", fieldID))
	}
}

func (d *dataContainer) Get(l Consumer, fieldID FieldID) (interface{}, error) {
	// 仅Consumer声明依赖的字段可以从容器中获取
	isDependent := false
	for handlerID := range l.ConsumeFields() {
		if handlerID == fieldID {
			isDependent = true
			break
		}
	}
	if !isDependent {
		panic(fmt.Errorf("the consumer do not depend on field:%s", fieldID))
	}
	val, ok := d.data.Load(fieldID)
	if !ok || val == nil {
		panic(fmt.Errorf("dependent data is not exist"))
	}
	data := val.(dataHolder)
	return data.data, data.err
}

type dataHolder struct {
	data interface{}
	err  error
}
