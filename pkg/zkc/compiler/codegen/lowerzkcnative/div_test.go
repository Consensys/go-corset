package lowerzkcnative_test

import (
    "testing"

    "github.com/consensys/go-corset/pkg/util/field"
    "github.com/consensys/go-corset/pkg/zkc/compiler/codegen"
    "github.com/consensys/go-corset/pkg/zkc/vm"
    "github.com/consensys/go-corset/pkg/zkc/vm/instruction/opcode"
    cmd_util "github.com/consensys/go-corset/pkg/cmd/zkc"
)

func Test_DivLowering(t *testing.T) {
    f := field.KOALABEAR_16
    program := cmd_util.CompileSourceFiles(f, "/tmp/test_div_simple.zkasm")
    cfg := codegen.DEFAULT_CONFIG.LowerZkcNative(true).Field(f)
    wm, errs := program.Compile(cfg)
    for _, e := range errs { t.Error(e) }
    if len(errs) > 0 { return }

    for _, mod := range wm.Modules() {
        fn, ok := mod.(*vm.WordFunction)
        if !ok { continue }
        t.Logf("function %s matched as *vm.WordFunction", fn.Name())
        for i, vec := range fn.Code() {
            for j, code := range vec.Codes {
                t.Logf("  [%d][%d] opcode=%v type=%T", i, j, code.OpCode(), code)
                if code.OpCode() == opcode.INT_DIV {
                    t.Errorf("found unlowered INT_DIV in %s at [%d][%d]", fn.Name(), i, j)
                }
            }
        }
    }
}
