mkdir -p bin

# Linux
echo Linux x64
env GOOS=linux GOARCH=amd64 go build -o bin/seec-linux-x64
echo Linux ARM64
env GOOS=linux GOARCH=arm64 go build -o bin/seec-linux-arm64
echo Linux ARM
env GOOS=linux GOARCH=arm go build -o bin/seec-linux-arm
echo Linux x86
env GOOS=linux GOARCH=386 go build -o bin/seec-linux-x86
echo Linux PowerPC64
env GOOS=linux GOARCH=ppc64 go build -o bin/seec-linux-ppc64
echo Linux RiscV64
env GOOS=linux GOARCH=riscv64 go build -o bin/seec-linux-riscv64

# Windows
echo Windows x64
env GOOS=windows GOARCH=amd64 go build -o bin/seec-windows-x64.exe
echo Windows ARM64
env GOOS=windows GOARCH=arm64 go build -o bin/seec-windows-arm64.exe
echo Windows x86
env GOOS=windows GOARCH=386 go build -o bin/seec-windows-x86.exe

# Darwin (macOS)
echo Darwin x64
env GOOS=darwin GOARCH=amd64 go build -o bin/seec-darwin-x64
echo Darwin ARM64
env GOOS=darwin GOARCH=arm64 go build -o bin/seec-darwin-arm64
