export GOPATH=/root/mm_server
set -x
cd ..
svn up
go install -v -work github.com/garyburd/redigo/internal
go install -v -work github.com/garyburd/redigo/redisx
go install -v -work github.com/garyburd/redigo/redis
go install -v -work youma/table_config
go install -v -work youma/rpc_common
go install -v -work youma/center_server
go install -v -work youma/login_server
go install -v -work youma/hall_server
go install -v -work youma/rpc_server
