package main

import "fmt"

type valueKind int

const (
	valueInteger valueKind = iota
	valueIdentifier
	valueFunction
	valueList
)

type value struct {
	kind       valueKind
	integer    *int32
	identifier *string
	function   func([]value, map[string]value) value
	list       *[]value
}

func (v value) toString() string {
	switch v.kind {
	case valueInteger:
		return fmt.Sprintf("%d", *v.integer)
	case valueIdentifier:
		return *v.identifier
	case valueFunction:
		return "<function>"
	case valueList:
		s := "("

		for i, l := range *v.list {
			s += l.toString()
			if i < len(*v.list)-1 {
				s += " "
			}
		}

		s += ")"

		return s
	}
	return ""
}

func valueInterpret(v value, ctx map[string]value) value {
	switch v.kind {
	case valueList:
		return valueInterpretList(*v.list, ctx)
	case valueIdentifier:
		i, ok := ctx[*v.identifier]
		if !ok {
			panic("Unknown identifier: " + *v.identifier)
		}
		return i
	default:
		return v
	}
}

func valueInterpretBlock(list []value, ctx map[string]value) value {
	var res value

	for _, v := range list {
		res = valueInterpret(v, ctx)
	}

	return res
}

func valueInterpretList(list []value, ctx map[string]value) value {
	fn := list[0]
	args := list[1:]

	switch *fn.identifier {
	case "begin":
		return valueInterpretBlock(args, ctx)
	case ">":
		i32 := int32(0)
		if *valueInterpret(args[0], ctx).integer > *valueInterpret(args[1], ctx).integer {
			i32 = 1
		}
		return value{
			kind:    valueInteger,
			integer: &i32,
		}
	case ">=":
		i32 := int32(0)
		if *valueInterpret(args[0], ctx).integer >= *valueInterpret(args[1], ctx).integer {
			i32 = 1
		}
		return value{
			kind:    valueInteger,
			integer: &i32,
		}
	case "<":
		i32 := int32(0)
		if *valueInterpret(args[0], ctx).integer < *valueInterpret(args[1], ctx).integer {
			i32 = 1
		}
		return value{
			kind:    valueInteger,
			integer: &i32,
		}
	case "<=":
		i32 := int32(0)
		if *valueInterpret(args[0], ctx).integer <= *valueInterpret(args[1], ctx).integer {
			i32 = 1
		}
		return value{
			kind:    valueInteger,
			integer: &i32,
		}
	case "+":
		i32 := int32(*valueInterpret(args[0], ctx).integer + *valueInterpret(args[1], ctx).integer)
		return value{
			kind:    valueInteger,
			integer: &i32,
		}
	case "-":
		i32 := int32(*valueInterpret(args[0], ctx).integer - *valueInterpret(args[1], ctx).integer)
		return value{
			kind:    valueInteger,
			integer: &i32,
		}
	case "if":
		if *valueInterpret(args[0], ctx).integer != 0 { // true
			return valueInterpret(args[1], ctx)
		}

		return valueInterpret(args[2], ctx)
	case "def":
		name := args[0]
		params := args[1]
		body := args[2:]
		ctx[*name.identifier] = value{
			kind: valueFunction,
			function: func(args []value, ctx map[string]value) value {
				childCtx := map[string]value{}
				for key, val := range ctx {
					childCtx[key] = val
				}

				for i, param := range *params.list {
					childCtx[*param.identifier] = valueInterpret(args[i], ctx)
				}

				return valueInterpretBlock(body, childCtx)
			},
		}
		return ctx[*name.identifier]
	}

	fnc := valueInterpret(fn, ctx)
	return fnc.function(args, ctx)
}

func astToValue(a ast) value {
	switch a.kind {
	case astInteger:
		return value{
			kind:    valueInteger,
			integer: a.integer,
		}
	case astIdentifier:
		return value{
			kind:       valueIdentifier,
			identifier: a.identifier,
		}
	default:
		var list []value
		for _, v := range *a.list {
			list = append(list, astToValue(v))
		}

		return value{
			kind: valueList,
			list: &list,
		}
	}
}

func astInterpret(a ast) value {
	astValue := astToValue(a)

	// Add on a call to main
	mainIdentifier := "main"
	program := append(
		*astValue.list,
		value{
			kind: valueList,
			list: &[]value{
				value{
					kind:       valueIdentifier,
					identifier: &mainIdentifier,
				},
			},
		})

	return valueInterpretBlock(program, map[string]value{})
}
