cd../public_message

md gen_go
cd gen_go

md server_message
md db_center
md db_hall

cd ../../tools

move protoc.exe ../public_message
move protoc-gen-go.exe ../public_message

cd ../public_message
protoc.exe --go_out=./gen_go/server_message/ server_message.proto
cd ../_
if errorlevel 1 goto exit

cd ../public_message
go install public_message/gen_go/server_message
cd ../_
if errorlevel 1 goto exit

cd ../public_message
protoc.exe --go_out=./gen_go/db_center/ db_center.proto
cd ../_
if errorlevel 1 goto exit

cd ../public_message
protoc.exe --go_out=./gen_go/db_hall/ db_hallsvr.proto
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