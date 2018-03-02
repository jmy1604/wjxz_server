go build -o ../youma/test_client/test_client.exe youma/test_client
if errorlevel 1 goto exit

go install youma/test_client
if errorlevel 1 goto exit

if errorlevel 0 goto ok

:exit
echo build test_client failed!!!!!!!!!!!!!!!!!!!

:ok
echo build test_client ok