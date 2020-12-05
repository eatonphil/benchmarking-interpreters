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
	instructionJumpZ
	instructionPush
	instructionPop
	instructionSwap
	instructionParam
	instructionPrint
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
		vc.compileBlock(args, ctx, false)
		return
	case "+":
		vc.compileValue(args[0], ctx)
		vc.compileValue(args[1], ctx)
		vc.emit(instructionAdd)
		return
	case "-":
		vc.compileValue(args[0], ctx)
		vc.compileValue(args[1], ctx)
		vc.emit(instructionSub)
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
		vc.emitWithArg(instructionJumpZ, 0)
		// Need to fix since we don't yet know where to jump to reach the else
		addressToFix1 := len(vc.bc)

		vc.compileValue(args[1], ctx)
		// Jump to after else condition
		vc.emitWithArg(instructionJump, 0)
		addressAfterFirstBlock := len(vc.bc)
		vc.bc[addressToFix1-1] |= instruction(addressAfterFirstBlock) << 8

		vc.compileValue(args[2], ctx)
		vc.bc[addressAfterFirstBlock - 1] |= instruction(len(vc.bc)) << 8
		return
	case "def":
		fn := args[0]
		params := args[1]
		body := args[2:]

		ctx[*fn.identifier] = int32(len(vc.bc))

		childCtx := map[string]int32{}
		for key, val := range ctx {
			childCtx[key] = val
		}

		for i, param := range *params.list {
			childCtx[*param.identifier] = int32(i)
		}

		vc.compileBlock(body, childCtx, false)

		vc.emit(instructionRet)
		return
	case "print":
		vc.compileValue(args[0], ctx)
		vc.emit(instructionPrint)
		return
	}

	for _, arg := range args {
		vc.compileValue(arg, ctx)
	}

	functionAddress, ok := ctx[*fn.identifier]
	if !ok {
		panic("Unknown function: "+*fn.identifier)
	}

	vc.emitWithArg(instructionCall, functionAddress)
	vc.emitWithArg(instructionSwap, int32(len(args)))

	for range args {
		vc.emit(instructionPop)
	}
}

func (vc *vmCompiler) compileValue(a ast, ctx map[string]int32) {
	switch a.kind {
	case astInteger:
		vc.emitWithArg(instructionPush, *a.integer)
	case astIdentifier:
		vc.emitWithArg(instructionParam, ctx[*a.identifier])
	case astList:
		vc.compileList(*a.list, ctx)
	}
}

func (vc *vmCompiler) compileBlock(a []ast, ctx map[string]int32, topLevel bool) {
	for i, value := range a {
		vc.compileValue(value, ctx)

		if i < len(a)-1 && !topLevel {
			vc.bc = append(vc.bc, instructionPop)
		}
	}
}

func vmCompile(a ast) ([]instruction, int32, map[string]int32) {
	vc := vmCompiler{}
	ctx := map[string]int32{}
	vc.compileBlock(*a.list, ctx, true)
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
		case instructionJumpZ:
			s = append(s, fmt.Sprintf("jump-zero %d", arg))
		case instructionPush:
			s = append(s, fmt.Sprintf("push %d", arg))
		case instructionParam:
			s = append(s, fmt.Sprintf("param %d", arg))
		case instructionSwap:
			s = append(s, fmt.Sprintf("swap %d", arg))
		case instructionPop:
			s = append(s, "pop")
		case instructionPrint:
			s = append(s, "print")
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
				label = key
			}
		}
		lines += fmt.Sprintf("%d\t\t%s\t\t%s\r\n", i, label, line)
	}

	return lines
}

func vmRun(bc []instruction, entrypoint int32) int32 {
	stack := make([]int32, 10240)
	ip := entrypoint
	var sp int32 = 0
	var fp int32 = 0

	stack[sp] = 0
	sp++
	stack[sp] = int32(len(bc))
	bc = append(bc, instructionHalt)

	for {
		in := bc[ip] & 0xFF
		if in == instructionHalt {
			break
		}

		inArg := int32(bc[ip] >> 8)

		if sp >= int32(len(stack)) {
			panic("Stack overflow")
		}

		//fmt.Printf("IP: %d,\tIN: %x, FP: %d, SP: %d\n", ip, []byte{byte(bc[ip] >> 24 & 0xFF), byte(bc[ip] >> 16 & 0xFF), byte(bc[ip] >> 8 & 0xFF), byte(bc[ip] & 0xFF)}, fp, sp)

		switch in {
		case instructionAdd:
			arg1 := stack[sp]
			sp--
			stack[sp] += arg1
		case instructionSub:
			arg1 := stack[sp]
			sp--
			arg2 := stack[sp]
			stack[sp] = arg2 - arg1
		case instructionGte:
			arg1 := stack[sp]
			sp--
			arg2 := stack[sp]
			stack[sp] = 0
			if arg2 >= arg1 {
				stack[sp] = 1
			}
		case instructionGt:
			arg1 := stack[sp]
			sp--
			arg2 := stack[sp]
			stack[sp] = 0
			if arg2 > arg1 {
				stack[sp] = 1
			}
		case instructionLte:
			arg1 := stack[sp]
			sp--
			arg2 := stack[sp]
			stack[sp] = 0
			if arg2 <= arg1 {
				stack[sp] = 1
			}
		case instructionLt:
			arg1 := stack[sp]
			sp--
			arg2 := stack[sp]
			stack[sp] = 0
			if arg2 < arg1 {
				stack[sp] = 1
			}
		case instructionCall:
			sp++
			stack[sp] = fp
			sp++
			stack[sp] = ip+1
			fp = sp
			ip = inArg
			continue
		case instructionRet:
			ret := stack[sp]
			sp--
			ip = stack[sp]
			sp--
			fp = stack[sp]
			stack[sp] = ret
			continue
		case instructionJump:
			ip = inArg
			continue
		case instructionJumpZ:
			arg1 := stack[sp]
			sp--

			if arg1 == 0 {
				ip = inArg
				continue
			}
		case instructionPush:
			sp++
			stack[sp] = inArg
		case instructionPop:
			sp--
		case instructionParam:
			sp++
			stack[sp] = stack[fp-inArg-2]
		case instructionSwap:
			a := stack[sp]
			stack[sp] = stack[sp-inArg]
			stack[sp-inArg] = a
		case instructionPrint:
			fmt.Println(stack[sp])
			// no sp--, every function must leave something on the stack
		default:
			panic(fmt.Sprintf("Unknown instruction: %x, ip: %d", in, ip))
		}

		ip++
	}

	return stack[sp]
}
