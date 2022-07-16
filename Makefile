#
build-sample:
	go build -o dist/hello-world recipes/hello-world.go
run-sample:
	go run recipes/hello-world.go
run-bin:
	./hello-world