package test

import (
	"context"
	"log"
	"school/filter"
	"time"
)

const (
	FieldUserBase  = "user_base"
	LoaderUserBase = "user_base_loader"
)

type userBaseLoader struct {
}

func (u *userBaseLoader) ID() string {
	return LoaderUserBase
}

func (u *userBaseLoader) ConsumeFields() filter.StringSet {
	return filter.NewStringSet()
}

func (u *userBaseLoader) ProduceFields() filter.StringSet {
	return filter.NewStringSet(FieldUserBase)
}

func (u *userBaseLoader) Load(ctx context.Context, req interface{}, container filter.DataContainer) {
	log.Printf("loader:%s,go:%v", LoaderUserBase, ctx.Value("id"))
	request := req.(Request)
	// rpc获取用户基础信息
	userBase := UserBase{
		ID:   request.userID,
		Name: "zhangSan",
		Age:  21,
		City: "shenzhen",
	}
	container.Set(u, FieldUserBase, userBase, nil)
	time.Sleep(time.Second)
}

type UserBase struct {
	ID   int64
	Name string
	Age  uint8
	City string
}
