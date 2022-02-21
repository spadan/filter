package test

import (
	"context"
	"log"
	"school/filter"
)

const (
	FilterName = "name_filter"
)

type nameFilter struct {
}

func (n *nameFilter) ID() string {
	return FilterName
}

func (n *nameFilter) DependentFields() filter.StringSet {
	return filter.NewStringSet(FieldUserBase)
}

func (n *nameFilter) DoFilter(ctx context.Context, req interface{}, container filter.DataContainer) bool {
	log.Printf("filter:%s,go:%v", FilterName, ctx.Value("id"))
	data, err := container.GetDependentData(n, FieldUserBase)
	if err != nil {
		log.Printf("get user name fail,err:%v", err)
		return false
	}
	return data.(UserBase).Name != "xx"
}
