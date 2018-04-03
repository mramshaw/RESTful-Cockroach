package main

import "os"

import "application"

func main() {
	app := application.App{}
	app.Initialize(
		os.Getenv("COCKROACH_USER"),
		os.Getenv("COCKROACH_DB"))
	app.Run(os.Getenv("PORT"))
}
