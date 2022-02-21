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

func (n *ageFilter) DependentFields() filter.StringSet {
	return filter.NewStringSet(FieldUserBase)
}

func (n *ageFilter) DoFilter(ctx context.Context, req interface{}, container filter.DataContainer) bool {
	log.Printf("filter:%s,go:%v", FilterAge, ctx.Value("id"))
	data, err := container.GetDependentData(n, FieldUserBase)
	if err != nil {
		log.Printf("get user name fail,err:%v", err)
		return false
	}
	return data.(UserBase).Age >= 18
}
