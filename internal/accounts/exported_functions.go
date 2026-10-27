package accounts

var functionExporters []FunctionExporter

// FunctionExporter is satisfied by the plugin registry, allowing the users
// package to call exported plugin functions without importing internal/plugins.
type FunctionExporter interface {
	GetExportedFunction(funcName string) (any, bool)
}

// AddFunctionExporter registers a function exporter (called from main.go with
// plugins.GetPluginRegistry()).
func AddFunctionExporter(f FunctionExporter) {
	functionExporters = append(functionExporters, f)
}

// GetExportedFunction looks up a named function across all registered exporters.
func GetExportedFunction(fName string) (any, bool) {
	for _, x := range functionExporters {
		if f, ok := x.GetExportedFunction(fName); ok {
			return f, ok
		}
	}
	return nil, false
}
