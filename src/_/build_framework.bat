call gen_server_message.bat
if errorlevel 1 goto exit

call gen_client_message.bat
if errorlevel 1 goto exit

go install libs/log
if errorlevel 1 goto exit

go install libs/timer
if errorlevel 1 goto exit

go install libs/perf
if errorlevel 1 goto exit

go install libs/socket
if errorlevel 1 goto exit

go install libs/web
if errorlevel 1 goto exit

go install libs/server_conn
if errorlevel 1 goto exit

if errorlevel 0 goto ok

:exit
echo build framework failed!!!!!!!!!!!!!!!!!!

:ok
echo build framework ok