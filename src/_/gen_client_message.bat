cd../public_message

md gen_go
cd gen_go

md client_message

cd ../../tools

move protoc.exe ../public_message
move protoc-gen-go.exe ../public_message

cd ../public_message
protoc.exe --go_out=./gen_go/client_message/ client_message.proto
cd ../_
if errorlevel 1 goto exit

cd ../public_message
go install public_message/gen_go/client_message
cd ../_
if errorlevel 1 goto exit

cd ../public_message
move protoc.exe ../tools
move protoc-gen-go.exe ../tools
cd ../_

goto ok

:exit
echo gen message failed!!!!!!!!!!!!!!!!!!!!!!!!!!!!

:ok
echo gen message ok