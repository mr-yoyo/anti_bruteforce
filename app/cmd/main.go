package main

import (
	_ "github.com/lib/pq"
	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/cmd"
)

func main() {
	_ = cmd.Execute()
}
