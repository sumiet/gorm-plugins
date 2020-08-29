package main

import (
	"fmt"
	"runtime"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	ORM_MAX_IDLE = 200
	ORM_MAX_CONN = 200
	ORM_SOURCE = "uber:uber@/orm_bench?charset=utf8"
	ORM_MULTI = 50

	registerGORM()
	registerRAW()
	registerXORM()

	for _, n := range BrandNames {
		fmt.Println(n)
		RunBenchmark(n)
	}

	fmt.Printf("\nReport: %d multiplier\n", ORM_MULTI)
	fmt.Print(MakeReport())

}
