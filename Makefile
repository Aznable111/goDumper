

all: mipsbe mipsle arm amd64

mipsbe:
	GOARCH=mips GOOS=linux go build -a -gcflags=all="-l -B" -ldflags="-w -s" -o bins/gd_mipsbe goDumper.go

mipsle:
	GOARCH=mipsle GOOS=linux go build -a -gcflags=all="-l -B" -ldflags="-w -s" -o bins/gd_mipsle goDumper.go

arm:
	GOARCH=arm GOOS=linux go build -a -gcflags=all="-l -B" -ldflags="-w -s" -o bins/gd_arm goDumper.go

amd64:
	GOARCH=amd64 GOOS=linux go build -a -gcflags=all="-l -B" -ldflags="-w -s" -o bins/gd_amd64 goDumper.go
