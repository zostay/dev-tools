package storage

import "errors"

var errCfg = errors.New("configuration is not writable")
var _ KV = &KVCfg{}

type KVCfg struct {
	KVMem
}

func (c *KVCfg) Clear() {
	panic(errCfg)
}

func (c *KVCfg) Set(string, any) {
	panic(errCfg)
}

func (c *KVCfg) Update(map[string]any) {
	panic(errCfg)
}

func (c *KVCfg) UpdateStrings(map[string]string) {
	panic(errCfg)
}
