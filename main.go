package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
)

func main() {
	args := os.Args[1:]
	checkArgs(args, 1, "must enter a command (n, copy, succ, add, double, multiply)")
	cmd := args[0]
	args = args[1:]

	// result
	res := NewInteger()

	switch cmd {
	case "n":
		checkArgs(args, 1, "must enter a number")
		n := parseInt(args[0])
		go N(n, res)
		fmt.Println(result(res))
	case "copy":
		checkArgs(args, 1, "must enter a number")
		n := parseInt(args[0])
		i := NewInteger()
		go N(n, i)
		go Copy(i, res)
		fmt.Println(result(res))
	case "succ":
		checkArgs(args, 1, "must enter a number")
		n := parseInt(args[0])
		i := NewInteger()
		go N(n, i)
		go Succ(i, res)
		fmt.Println(result(res))
	case "add":
		checkArgs(args, 2, "must enter two numbers to add")
		n1 := parseInt(args[0])
		n2 := parseInt(args[1])
		i1, i2 := NewInteger(), NewInteger()
		go N(n1, i1)
		go N(n2, i2)
		go Add(i1, i2, res)
		fmt.Println(result(res))
	case "double":
		checkArgs(args, 1, "must enter a number")
		n := parseInt(args[0])
		i, res2 := NewInteger(), NewInteger()
		go N(n, i)
		go Double(i, res, res2)

		wg := new(sync.WaitGroup)
		wg.Add(2)
		go printResult(res, wg)
		go printResult(res2, wg)
		wg.Wait()
	case "multiply":
		checkArgs(args, 2, "must enter two numbers to multiply")
		n1 := parseInt(args[0])
		n2 := parseInt(args[1])
		i1, i2 := NewInteger(), NewInteger()
		go N(n1, i1)
		go N(n2, i2)
		go Multiply(i1, i2, res)
		fmt.Println(result(res))
	default:
		fmt.Println("unknown command", cmd)
	}
}

func printResult(res *Integer, wg *sync.WaitGroup) {
	v := result(res)
	fmt.Println(v)
	wg.Done()
}

func result(i *Integer) (res int) {
TALLY:
	for {
		select {
		case <-i.X:
			res += 1
		case <-i.Z:
			break TALLY
		}
	}
	return res
}

func checkArgs(args []string, n int, s string) {
	if len(args) < n {
		fmt.Println(s)
		Exit()
	}
}

func parseInt(s string) int {
	n, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		fmt.Println("expected an integer. got", s)
		Exit()
	}
	return int(n)
}

func Exit() {
	os.Exit(1)
}
