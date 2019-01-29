.PHONY: run build clean

build: clean
	go build -o vault-k8s-init
	chmod +x vault-k8s-init

clean:
	rm -f ./vault-k8s-init

run: build
	./vault-k8s-init
	rm -f ./vault-k8s-init
