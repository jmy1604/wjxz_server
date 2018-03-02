call build_framework.bat
if errorlevel 1 goto exit

go build -o ../youma/center_server/center_server_server.exe youma/center_server
if errorlevel 1 goto exit

go install youma/center_server
if errorlevel 1 goto exit

if errorlevel 0 goto ok

:exit
echo build center_server failed!!!!!!!!!!!!!!!!!!!

:ok
echo build center_server ok