set GOPATH=D:\work\wjxz_server

call build_framework.bat
if errorlevel 1 goto exit

go build -o ../main/center_server/center_server_server.exe main/center_server
if errorlevel 1 goto exit

go install main/center_server
if errorlevel 1 goto exit

if errorlevel 0 goto ok

:exit
echo build center_server failed!!!!!!!!!!!!!!!!!!!

:ok
echo build center_server ok