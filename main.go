package main

import (
	"dashboard/app/data"
	"dashboard/service"
)

func main() {
	data.Run()
	service.Run()
}
