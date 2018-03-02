call build_table_config.bat
go install youma/rpc_common
go install youma/rpc_server
if errorlevel 1 goto exit

if errorlevel 0 goto ok

:exit
echo build rpc_server failed!!!!!!!!!!!!!!!!!!!

:ok
echo build rpc_server ok