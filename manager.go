package filter

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type handleType uint8

const (
	handlerType_Loader handleType = 0
	handlerType_Filter handleType = 1
)

// Manager 负责编排任务执行的管理器
type Manager interface {
	// Filter 编排执行，指定需要执行的filter
	Filter(ctx context.Context, req interface{}, dataContainer DataContainer, filterIDs ...string) bool
}

// dagManager
type dagManager struct {
	nodes   []*Node
	nodeMap map[string]*Node
}

func NewDAGManager(loaders []Loader, filters []Filter) Manager {
	nodes := make([]*Node, 0, len(loaders)+len(filters))
	nodeMap := make(map[string]*Node)
	// 将loader和filter转化成统一的节点
	for _, l := range loaders {
		if _, exist := nodeMap[l.ID()]; exist {
			panic(fmt.Errorf("node id:%s is exist", l.ID()))
		}
		node := newNode(loaderAdapter{l})
		nodes = append(nodes, node)
		nodeMap[l.ID()] = node
	}
	for _, f := range filters {
		if _, exist := nodeMap[f.ID()]; exist {
			panic(fmt.Errorf("node id:%s is exist", f.ID()))
		}
		node := newNode(filterAdapter{f})
		nodes = append(nodes, node)
		nodeMap[f.ID()] = node
	}
	getNodeByField := func(fieldID FieldID) *Node {
		for _, h := range nodes {
			// 只有loader生产数据，filter是纯消费数据
			if h.Type() == handlerType_Loader && h.OutputFields().Contains(fieldID) {
				return h
			}
		}
		panic(fmt.Errorf("unknown field:%s", fieldID))
	}
	// 构建node的父子关系
	for _, node := range nodes {
		for fieldID := range node.DependentFields() {
			parent := getNodeByField(fieldID)
			node.parents = append(node.parents, parent)
			parent.children = append(parent.children, node)
		}
	}

	wf := &dagManager{
		nodes:   nodes,
		nodeMap: nodeMap,
	}

	if wf.hasCycle() {
		panic("DAG has cycle")
	}

	return wf
}

// Filter 工作流执行入口
// filterIDs 指定需要执行的任务id，程序将自动计算执行链路
func (w *dagManager) Filter(ctx context.Context, req interface{}, dataContainer DataContainer, filterIDs ...string) bool {
	// 获取子DAG任务节点
	targetNodes := w.target(filterIDs...)
	// Node -> 入度
	nodeInDegree := make(map[string]int)
	// 入度为0的节点，即无数据依赖的节点，流程将从这些节点开始往下执行
	var startLoaderNodes []*Node
	var startFilterNodes []*Node // filterNode为叶子节点，且约定为简单的本地计算，故可以合并在一个协程中执行
	for _, h := range targetNodes {
		nodeInDegree[h.ID()] = len(h.parents)
		if nodeInDegree[h.ID()] == 0 {
			if h.Type() == handlerType_Loader {
				startLoaderNodes = append(startLoaderNodes, h)
			}
			if h.Type() == handlerType_Filter {
				startFilterNodes = append(startFilterNodes, h)
			}
		}
	}
	var mutex sync.Mutex
	resultChan := make(chan bool, len(targetNodes))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if len(startFilterNodes) > 0 {
		go w.runFilterNodes(context.WithValue(ctx, "id", time.Now().Nanosecond()), startFilterNodes, req, dataContainer, resultChan)
	}
	for _, node := range startLoaderNodes {
		go w.runLoaderNode(context.WithValue(ctx, "id", time.Now().Nanosecond()), node, nodeInDegree, req, dataContainer, &mutex, resultChan)
	}
	var doneNum int
	for isStop := range resultChan {
		// 任一filter返回不通过则流程结束，最终结果为不通过
		if isStop {
			return true
		}
		doneNum++
		// 所有节点执行完毕
		if doneNum == len(targetNodes) {
			return false
		}
	}
	return false
}

// runLoaderNode 编排执行DAG工作流
func (w *dagManager) runLoaderNode(ctx context.Context, node *Node, inDegree map[string]int, req interface{}, dataContainer DataContainer, mutex *sync.Mutex, done chan bool) {
	// 检查流程是否已终止
	select {
	case <-ctx.Done():
		return
	default:
	}
	// 执行节点业务逻辑
	isStop := node.Process(ctx, req, dataContainer)
	done <- isStop
	if isStop {
		return
	}
	// 检查流程是否已终止
	select {
	case <-ctx.Done():
		return
	default:
	}
	// 推动子节点执行
	var runnableLoaders []*Node
	var runnableFilters []*Node
	func() {
		mutex.Lock()
		defer mutex.Unlock()
		for _, child := range node.children {
			inDegree[child.ID()]--
			if inDegree[child.ID()] == 0 {
				if child.Type() == handlerType_Loader {
					runnableLoaders = append(runnableLoaders, child)
				}
				if child.Type() == handlerType_Filter {
					runnableFilters = append(runnableFilters, child)
				}
			}
		}
	}()

	if len(runnableFilters) > 0 {
		if len(runnableLoaders) > 0 {
			go w.runFilterNodes(context.WithValue(ctx, "id", time.Now().Nanosecond()), runnableFilters, req, dataContainer, done)
		} else { // 最后一个任务可以复用当前协程
			w.runFilterNodes(ctx, runnableFilters, req, dataContainer, done)
		}
	}

	for i, aNode := range runnableLoaders {
		if i != len(runnableLoaders)-1 {
			go w.runLoaderNode(context.WithValue(ctx, "id", time.Now().Nanosecond()), aNode, inDegree, req, dataContainer, mutex, done)
		} else { // 最后一个任务可以复用当前协程
			w.runLoaderNode(ctx, aNode, inDegree, req, dataContainer, mutex, done)
		}
	}

}

// runFilterNodes 串行执行filter
func (w *dagManager) runFilterNodes(ctx context.Context, nodes []*Node, req interface{}, dataContainer DataContainer, done chan bool) {
	for _, node := range nodes {
		isStop := node.Process(ctx, req, dataContainer)
		done <- isStop
		if isStop {
			return
		}
	}
}

// target 根据目标任务，获取子工作流（执行该工作流可完成目标任务），从而实现按需执行任务
func (w *dagManager) target(filterIDs ...string) []*Node {
	// 未指定目标任务默认执行完整工作流
	if len(filterIDs) == 0 {
		return w.nodes
	}
	var (
		targetNodes []*Node                     // 子DAG节点
		visited     = make(map[string]struct{}) // 保存DFS过的任务节点，减少重复搜索
		dfs         func(handler *Node)
	)
	// DFS搜集路径经过的节点，形成子DAG
	dfs = func(node *Node) {
		targetNodes = append(targetNodes, node)
		for _, parent := range node.parents {
			if _, ok := visited[parent.ID()]; !ok {
				dfs(parent)
			}
		}
		visited[node.ID()] = struct{}{}
	}
	// 从目标filter开始dfs
	for _, filterID := range filterIDs {
		h, ok := w.nodeMap[filterID]
		if !ok || h.Type() != handlerType_Filter {
			panic("filter is not exist or not a filter type")
		}
		dfs(h)
	}
	return targetNodes
}

// hasCycle DAG是否存在环
func (w *dagManager) hasCycle() bool {
	var (
		visited = make(map[string]uint8)
		dfs     func(node *Node)
		cycle   = false
	)
	dfs = func(handler *Node) {
		visited[handler.ID()] = 1 // 遍历中
		for _, child := range handler.children {
			if visited[child.ID()] == 0 {
				dfs(child)
				if cycle {
					return
				}
			} else if visited[child.ID()] == 1 {
				cycle = true
				return
			}
		}
		visited[handler.ID()] = 2 // 遍历完成
	}
	for _, handler := range w.nodes {
		if visited[handler.ID()] == 0 {
			dfs(handler)
			if cycle {
				return true
			}
		}
	}
	return cycle
}

// processingUnit 处理单元，loader和filter被封装成处理单元
type processingUnit interface {
	Consumer
	Producer
	ID() string                                                                        // 唯一id
	Type() handleType                                                                  // 类型
	Process(ctx context.Context, req interface{}, container DataContainer) (stop bool) // 处理逻辑，load数据或执行filter，返回是否流程结束
}

type loaderAdapter struct {
	Loader
}

func (l loaderAdapter) Type() handleType {
	return handlerType_Loader
}

func (l loaderAdapter) Process(ctx context.Context, req interface{}, container DataContainer) (stop bool) {
	l.Load(ctx, req, container)
	return false
}

type filterAdapter struct {
	Filter
}

func (f filterAdapter) OutputFields() StringSet {
	return NewStringSet()
}

func (f filterAdapter) Type() handleType {
	return handlerType_Filter
}

func (f filterAdapter) Process(ctx context.Context, req interface{}, container DataContainer) (stop bool) {
	return f.DoFilter(ctx, req, container)
}

type Node struct {
	processingUnit
	parents  []*Node
	children []*Node
}

func newNode(handler processingUnit) *Node {
	return &Node{
		processingUnit: handler,
		parents:        make([]*Node, 0),
		children:       make([]*Node, 0),
	}
}
