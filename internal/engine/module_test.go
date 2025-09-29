package engine_test

import (
	"fmt"
	"testing"

	"github.com/direktiv/direktiv/internal/engine"
	"github.com/google/uuid"
	"github.com/grafana/sobek"
)

// type ModuleImpl struct {
// }

// func (mi *ModuleImpl) ResolveExport(exportName string, resolveset ...sobek.ResolveSetElement) (*sobek.ResolvedBinding, bool) {
// 	return nil, false
// }

// func (mi *ModuleImpl) GetExportedNames(callback func([]string), records ...sobek.ModuleRecord) bool {
// 	return true
// }

// func (mi *ModuleImpl) Link() error {
// 	// this does nothing on this
// 	return nil
// }

// func (mi *ModuleImpl) Evaluate(rt *sobek.Runtime) *sobek.Promise {
// 	return nil
// }

// func gg(referencingScriptOrModule interface{}, specifier string) (sobek.ModuleRecord, error) {
// 	return nil, nil
// }

// (function () {
//   function gg() {
//     return "hello";
//   }

//   // Register under a namespace
//   globalThis.__modules = globalThis.__modules || {};
//   globalThis.__modules["gg"] = { gg };
// })();

func TestXxx(t *testing.T) {

	var scriptmodule = `
	function jens() {
		log("JENS")
	}
	`

		_ CyclicModuleRecord   = &SourceTextModuleRecord{}
	_ CyclicModuleInstance = &SourceTextModuleInstance{}
	// p := m.(*sobek.SourceTextModuleRecord)

	vm := sobek.New()
	engine.InjectCommands(vm, uuid.New())

	// promise := vm.CyclicModuleRecordEvaluate(p, gg)

}

// cmds.vm.CyclicModuleRecordEvaluate()
// cmds.vm.FinishLoadingImportModule()
// cmds.vm.GetActiveScriptOrModule()
// cmds.vm.GetModuleInstance()
// cmds.vm.NamespaceObjectFor()
// cmds.vm.SetFinalImportMeta()
// cmds.vm.SetGetImportMetaProperties()
