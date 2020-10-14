echo Creating folders
mkdir -p bin/linux-amd64
#mkdir -p bin/windows-amd64
mkdir -p bin/darwin-amd64
echo Building linux
env GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/memory
cd bin/linux-amd64
zip memory-linux-amd64.zip memory
cd .. && cd ..
#echo Building windows
#env GOOS=windows GOARCH=amd64 go build -o bin/windows-amd64/memory
echo Building darwin
env GOOS=darwin GOARCH=amd64 go build -o bin/darwin-amd64/memory
cd bin/darwin-amd64
zip memory-darwin-amd64.zip memory
cd .. && cd ..
echo Installing
go install
