package main

import "fmt"

type SafeMap interface {
	Get(key string) interface{}
	Set(key string, value interface{})
	Delete(key string)
}

type Map struct {
	M     map[string]interface{}
	C     chan *mapOperation
	Close func()
}

func NewMap() Map {
	m := Map{}
	m.C = make(chan *mapOperation)
	m.M = make(map[string]interface{})
	done := make(chan struct{}) // 用于关闭 goroutine 的信号
	// 管理器 goroutine
	go func() {
		for {
			select {
			case op := <-m.C:
				switch op.action {
				case 1: //set
					m.M[op.key] = op.value
				case 0: //get
					if value, exists := m.M[op.key]; exists {
						op.result <- value
					} else {
						op.result <- nil
					}
				case 2: //delete
					delete(m.M, op.key)
				}
			case <-done:
				return
			}
		}
	}()
	m.Close = func() {
		close(done)
	}
	return m
}

type mapOperation struct {
	action int         // 操作类型："0get"、"1set"、"2delete"
	key    string      // 键
	value  interface{} // 值（用于 set 操作）
	result chan interface{}
}

func (m *Map) Get(key string) interface{} {
	data := mapOperation{
		action: 0,
		key:    key,
		result: make(chan interface{}, 1),
	}
	m.C <- &data
	return <-data.result
}
func (m *Map) Set(key string, value interface{}) {
	data := mapOperation{
		action: 1,
		key:    key,
		value:  value,
	}
	m.C <- &data
}
func (m *Map) Delete(key string) {
	data := mapOperation{
		action: 2,
		key:    key,
	}
	m.C <- &data
}

func main() {
	safeMap := NewMap()

	safeMap.Set("key1", "value1")
	value := safeMap.Get("key1")
	fmt.Println(value)
	safeMap.Delete("key1")
	value = safeMap.Get("key1")
	fmt.Println(value)
	// 关闭 Map
	safeMap.Close()
}
