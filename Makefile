.PHONY: medea
medea:
	@rm -rf medea
	go build -o medea cmd/main.go

clean:
	rm -rf medea
