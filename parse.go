package main

import (
	"fmt"
	"strconv"
)

type astKind uint

const (
	astInteger astKind = iota
	astIdentifier
	astList
)

type ast struct {
	kind       astKind
	integer    *int32
	identifier *string
	list       *[]ast
}

func (a ast) toString() string {
	switch a.kind {
	case astInteger:
		return fmt.Sprintf("%d", *a.integer)
	case astIdentifier:
		return *a.identifier
	case astList:
		s := "("

		for i, l := range *a.list {
			s += l.toString()
			if i < len(*a.list)-1 {
				s += " "
			}
		}

		s += ")"

		return s
	}
	return ""
}

func parseInteger(source string, cursor int) (*ast, bool, int) {
	t := ""
	for source[cursor] >= '0' && source[cursor] <= '9' {
		t += string(source[cursor])
		cursor++
	}

	if t == "" {
		return nil, false, 0
	}

	i, err := strconv.Atoi(t)
	if err != nil {
		return nil, false, 0
	}

	i32 := int32(i)

	return &ast{
		kind:    astInteger,
		integer: &i32,
	}, true, cursor
}

func parseIdentifier(source string, cursor int) (*ast, bool, int) {
	t := ""
	for {
		c := source[cursor]
		if (c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			c == '>' || c == '<' || c == '=' || c == '!' ||
			c == '+' || c == '-' {
			t += string(c)
			cursor++
		} else {
			break
		}
	}

	ok := t != ""

	return &ast{
		kind:       astIdentifier,
		identifier: &t,
	}, ok, cursor
}

func parseList(source string, cursor int) (*ast, bool, int) {
	list := []ast{}

	foundList := false

	if cursor == len(source)-1 {
		return nil, true, cursor
	}

	for {
		c := source[cursor]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			cursor++
			continue
		}

		// !foundList so that child parens don't go to this case
		if c == '(' && !foundList {
			foundList = true
			cursor++
			continue
		}

		if !foundList {
			return nil, false, 0
		}

		if c == ')' {
			cursor++
			break
		}

		integer, ok, newCursor := parseInteger(source, cursor)
		if ok {
			cursor = newCursor
			list = append(list, *integer)
			continue
		}

		identifier, ok, newCursor := parseIdentifier(source, cursor)
		if ok {
			cursor = newCursor
			list = append(list, *identifier)
			continue
		}

		l, ok, newCursor := parseList(source, cursor)
		if ok {
			cursor = newCursor
			list = append(list, *l)
			continue
		}

		fmt.Println("Error, expected valid list child")
		return nil, false, 0
	}

	return &ast{
		kind: astList,
		list: &list,
	}, true, cursor
}

func parse(source string) (*ast, bool) {
	program := []ast{}
	cursor := 0
	for {
		a, ok, newCursor := parseList(source, cursor)
		if !ok {
			if len(program) > 0 {
				fmt.Println("Error after: " + program[len(program)-1].toString())
			}
			return nil, false
		}

		if a == nil {
			break
		}

		program = append(program, *a)
		cursor = newCursor
	}

	return &ast{
		kind: astList,
		list: &program,
	}, true
}
