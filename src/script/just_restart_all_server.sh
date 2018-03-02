set -x

sh ./kill_all_server.sh

sleep 1s

nohup `pwd`/center_server &
sleep 1s
nohup `pwd`/login_server &
sleep 1s
nohup `pwd`/hall_server &
sleep 1s
nohup `pwd`/match_server &
sleep 1s
nohup `pwd`/room_server &
