package test

import (
	"context"
	"log"
	"school/filter"
	"time"
)

const (
	FieldUserRelation  = "user_relation"
	LoaderUserRelation = "user_relation_loader"
)

type userRelationLoader struct {
}

func (u *userRelationLoader) ID() string {
	return LoaderUserRelation
}

func (u *userRelationLoader) ConsumeFields() filter.StringSet {
	return filter.NewStringSet(FieldUserBase)
}

func (u *userRelationLoader) ProduceFields() filter.StringSet {
	return filter.NewStringSet(FieldUserRelation)
}

func (u *userRelationLoader) Load(ctx context.Context, req interface{}, container filter.DataContainer) {
	log.Printf("loader:%s,go:%v", LoaderUserRelation, ctx.Value("id"))
	// rpc获取关系信息
	var relation uint8 = 1
	container.Set(u, FieldUserRelation, relation, nil)
	time.Sleep(time.Second)
}
