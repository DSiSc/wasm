package exec

import (
	"fmt"
	"github.com/DSiSc/wasm/wasm"
	"reflect"
)

var nativeExporter = make(map[string]*wasm.Module)

func init() {
	nativeExporter["env"] = initEnvModule()
}

// NativeResolve resolve host dependency
func NativeResolve(name string) (*wasm.Module, error) {
	if _, ok := nativeExporter[name]; !ok {
		return nil, fmt.Errorf("Unknown import %s. ", name)
	}
	return nativeExporter[name], nil
}

func initEnvModule() *wasm.Module {
	envModule := wasm.NewModule()
	envModule.Types = &wasm.SectionTypes{
		Entries: []wasm.FunctionSig{},
	}
	envModule.FunctionIndexSpace = []wasm.Function{}
	envModule.Export = &wasm.SectionExports{Entries: make(map[string]wasm.ExportEntry)}

	appendFunc(envModule, "malloc", Malloc)
	appendFunc(envModule, "memcpy", Memcpy)
	appendFunc(envModule, "get_state", GetState)
	appendFunc(envModule, "set_state", SetState)
	return envModule
}

func appendFunc(m *wasm.Module, methodName string, method interface{}) {
	index := len(m.Types.Entries)
	funcSig := wasm.FunctionSig{
		Form:        0, // value for the 'func' type constructor
		ParamTypes:  funcArgTypes(method),
		ReturnTypes: funcReturnTypes(method),
	}
	m.Types.Entries = append(m.Types.Entries, funcSig)

	funcBody := wasm.Function{
		Sig:  &m.Types.Entries[index],
		Host: reflect.ValueOf(method),
		Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
	}
	m.FunctionIndexSpace = append(m.FunctionIndexSpace, funcBody)

	funcExport := wasm.ExportEntry{
		FieldStr: methodName,
		Kind:     wasm.ExternalFunction,
		Index:    uint32(index),
	}
	m.Export.Entries[methodName] = funcExport
}

// return a function's argument types
func funcArgTypes(f interface{}) []wasm.ValueType {
	t := reflect.TypeOf(f)
	n := t.NumIn()
	// skip the first param(first param must be exec.Process)
	wvt := make([]wasm.ValueType, n-1)
	for i := 1; i < n; i++ {
		wvt[i-1] = wasmType(t.In(i))
	}
	return wvt
}

// return a function's return types
func funcReturnTypes(f interface{}) []wasm.ValueType {
	t := reflect.TypeOf(f)
	n := t.NumOut()
	wvt := make([]wasm.ValueType, n)
	for i := 0; i < n; i++ {
		wvt[i] = wasmType(t.Out(i))
	}
	return wvt
}

// return wasm-type corresponding to the go-type
func wasmType(goType reflect.Type) wasm.ValueType {
	switch goType.Kind() {
	case reflect.Int32:
		return wasm.ValueTypeI32
	case reflect.Int64:
		return wasm.ValueTypeI64
	case reflect.Float32:
		return wasm.ValueTypeF32
	case reflect.Float64:
		return wasm.ValueTypeF32
	default:
		panic("unsupported param type")
	}
}
