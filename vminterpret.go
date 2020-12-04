package main

import "fmt"

type instruction int64

const (
	instructionAdd instruction = iota
	instructionSub
	instructionGte
	instructionGt
	instructionLte
	instructionLt
	instructionCall
	instructionRet
	instructionJump
	instructionPush
	instructionPushFromFrameOffset
	instructionPop
	instructionHalt
)

type vmCompiler struct {
	bc []instruction
}

func (vc *vmCompiler) emit(in instruction) {
	vc.bc = append(vc.bc, in)
}

func (vc *vmCompiler) emitWithArg(in instruction, arg int32) {
	vc.emit((instruction(arg) << 8) | in)
}

func (vc *vmCompiler) compileList(a []ast, ctx map[string]int32) {
	fn := a[0]
	args := a[1:]

	switch *fn.identifier {
	case "begin":
		vc.compileBlock(args, ctx)
		return
	case "+":
		vc.compileValue(args[0], ctx)
		vc.compileValue(args[1], ctx)
		vc.emit(instructionAdd)
		return
	case "-":
		vc.compileValue(args[0], ctx)
		vc.compileValue(args[1], ctx)
		vc.emit(instructionAdd)
		return
	case ">=":
		vc.compileValue(args[0], ctx)
		vc.compileValue(args[1], ctx)
		vc.emit(instructionGte)
		return
	case ">":
		vc.compileValue(args[0], ctx)
		vc.compileValue(args[1], ctx)
		vc.emit(instructionGt)
		return
	case "<=":
		vc.compileValue(args[0], ctx)
		vc.compileValue(args[1], ctx)
		vc.emit(instructionLte)
		return
	case "<":
		vc.compileValue(args[0], ctx)
		vc.compileValue(args[1], ctx)
		vc.emit(instructionLt)
		return
	case "if":
		vc.compileValue(args[0], ctx)
		vc.emitWithArg(instructionJump, 0)
		// Need to fix since we don't yet know where to jump to reach the else
		addressToFix := len(vc.bc)

		vc.compileValue(args[1], ctx)
		addressAfter := len(vc.bc)

		vc.bc[addressToFix] |= instruction(addressAfter) << 8

		vc.compileValue(args[2], ctx)
		return
	case "def":
		fn := args[0]
		params := args[1]
		body := args[2:]

		ctx[*fn.identifier] = int32(len(vc.bc))
		fmt.Println("assigning fn", *fn.identifier, ctx[*fn.identifier])

		childCtx := map[string]int32{}
		for key, val := range ctx {
			childCtx[key] = val
		}

		for i, param := range *params.list {
			childCtx[*param.identifier] = int32(i)
		}

		vc.compileBlock(body, childCtx)

		vc.emit(instructionRet)
		return
	}

	for _, arg := range args {
		vc.compileValue(arg, ctx)
	}

	fmt.Println("calling", *fn.identifier, ctx[*fn.identifier])
	vc.emitWithArg(instructionCall, ctx[*fn.identifier])

	for range args {
		vc.emit(instructionPop)
	}
}

func (vc *vmCompiler) compileValue(a ast, ctx map[string]int32) {
	switch a.kind {
	case astInteger:
		vc.emitWithArg(instructionPush, *a.integer)
	case astIdentifier:
		vc.emitWithArg(instructionPushFromFrameOffset, ctx[*a.identifier])
	case astList:
		vc.compileList(*a.list, ctx)
	}
}

func (vc *vmCompiler) compileBlock(a []ast, ctx map[string]int32) {
	for i, value := range a {
		vc.compileValue(value, ctx)

		if i < len(a)-1 {
			vc.bc = append(vc.bc, instructionPop)
		}
	}
}

func vmCompile(a ast) ([]instruction, int32, map[string]int32) {
	vc := vmCompiler{}
	ctx := map[string]int32{}
	vc.compileBlock(*a.list, ctx)
	return vc.bc, ctx["main"], ctx
}

func vmDisassemble(bc []instruction, ctx map[string]int32) string {
	var s []string

	for i := 0; i < len(bc); i++ {
		in := bc[i] & 0xFF
		arg := bc[i] >> 8
		switch in {
		case instructionAdd:
			s = append(s, "add")
		case instructionSub:
			s = append(s, "sub")
		case instructionGte:
			s = append(s, "gte")
		case instructionGt:
			s = append(s, "gt")
		case instructionLte:
			s = append(s, "lte")
		case instructionLt:
			s = append(s, "lt")
		case instructionCall:
			s = append(s, fmt.Sprintf("call %d", arg))
		case instructionRet:
			s = append(s, fmt.Sprintf("ret"))
		case instructionJump:
			s = append(s, fmt.Sprintf("jump %d", arg))
		case instructionPush:
			fmt.Println("push", in, bc[i], arg)
			s = append(s, fmt.Sprintf("push %d", arg))
		case instructionPushFromFrameOffset:
			s = append(s, fmt.Sprintf("push-from-frame-offset %d", arg))
		case instructionPop:
			s = append(s, "pop")
		case instructionHalt:
			s = append(s, "halt")
		default:
			s = append(s, fmt.Sprintf("Unknown instruction: %d, arg: %d", in, arg))
		}
	}

	lines := ""
	for i, line := range s {
		label := ""
		for key, val := range ctx {
			if val == int32(i) {
				label = "# " + key
			}
		}
		lines += fmt.Sprintf("%d\t\t%s\t\t%s\r\n", i, label, line)
	}

	return lines
}

func vmRun(bc []instruction, entrypoint int32) int32 {
	stack := make([]int32, 10240)
	ip := entrypoint
	sp := 0
	fp := int32(0)

	for in := bc[ip] & 0xFF; in != instructionHalt; in = bc[ip] {
		switch in {
		case instructionAdd:
			arg1 := stack[sp]
			sp--
			arg2 := stack[sp]
			stack[sp] = arg1 + arg2
		case instructionSub:
			arg1 := stack[sp]
			sp--
			arg2 := stack[sp]
			stack[sp] = arg1 + arg2
		case instructionGte:
			arg1 := stack[sp]
			sp--
			arg2 := stack[sp]
			stack[sp] = 0
			if arg1 >= arg2 {
				stack[sp] = 1
			}
		case instructionGt:
			arg1 := stack[sp]
			sp--
			arg2 := stack[sp]
			stack[sp] = 0
			if arg1 > arg2 {
				stack[sp] = 1
			}
		case instructionLte:
			arg1 := stack[sp]
			sp--
			arg2 := stack[sp]
			stack[sp] = 0
			if arg1 <= arg2 {
				stack[sp] = 1
			}
		case instructionLt:
			arg1 := stack[sp]
			sp--
			arg2 := stack[sp]
			stack[sp] = 0
			if arg1 < arg2 {
				stack[sp] = 1
			}
		case instructionCall:
			stack[sp] = fp
			sp++
			stack[sp] = ip
			sp++
			ip = int32(bc[ip] >> 8)
			continue
		case instructionRet:
			ip = stack[sp]
			sp--
			fp = stack[sp]
			sp--
			continue
		case instructionJump:
			arg1 := stack[sp]
			sp--

			// zero check may be backwards but makes `if` easier
			if arg1 == 0 {
				ip = int32(bc[ip] >> 8)
				continue
			}
		case instructionPush:
			stack[sp] = int32(bc[ip] >> 8)
			sp++
		case instructionPushFromFrameOffset:
			stack[sp] = stack[fp-int32(bc[ip]>>8)-1]
		case instructionPop:
			sp--
		default:
			panic(fmt.Sprintf("Unknown instruction: %+v", in))
		}

		ip++
	}

	return stack[sp]
}
