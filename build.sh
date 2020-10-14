echo Creating folders
mkdir -p bin/linux-amd64
#mkdir -p bin/windows-amd64
mkdir -p bin/darwin-amd64
echo Building linux
env GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/memory
#echo Building windows
#env GOOS=windows GOARCH=amd64 go build -o bin/windows-amd64/memory
echo Building darwin
env GOOS=darwin GOARCH=amd64 go build -o bin/darwin-amd64/memory
echo Installing
go install
