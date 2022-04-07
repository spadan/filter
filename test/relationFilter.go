package test

import (
	"context"
	"log"
	"school/filter"
)

const (
	FilterRelation = "relation_filter"
)

type relationFilter struct {
}

func (n *relationFilter) ID() string {
	return FilterRelation
}

func (n *relationFilter) ConsumeFields() filter.StringSet {
	return filter.NewStringSet(FieldUserRelation)
}

func (n *relationFilter) Filter(ctx context.Context, req interface{}, container filter.DataContainer) bool {
	log.Printf("filter:%s,go:%v", FilterRelation, ctx.Value("id"))
	data, err := container.Get(n, FieldUserRelation)
	if err != nil {
		log.Printf("get user relation fail,err:%v", err)
		return false
	}
	return data.(uint8) == 0
}
