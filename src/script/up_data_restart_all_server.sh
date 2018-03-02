set -x
cd ../game_data
svn up
cd ../bin

sh ./kill_all_server.sh

sleep 5s

nohup `pwd`/center_server &
sleep 1s
nohup `pwd`/rpc_server &
sleep 1s
nohup `pwd`/hall_server &
sleep 1s
nohup `pwd`/login_server &


