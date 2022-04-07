package test

import (
	"context"
	"log"
	"school/filter"
)

const (
	FilterAge = "age_filter"
)

type ageFilter struct {
}

func (n *ageFilter) ID() string {
	return FilterAge
}

func (n *ageFilter) ConsumeFields() filter.StringSet {
	return filter.NewStringSet(FieldUserBase)
}

func (n *ageFilter) Filter(ctx context.Context, req interface{}, container filter.DataContainer) bool {
	log.Printf("filter:%s,go:%v", FilterAge, ctx.Value("id"))
	data, err := container.Get(n, FieldUserBase)
	if err != nil {
		log.Printf("get user name fail,err:%v", err)
		return false
	}
	return data.(UserBase).Age >= 18
}
