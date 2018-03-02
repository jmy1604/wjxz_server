call build_framework.bat
if errorlevel 1 goto exit

call build_table_config.bat
if errorlevel 1 goto exit

go build -o ../youma/hall_server/hall_server.exe youma/hall_server
if errorlevel 1 goto exit

go install youma/rpc_common
go install youma/hall_server
if errorlevel 1 goto exit

if errorlevel 0 goto ok

:exit
echo build hall_server failed!!!!!!!!!!!!!!!!!!!

:ok
echo build hall_server ok