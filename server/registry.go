package server

import (
	"sync"
	"time"
)

type cachedValue struct {
	value     string
	createdAt int64
}

type TTLMap struct {
	ttlMap map[string]map[string]cachedValue
	mutex  sync.Mutex
}

func CreateTTLMap(maxTTL int) (ttlMap *TTLMap) {
	ttlMap = &TTLMap{ttlMap: make(map[string]map[string]cachedValue)}
	go func() {
		for now := range time.Tick(time.Minute * 10) {
			ttlMap.mutex.Lock()
			for key, valueList := range ttlMap.ttlMap {
				for index, value := range valueList {
					if now.Unix()-value.createdAt > int64(maxTTL) {
						delete(valueList, index)
					}
				}
				if 0 == len(valueList) {
					delete(ttlMap.ttlMap, key)
				}
			}
			ttlMap.mutex.Unlock()
		}
	}()
	return
}

func (ttlMap *TTLMap) Len() int {
	return len(ttlMap.ttlMap)
}

func (ttlMap *TTLMap) Put(key, id, value string) {
	ttlMap.mutex.Lock()
	group, ok := ttlMap.ttlMap[key]
	if !ok {
		group = make(map[string]cachedValue)
		ttlMap.ttlMap[key] = group
	}
	it, ok := group[id]
	if !ok {
		it = cachedValue{}
	}
	it.value = value
	it.createdAt = time.Now().Unix()
	ttlMap.ttlMap[key][id] = it
	ttlMap.mutex.Unlock()
}

func (ttlMap *TTLMap) Get(key string) (values []string) {
	ttlMap.mutex.Lock()
	if list, ok := ttlMap.ttlMap[key]; ok {
		values = make([]string, len(list))
		i := 0
		for _, cachedValue := range list {
			values[i] = cachedValue.value
			i++
		}
	} else {
		values = make([]string, 0)
	}
	ttlMap.mutex.Unlock()
	return
}

func (ttlMap *TTLMap) Keys() (keys string) {
	keys = ""
	for key, _ := range ttlMap.ttlMap {
		keys += key
	}
	return
}
