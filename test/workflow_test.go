package test

import (
	"context"
	"log"
	"school/filter"
	"testing"
)

func TestWorkflow(t *testing.T) {
	loaders := []filter.Loader{&userBaseLoader{}, &userRelationLoader{}}
	filters := []filter.Filter{&nameFilter{}, &ageFilter{}, &relationFilter{}}
	manager := filter.NewDAGManager(loaders, filters) // 单实例，无状态
	var req = Request{
		userID:   10010,
		anchorID: 10011,
	}
	res := manager.Filter(context.Background(), req, filter.NewDataContainer(), FilterName, FilterAge, FilterRelation)
	log.Printf("---- result1:%v", res)
	// res = manager.Filter(context.Background(), req, filter.NewDataContainer(), FilterName, FilterRelation)
	// log.Printf("---- result2:%v", res)

	// var wg sync.WaitGroup
	// for i := 0; i < 10; i++ {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		res := manager.Filter(context.Background(), req, filter.NewDataContainer(), FilterName, FilterAge, FilterRelation)
	// 		log.Printf("---- result:%v", res)
	// 	}()
	// }
	// wg.Wait()
}
