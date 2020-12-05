package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing entry program")
		os.Exit(1)
	}

	mode := "ast"
	for i, arg := range os.Args[1:] {
		if arg == "--mode" || arg == "-mode" {
			mode = os.Args[i+2]
		}
	}

	f, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading entry program: %s", err)
		os.Exit(1)
	}

	a, ok := parse(string(f))
	if !ok {
		fmt.Println("Error parsing entry program")
		os.Exit(1)
	}

	var t1 time.Time
	var res int32
	switch mode {
	case "ast":
		t1 = time.Now()
		v := astInterpret(*a)
		res = *v.integer
	case "vm":
		bc, entrypoint, _ := vmCompile(*a)
		t1 = time.Now()
		res = vmRun(bc, entrypoint)
	case "compile":
		// code = compile(*a)
		t1 = time.Now()
		// res = run(code)
	default:
		panic("Unknown mode: " + mode)
	}

	t2 := time.Now()
	fmt.Printf("Result: %d, Time: %s\n", res, t2.Sub(t1))
}
