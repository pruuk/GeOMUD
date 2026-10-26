package scripting

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["roomVMCache"] = util.MemoryResult{Memory: util.MemoryUsage(roomVMCache), Count: len(roomVMCache)}
	ret["mobVMCache"] = util.MemoryResult{Memory: util.MemoryUsage(mobVMCache), Count: len(mobVMCache)}
	ret["itemVMCache"] = util.MemoryResult{Memory: util.MemoryUsage(itemVMCache), Count: len(itemVMCache)}
	ret["buffVMCache"] = util.MemoryResult{Memory: util.MemoryUsage(buffVMCache), Count: len(buffVMCache)}
	ret["spellVMCache"] = util.MemoryResult{Memory: util.MemoryUsage(spellVMCache), Count: len(spellVMCache)}
	ret["moduleFunctions"] = util.MemoryResult{Memory: util.MemoryUsage(moduleFunctions), Count: len(moduleFunctions)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Scripting`, GetMemoryUsage)
}
