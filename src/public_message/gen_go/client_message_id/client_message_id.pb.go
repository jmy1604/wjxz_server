// Code generated by protoc-gen-go. DO NOT EDIT.
// source: client_message_id.proto

/*
Package msg_client_message_id is a generated protocol buffer package.

It is generated from these files:
	client_message_id.proto

It has these top-level messages:
*/
package msg_client_message_id

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type MSGID int32

const (
	MSGID_NONE                                         MSGID = 0
	MSGID_C2S_TEST_COMMAND                             MSGID = 1
	MSGID_C2S_HEARTBEAT                                MSGID = 2
	MSGID_S2C_STATE_NOTIFY                             MSGID = 3
	MSGID_C2S_DATA_SYNC_REQUEST                        MSGID = 4
	MSGID_C2S_LOGIN_REQUEST                            MSGID = 10000
	MSGID_S2C_LOGIN_RESPONSE                           MSGID = 10001
	MSGID_S2C_OTHER_PLACE_LOGIN                        MSGID = 10002
	MSGID_C2S_SELECT_SERVER_REQUEST                    MSGID = 10003
	MSGID_S2C_SELECT_SERVER_RESPONSE                   MSGID = 10004
	MSGID_C2S_ENTER_GAME_REQUEST                       MSGID = 10020
	MSGID_S2C_ENTER_GAME_RESPONSE                      MSGID = 10021
	MSGID_S2C_ENTER_GAME_COMPLETE_NOTIFY               MSGID = 10022
	MSGID_C2S_LEAVE_GAME_REQUEST                       MSGID = 10023
	MSGID_S2C_LEAVE_GAME_RESPONSE                      MSGID = 10024
	MSGID_S2C_PLAYER_INFO_RESPONSE                     MSGID = 10025
	MSGID_C2S_ROLES_REQUEST                            MSGID = 10050
	MSGID_S2C_ROLES_RESPONSE                           MSGID = 10051
	MSGID_S2C_ROLES_CHANGE_NOTIFY                      MSGID = 10052
	MSGID_C2S_ROLE_ATTRS_REQUEST                       MSGID = 10053
	MSGID_S2C_ROLE_ATTRS_RESPONSE                      MSGID = 10054
	MSGID_C2S_ROLE_LEVELUP_REQUEST                     MSGID = 10055
	MSGID_S2C_ROLE_LEVELUP_RESPONSE                    MSGID = 10056
	MSGID_C2S_ROLE_RANKUP_REQUEST                      MSGID = 10057
	MSGID_S2C_ROLE_RANKUP_RESPONSE                     MSGID = 10058
	MSGID_C2S_ROLE_DECOMPOSE_REQUEST                   MSGID = 10059
	MSGID_S2C_ROLE_DECOMPOSE_RESPONSE                  MSGID = 10060
	MSGID_C2S_ROLE_FUSION_REQUEST                      MSGID = 10061
	MSGID_S2C_ROLE_FUSION_RESPONSE                     MSGID = 10062
	MSGID_C2S_ROLE_LOCK_REQUEST                        MSGID = 10063
	MSGID_S2C_ROLE_LOCK_RESPONSE                       MSGID = 10064
	MSGID_C2S_ROLE_HANDBOOK_REQUEST                    MSGID = 10065
	MSGID_S2C_ROLE_HANDBOOK_RESPONSE                   MSGID = 10066
	MSGID_C2S_ROLE_LEFTSLOT_OPEN_REQUEST               MSGID = 10067
	MSGID_S2C_ROLE_LEFTSLOT_OPEN_RESPONSE              MSGID = 10068
	MSGID_C2S_ROLE_ONEKEY_EQUIP_REQUEST                MSGID = 10069
	MSGID_S2C_ROLE_ONEKEY_EQUIP_RESPONSE               MSGID = 10070
	MSGID_C2S_ROLE_ONEKEY_UNEQUIP_REQUEST              MSGID = 10071
	MSGID_S2C_ROLE_ONEKEY_UNEQUIP_RESPONSE             MSGID = 10072
	MSGID_C2S_ROLE_LEFTSLOT_RESULT_SAVE_REQUEST        MSGID = 10073
	MSGID_S2C_ROLE_LEFTSLOT_RESULT_SAVE_RESPONSE       MSGID = 10074
	MSGID_C2S_ROLE_LEFTSLOT_RESULT_CANCEL_REQUEST      MSGID = 10075
	MSGID_S2C_ROLE_LEFTSLOT_RESULT_CANCEL_RESPONSE     MSGID = 10076
	MSGID_C2S_BATTLE_RESULT_REQUEST                    MSGID = 10100
	MSGID_S2C_BATTLE_RESULT_RESPONSE                   MSGID = 10101
	MSGID_C2S_BATTLE_RECORD_REQUEST                    MSGID = 10102
	MSGID_S2C_BATTLE_RECORD_RESPONSE                   MSGID = 10103
	MSGID_C2S_BATTLE_RECORD_LIST_REQUEST               MSGID = 10104
	MSGID_S2C_BATTLE_RECORD_LIST_RESPONSE              MSGID = 10105
	MSGID_C2S_BATTLE_RECORD_DELETE_REQUEST             MSGID = 10106
	MSGID_S2C_BATTLE_RECORD_DELETE_RESPONSE            MSGID = 10107
	MSGID_C2S_SET_TEAM_REQUEST                         MSGID = 10200
	MSGID_S2C_SET_TEAM_RESPONSE                        MSGID = 10201
	MSGID_S2C_TEAMS_RESPONSE                           MSGID = 10202
	MSGID_C2S_ITEMS_SYNC_REQUEST                       MSGID = 10300
	MSGID_S2C_ITEMS_SYNC                               MSGID = 10301
	MSGID_S2C_ITEMS_UPDATE                             MSGID = 10302
	MSGID_C2S_ITEM_FUSION_REQUEST                      MSGID = 10303
	MSGID_S2C_ITEM_FUSION_RESPONSE                     MSGID = 10304
	MSGID_C2S_ITEM_SELL_REQUEST                        MSGID = 10305
	MSGID_S2C_ITEM_SELL_RESPONSE                       MSGID = 10306
	MSGID_C2S_ITEM_EQUIP_REQUEST                       MSGID = 10307
	MSGID_S2C_ITEM_EQUIP_RESPONSE                      MSGID = 10308
	MSGID_C2S_ITEM_UNEQUIP_REQUEST                     MSGID = 10309
	MSGID_S2C_ITEM_UNEQUIP_RESPONSE                    MSGID = 10310
	MSGID_C2S_ITEM_UPGRADE_REQUEST                     MSGID = 10311
	MSGID_S2C_ITEM_UPGRADE_RESPONSE                    MSGID = 10312
	MSGID_C2S_ITEM_ONEKEY_UPGRADE_REQUEST              MSGID = 10313
	MSGID_S2C_ITEM_ONEKEY_UPGRADE_RESPONSE             MSGID = 10314
	MSGID_C2S_CAMPAIGN_DATA_REQUEST                    MSGID = 10400
	MSGID_S2C_CAMPAIGN_DATA_RESPONSE                   MSGID = 10401
	MSGID_C2S_CAMPAIGN_HANGUP_INCOME_REQUEST           MSGID = 10402
	MSGID_S2C_CAMPAIGN_HANGUP_INCOME_RESPONSE          MSGID = 10403
	MSGID_C2S_BATTLE_SET_HANGUP_CAMPAIGN_REQUEST       MSGID = 10404
	MSGID_S2C_BATTLE_SET_HANGUP_CAMPAIGN_RESPONSE      MSGID = 10405
	MSGID_C2S_MAIL_SEND_REQUEST                        MSGID = 10500
	MSGID_S2C_MAIL_SEND_RESPONSE                       MSGID = 10501
	MSGID_C2S_MAIL_LIST_REQUEST                        MSGID = 10502
	MSGID_S2C_MAIL_LIST_RESPONSE                       MSGID = 10503
	MSGID_C2S_MAIL_DETAIL_REQUEST                      MSGID = 10504
	MSGID_S2C_MAIL_DETAIL_RESPONSE                     MSGID = 10505
	MSGID_C2S_MAIL_GET_ATTACHED_ITEMS_REQUEST          MSGID = 10506
	MSGID_S2C_MAIL_GET_ATTACHED_ITEMS_RESPONSE         MSGID = 10507
	MSGID_C2S_MAIL_DELETE_REQUEST                      MSGID = 10508
	MSGID_S2C_MAIL_DELETE_RESPONSE                     MSGID = 10509
	MSGID_S2C_MAILS_NEW_NOTIFY                         MSGID = 10510
	MSGID_C2S_TALENT_UP_REQUEST                        MSGID = 10600
	MSGID_S2C_TALENT_UP_RESPONSE                       MSGID = 10601
	MSGID_C2S_TALENT_LIST_REQUEST                      MSGID = 10602
	MSGID_S2C_TALENT_LIST_RESPONSE                     MSGID = 10603
	MSGID_C2S_TOWER_DATA_REQUEST                       MSGID = 10700
	MSGID_S2C_TOWER_DATA_RESPONSE                      MSGID = 10701
	MSGID_C2S_TOWER_RECORDS_INFO_REQUEST               MSGID = 10702
	MSGID_S2C_TOWER_RECORDS_INFO_RESPONSE              MSGID = 10703
	MSGID_C2S_TOWER_RECORD_DATA_REQUEST                MSGID = 10704
	MSGID_S2C_TOWER_RECORD_DATA_RESPONSE               MSGID = 10705
	MSGID_C2S_TOWER_RANKING_LIST_REQUEST               MSGID = 10706
	MSGID_S2C_TOWER_RANKING_LIST_RESPONSE              MSGID = 10707
	MSGID_C2S_DRAW_CARD_REQUEST                        MSGID = 10800
	MSGID_S2C_DRAW_CARD_RESPONSE                       MSGID = 10801
	MSGID_C2S_DRAW_DATA_REQUEST                        MSGID = 10802
	MSGID_S2C_DRAW_DATA_RESPONSE                       MSGID = 10803
	MSGID_C2S_TOUCH_GOLD_REQUEST                       MSGID = 10900
	MSGID_S2C_TOUCH_GOLD_RESPONSE                      MSGID = 10901
	MSGID_C2S_GOLD_HAND_DATA_REQUEST                   MSGID = 10902
	MSGID_S2C_GOLD_HAND_DATA_RESPONSE                  MSGID = 10903
	MSGID_C2S_SHOP_DATA_REQUEST                        MSGID = 11000
	MSGID_S2C_SHOP_DATA_RESPONSE                       MSGID = 11001
	MSGID_C2S_SHOP_BUY_ITEM_REQUEST                    MSGID = 11002
	MSGID_S2C_SHOP_BUY_ITEM_RESPONSE                   MSGID = 11003
	MSGID_C2S_SHOP_REFRESH_REQUEST                     MSGID = 11004
	MSGID_S2C_SHOP_REFRESH_RESPONSE                    MSGID = 11005
	MSGID_S2C_SHOP_AUTO_REFRESH_NOTIFY                 MSGID = 11006
	MSGID_C2S_RANK_LIST_REQUEST                        MSGID = 11100
	MSGID_S2C_RANK_LIST_RESPONSE                       MSGID = 11101
	MSGID_C2S_ARENA_DATA_REQUEST                       MSGID = 11200
	MSGID_S2C_ARENA_DATA_RESPONSE                      MSGID = 11201
	MSGID_C2S_ARENA_PLAYER_DEFENSE_TEAM_REQUEST        MSGID = 11202
	MSGID_S2C_ARENA_PLAYER_DEFENSE_TEAM_RESPONSE       MSGID = 11203
	MSGID_C2S_ARENA_MATCH_PLAYER_REQUEST               MSGID = 11204
	MSGID_S2C_ARENA_MATCH_PLAYER_RESPONSE              MSGID = 11205
	MSGID_S2C_ARENA_GRADE_REWARD_NOTIFY                MSGID = 11206
	MSGID_C2S_ACTIVE_STAGE_DATA_REQUEST                MSGID = 11300
	MSGID_S2C_ACTIVE_STAGE_DATA_RESPONSE               MSGID = 11301
	MSGID_C2S_ACTIVE_STAGE_BUY_CHALLENGE_NUM_REQUEST   MSGID = 11302
	MSGID_S2C_ACTIVE_STAGE_BUY_CHALLENGE_NUM_RESPONSE  MSGID = 11303
	MSGID_S2C_ACTIVE_STAGE_REFRESH_NOTIFY              MSGID = 11304
	MSGID_C2S_ACTIVE_STAGE_SELECT_ASSIST_ROLE_REQUEST  MSGID = 11305
	MSGID_S2C_ACTIVE_STAGE_SELECT_ASSIST_ROLE_RESPONSE MSGID = 11306
	MSGID_C2S_ACTIVE_STAGE_ASSIST_ROLE_LIST_REQUEST    MSGID = 11307
	MSGID_S2C_ACTIVE_STAGE_ASSIST_ROLE_LIST_RESPONSE   MSGID = 11308
	MSGID_C2S_FRIEND_RECOMMEND_REQUEST                 MSGID = 11400
	MSGID_S2C_FRIEND_RECOMMEND_RESPONSE                MSGID = 11401
	MSGID_C2S_FRIEND_LIST_REQUEST                      MSGID = 11402
	MSGID_S2C_FRIEND_LIST_RESPONSE                     MSGID = 11403
	MSGID_C2S_FRIEND_ASK_REQUEST                       MSGID = 11404
	MSGID_S2C_FRIEND_ASK_RESPONSE                      MSGID = 11405
	MSGID_C2S_FRIEND_ASK_PLAYER_LIST_REQUEST           MSGID = 11406
	MSGID_S2C_FRIEND_ASK_PLAYER_LIST_RESPONSE          MSGID = 11407
	MSGID_S2C_FRIEND_ASK_PLAYER_LIST_ADD_NOTIFY        MSGID = 11408
	MSGID_C2S_FRIEND_AGREE_REQUEST                     MSGID = 11409
	MSGID_S2C_FRIEND_AGREE_RESPONSE                    MSGID = 11410
	MSGID_S2C_FRIEND_LIST_ADD_NOTIFY                   MSGID = 11411
	MSGID_C2S_FRIEND_REFUSE_REQUEST                    MSGID = 11412
	MSGID_S2C_FRIEND_REFUSE_RESPONSE                   MSGID = 11413
	MSGID_C2S_FRIEND_REMOVE_REQUEST                    MSGID = 11414
	MSGID_S2C_FRIEND_REMOVE_RESPONSE                   MSGID = 11415
	MSGID_C2S_FRIEND_GIVE_POINTS_REQUEST               MSGID = 11416
	MSGID_S2C_FRIEND_GIVE_POINTS_RESPONSE              MSGID = 11417
	MSGID_C2S_FRIEND_GET_POINTS_REQUEST                MSGID = 11418
	MSGID_S2C_FRIEND_GET_POINTS_RESPONSE               MSGID = 11419
	MSGID_C2S_FRIEND_SEARCH_BOSS_REQUEST               MSGID = 11420
	MSGID_S2C_FRIEND_SEARCH_BOSS_RESPONSE              MSGID = 11421
	MSGID_C2S_FRIENDS_BOSS_LIST_REQUEST                MSGID = 11422
	MSGID_S2C_FRIENDS_BOSS_LIST_RESPONSE               MSGID = 11423
	MSGID_C2S_FRIEND_BOSS_ATTACK_LIST_REQUEST          MSGID = 11424
	MSGID_S2C_FRIEND_BOSS_ATTACK_LIST_RESPONSE         MSGID = 11425
	MSGID_C2S_FRIEND_DATA_REQUEST                      MSGID = 11426
	MSGID_S2C_FRIEND_DATA_RESPONSE                     MSGID = 11427
	MSGID_S2C_FRIEND_BOSS_ATTACK_REWARD_NOTIFY         MSGID = 11428
	MSGID_C2S_FRIEND_SET_ASSIST_ROLE_REQUEST           MSGID = 11429
	MSGID_S2C_FRIEND_SET_ASSIST_ROLE_RESPONSE          MSGID = 11430
)

var MSGID_name = map[int32]string{
	0:     "NONE",
	1:     "C2S_TEST_COMMAND",
	2:     "C2S_HEARTBEAT",
	3:     "S2C_STATE_NOTIFY",
	4:     "C2S_DATA_SYNC_REQUEST",
	10000: "C2S_LOGIN_REQUEST",
	10001: "S2C_LOGIN_RESPONSE",
	10002: "S2C_OTHER_PLACE_LOGIN",
	10003: "C2S_SELECT_SERVER_REQUEST",
	10004: "S2C_SELECT_SERVER_RESPONSE",
	10020: "C2S_ENTER_GAME_REQUEST",
	10021: "S2C_ENTER_GAME_RESPONSE",
	10022: "S2C_ENTER_GAME_COMPLETE_NOTIFY",
	10023: "C2S_LEAVE_GAME_REQUEST",
	10024: "S2C_LEAVE_GAME_RESPONSE",
	10025: "S2C_PLAYER_INFO_RESPONSE",
	10050: "C2S_ROLES_REQUEST",
	10051: "S2C_ROLES_RESPONSE",
	10052: "S2C_ROLES_CHANGE_NOTIFY",
	10053: "C2S_ROLE_ATTRS_REQUEST",
	10054: "S2C_ROLE_ATTRS_RESPONSE",
	10055: "C2S_ROLE_LEVELUP_REQUEST",
	10056: "S2C_ROLE_LEVELUP_RESPONSE",
	10057: "C2S_ROLE_RANKUP_REQUEST",
	10058: "S2C_ROLE_RANKUP_RESPONSE",
	10059: "C2S_ROLE_DECOMPOSE_REQUEST",
	10060: "S2C_ROLE_DECOMPOSE_RESPONSE",
	10061: "C2S_ROLE_FUSION_REQUEST",
	10062: "S2C_ROLE_FUSION_RESPONSE",
	10063: "C2S_ROLE_LOCK_REQUEST",
	10064: "S2C_ROLE_LOCK_RESPONSE",
	10065: "C2S_ROLE_HANDBOOK_REQUEST",
	10066: "S2C_ROLE_HANDBOOK_RESPONSE",
	10067: "C2S_ROLE_LEFTSLOT_OPEN_REQUEST",
	10068: "S2C_ROLE_LEFTSLOT_OPEN_RESPONSE",
	10069: "C2S_ROLE_ONEKEY_EQUIP_REQUEST",
	10070: "S2C_ROLE_ONEKEY_EQUIP_RESPONSE",
	10071: "C2S_ROLE_ONEKEY_UNEQUIP_REQUEST",
	10072: "S2C_ROLE_ONEKEY_UNEQUIP_RESPONSE",
	10073: "C2S_ROLE_LEFTSLOT_RESULT_SAVE_REQUEST",
	10074: "S2C_ROLE_LEFTSLOT_RESULT_SAVE_RESPONSE",
	10075: "C2S_ROLE_LEFTSLOT_RESULT_CANCEL_REQUEST",
	10076: "S2C_ROLE_LEFTSLOT_RESULT_CANCEL_RESPONSE",
	10100: "C2S_BATTLE_RESULT_REQUEST",
	10101: "S2C_BATTLE_RESULT_RESPONSE",
	10102: "C2S_BATTLE_RECORD_REQUEST",
	10103: "S2C_BATTLE_RECORD_RESPONSE",
	10104: "C2S_BATTLE_RECORD_LIST_REQUEST",
	10105: "S2C_BATTLE_RECORD_LIST_RESPONSE",
	10106: "C2S_BATTLE_RECORD_DELETE_REQUEST",
	10107: "S2C_BATTLE_RECORD_DELETE_RESPONSE",
	10200: "C2S_SET_TEAM_REQUEST",
	10201: "S2C_SET_TEAM_RESPONSE",
	10202: "S2C_TEAMS_RESPONSE",
	10300: "C2S_ITEMS_SYNC_REQUEST",
	10301: "S2C_ITEMS_SYNC",
	10302: "S2C_ITEMS_UPDATE",
	10303: "C2S_ITEM_FUSION_REQUEST",
	10304: "S2C_ITEM_FUSION_RESPONSE",
	10305: "C2S_ITEM_SELL_REQUEST",
	10306: "S2C_ITEM_SELL_RESPONSE",
	10307: "C2S_ITEM_EQUIP_REQUEST",
	10308: "S2C_ITEM_EQUIP_RESPONSE",
	10309: "C2S_ITEM_UNEQUIP_REQUEST",
	10310: "S2C_ITEM_UNEQUIP_RESPONSE",
	10311: "C2S_ITEM_UPGRADE_REQUEST",
	10312: "S2C_ITEM_UPGRADE_RESPONSE",
	10313: "C2S_ITEM_ONEKEY_UPGRADE_REQUEST",
	10314: "S2C_ITEM_ONEKEY_UPGRADE_RESPONSE",
	10400: "C2S_CAMPAIGN_DATA_REQUEST",
	10401: "S2C_CAMPAIGN_DATA_RESPONSE",
	10402: "C2S_CAMPAIGN_HANGUP_INCOME_REQUEST",
	10403: "S2C_CAMPAIGN_HANGUP_INCOME_RESPONSE",
	10404: "C2S_BATTLE_SET_HANGUP_CAMPAIGN_REQUEST",
	10405: "S2C_BATTLE_SET_HANGUP_CAMPAIGN_RESPONSE",
	10500: "C2S_MAIL_SEND_REQUEST",
	10501: "S2C_MAIL_SEND_RESPONSE",
	10502: "C2S_MAIL_LIST_REQUEST",
	10503: "S2C_MAIL_LIST_RESPONSE",
	10504: "C2S_MAIL_DETAIL_REQUEST",
	10505: "S2C_MAIL_DETAIL_RESPONSE",
	10506: "C2S_MAIL_GET_ATTACHED_ITEMS_REQUEST",
	10507: "S2C_MAIL_GET_ATTACHED_ITEMS_RESPONSE",
	10508: "C2S_MAIL_DELETE_REQUEST",
	10509: "S2C_MAIL_DELETE_RESPONSE",
	10510: "S2C_MAILS_NEW_NOTIFY",
	10600: "C2S_TALENT_UP_REQUEST",
	10601: "S2C_TALENT_UP_RESPONSE",
	10602: "C2S_TALENT_LIST_REQUEST",
	10603: "S2C_TALENT_LIST_RESPONSE",
	10700: "C2S_TOWER_DATA_REQUEST",
	10701: "S2C_TOWER_DATA_RESPONSE",
	10702: "C2S_TOWER_RECORDS_INFO_REQUEST",
	10703: "S2C_TOWER_RECORDS_INFO_RESPONSE",
	10704: "C2S_TOWER_RECORD_DATA_REQUEST",
	10705: "S2C_TOWER_RECORD_DATA_RESPONSE",
	10706: "C2S_TOWER_RANKING_LIST_REQUEST",
	10707: "S2C_TOWER_RANKING_LIST_RESPONSE",
	10800: "C2S_DRAW_CARD_REQUEST",
	10801: "S2C_DRAW_CARD_RESPONSE",
	10802: "C2S_DRAW_DATA_REQUEST",
	10803: "S2C_DRAW_DATA_RESPONSE",
	10900: "C2S_TOUCH_GOLD_REQUEST",
	10901: "S2C_TOUCH_GOLD_RESPONSE",
	10902: "C2S_GOLD_HAND_DATA_REQUEST",
	10903: "S2C_GOLD_HAND_DATA_RESPONSE",
	11000: "C2S_SHOP_DATA_REQUEST",
	11001: "S2C_SHOP_DATA_RESPONSE",
	11002: "C2S_SHOP_BUY_ITEM_REQUEST",
	11003: "S2C_SHOP_BUY_ITEM_RESPONSE",
	11004: "C2S_SHOP_REFRESH_REQUEST",
	11005: "S2C_SHOP_REFRESH_RESPONSE",
	11006: "S2C_SHOP_AUTO_REFRESH_NOTIFY",
	11100: "C2S_RANK_LIST_REQUEST",
	11101: "S2C_RANK_LIST_RESPONSE",
	11200: "C2S_ARENA_DATA_REQUEST",
	11201: "S2C_ARENA_DATA_RESPONSE",
	11202: "C2S_ARENA_PLAYER_DEFENSE_TEAM_REQUEST",
	11203: "S2C_ARENA_PLAYER_DEFENSE_TEAM_RESPONSE",
	11204: "C2S_ARENA_MATCH_PLAYER_REQUEST",
	11205: "S2C_ARENA_MATCH_PLAYER_RESPONSE",
	11206: "S2C_ARENA_GRADE_REWARD_NOTIFY",
	11300: "C2S_ACTIVE_STAGE_DATA_REQUEST",
	11301: "S2C_ACTIVE_STAGE_DATA_RESPONSE",
	11302: "C2S_ACTIVE_STAGE_BUY_CHALLENGE_NUM_REQUEST",
	11303: "S2C_ACTIVE_STAGE_BUY_CHALLENGE_NUM_RESPONSE",
	11304: "S2C_ACTIVE_STAGE_REFRESH_NOTIFY",
	11305: "C2S_ACTIVE_STAGE_SELECT_ASSIST_ROLE_REQUEST",
	11306: "S2C_ACTIVE_STAGE_SELECT_ASSIST_ROLE_RESPONSE",
	11307: "C2S_ACTIVE_STAGE_ASSIST_ROLE_LIST_REQUEST",
	11308: "S2C_ACTIVE_STAGE_ASSIST_ROLE_LIST_RESPONSE",
	11400: "C2S_FRIEND_RECOMMEND_REQUEST",
	11401: "S2C_FRIEND_RECOMMEND_RESPONSE",
	11402: "C2S_FRIEND_LIST_REQUEST",
	11403: "S2C_FRIEND_LIST_RESPONSE",
	11404: "C2S_FRIEND_ASK_REQUEST",
	11405: "S2C_FRIEND_ASK_RESPONSE",
	11406: "C2S_FRIEND_ASK_PLAYER_LIST_REQUEST",
	11407: "S2C_FRIEND_ASK_PLAYER_LIST_RESPONSE",
	11408: "S2C_FRIEND_ASK_PLAYER_LIST_ADD_NOTIFY",
	11409: "C2S_FRIEND_AGREE_REQUEST",
	11410: "S2C_FRIEND_AGREE_RESPONSE",
	11411: "S2C_FRIEND_LIST_ADD_NOTIFY",
	11412: "C2S_FRIEND_REFUSE_REQUEST",
	11413: "S2C_FRIEND_REFUSE_RESPONSE",
	11414: "C2S_FRIEND_REMOVE_REQUEST",
	11415: "S2C_FRIEND_REMOVE_RESPONSE",
	11416: "C2S_FRIEND_GIVE_POINTS_REQUEST",
	11417: "S2C_FRIEND_GIVE_POINTS_RESPONSE",
	11418: "C2S_FRIEND_GET_POINTS_REQUEST",
	11419: "S2C_FRIEND_GET_POINTS_RESPONSE",
	11420: "C2S_FRIEND_SEARCH_BOSS_REQUEST",
	11421: "S2C_FRIEND_SEARCH_BOSS_RESPONSE",
	11422: "C2S_FRIENDS_BOSS_LIST_REQUEST",
	11423: "S2C_FRIENDS_BOSS_LIST_RESPONSE",
	11424: "C2S_FRIEND_BOSS_ATTACK_LIST_REQUEST",
	11425: "S2C_FRIEND_BOSS_ATTACK_LIST_RESPONSE",
	11426: "C2S_FRIEND_DATA_REQUEST",
	11427: "S2C_FRIEND_DATA_RESPONSE",
	11428: "S2C_FRIEND_BOSS_ATTACK_REWARD_NOTIFY",
	11429: "C2S_FRIEND_SET_ASSIST_ROLE_REQUEST",
	11430: "S2C_FRIEND_SET_ASSIST_ROLE_RESPONSE",
}
var MSGID_value = map[string]int32{
	"NONE":                                         0,
	"C2S_TEST_COMMAND":                             1,
	"C2S_HEARTBEAT":                                2,
	"S2C_STATE_NOTIFY":                             3,
	"C2S_DATA_SYNC_REQUEST":                        4,
	"C2S_LOGIN_REQUEST":                            10000,
	"S2C_LOGIN_RESPONSE":                           10001,
	"S2C_OTHER_PLACE_LOGIN":                        10002,
	"C2S_SELECT_SERVER_REQUEST":                    10003,
	"S2C_SELECT_SERVER_RESPONSE":                   10004,
	"C2S_ENTER_GAME_REQUEST":                       10020,
	"S2C_ENTER_GAME_RESPONSE":                      10021,
	"S2C_ENTER_GAME_COMPLETE_NOTIFY":               10022,
	"C2S_LEAVE_GAME_REQUEST":                       10023,
	"S2C_LEAVE_GAME_RESPONSE":                      10024,
	"S2C_PLAYER_INFO_RESPONSE":                     10025,
	"C2S_ROLES_REQUEST":                            10050,
	"S2C_ROLES_RESPONSE":                           10051,
	"S2C_ROLES_CHANGE_NOTIFY":                      10052,
	"C2S_ROLE_ATTRS_REQUEST":                       10053,
	"S2C_ROLE_ATTRS_RESPONSE":                      10054,
	"C2S_ROLE_LEVELUP_REQUEST":                     10055,
	"S2C_ROLE_LEVELUP_RESPONSE":                    10056,
	"C2S_ROLE_RANKUP_REQUEST":                      10057,
	"S2C_ROLE_RANKUP_RESPONSE":                     10058,
	"C2S_ROLE_DECOMPOSE_REQUEST":                   10059,
	"S2C_ROLE_DECOMPOSE_RESPONSE":                  10060,
	"C2S_ROLE_FUSION_REQUEST":                      10061,
	"S2C_ROLE_FUSION_RESPONSE":                     10062,
	"C2S_ROLE_LOCK_REQUEST":                        10063,
	"S2C_ROLE_LOCK_RESPONSE":                       10064,
	"C2S_ROLE_HANDBOOK_REQUEST":                    10065,
	"S2C_ROLE_HANDBOOK_RESPONSE":                   10066,
	"C2S_ROLE_LEFTSLOT_OPEN_REQUEST":               10067,
	"S2C_ROLE_LEFTSLOT_OPEN_RESPONSE":              10068,
	"C2S_ROLE_ONEKEY_EQUIP_REQUEST":                10069,
	"S2C_ROLE_ONEKEY_EQUIP_RESPONSE":               10070,
	"C2S_ROLE_ONEKEY_UNEQUIP_REQUEST":              10071,
	"S2C_ROLE_ONEKEY_UNEQUIP_RESPONSE":             10072,
	"C2S_ROLE_LEFTSLOT_RESULT_SAVE_REQUEST":        10073,
	"S2C_ROLE_LEFTSLOT_RESULT_SAVE_RESPONSE":       10074,
	"C2S_ROLE_LEFTSLOT_RESULT_CANCEL_REQUEST":      10075,
	"S2C_ROLE_LEFTSLOT_RESULT_CANCEL_RESPONSE":     10076,
	"C2S_BATTLE_RESULT_REQUEST":                    10100,
	"S2C_BATTLE_RESULT_RESPONSE":                   10101,
	"C2S_BATTLE_RECORD_REQUEST":                    10102,
	"S2C_BATTLE_RECORD_RESPONSE":                   10103,
	"C2S_BATTLE_RECORD_LIST_REQUEST":               10104,
	"S2C_BATTLE_RECORD_LIST_RESPONSE":              10105,
	"C2S_BATTLE_RECORD_DELETE_REQUEST":             10106,
	"S2C_BATTLE_RECORD_DELETE_RESPONSE":            10107,
	"C2S_SET_TEAM_REQUEST":                         10200,
	"S2C_SET_TEAM_RESPONSE":                        10201,
	"S2C_TEAMS_RESPONSE":                           10202,
	"C2S_ITEMS_SYNC_REQUEST":                       10300,
	"S2C_ITEMS_SYNC":                               10301,
	"S2C_ITEMS_UPDATE":                             10302,
	"C2S_ITEM_FUSION_REQUEST":                      10303,
	"S2C_ITEM_FUSION_RESPONSE":                     10304,
	"C2S_ITEM_SELL_REQUEST":                        10305,
	"S2C_ITEM_SELL_RESPONSE":                       10306,
	"C2S_ITEM_EQUIP_REQUEST":                       10307,
	"S2C_ITEM_EQUIP_RESPONSE":                      10308,
	"C2S_ITEM_UNEQUIP_REQUEST":                     10309,
	"S2C_ITEM_UNEQUIP_RESPONSE":                    10310,
	"C2S_ITEM_UPGRADE_REQUEST":                     10311,
	"S2C_ITEM_UPGRADE_RESPONSE":                    10312,
	"C2S_ITEM_ONEKEY_UPGRADE_REQUEST":              10313,
	"S2C_ITEM_ONEKEY_UPGRADE_RESPONSE":             10314,
	"C2S_CAMPAIGN_DATA_REQUEST":                    10400,
	"S2C_CAMPAIGN_DATA_RESPONSE":                   10401,
	"C2S_CAMPAIGN_HANGUP_INCOME_REQUEST":           10402,
	"S2C_CAMPAIGN_HANGUP_INCOME_RESPONSE":          10403,
	"C2S_BATTLE_SET_HANGUP_CAMPAIGN_REQUEST":       10404,
	"S2C_BATTLE_SET_HANGUP_CAMPAIGN_RESPONSE":      10405,
	"C2S_MAIL_SEND_REQUEST":                        10500,
	"S2C_MAIL_SEND_RESPONSE":                       10501,
	"C2S_MAIL_LIST_REQUEST":                        10502,
	"S2C_MAIL_LIST_RESPONSE":                       10503,
	"C2S_MAIL_DETAIL_REQUEST":                      10504,
	"S2C_MAIL_DETAIL_RESPONSE":                     10505,
	"C2S_MAIL_GET_ATTACHED_ITEMS_REQUEST":          10506,
	"S2C_MAIL_GET_ATTACHED_ITEMS_RESPONSE":         10507,
	"C2S_MAIL_DELETE_REQUEST":                      10508,
	"S2C_MAIL_DELETE_RESPONSE":                     10509,
	"S2C_MAILS_NEW_NOTIFY":                         10510,
	"C2S_TALENT_UP_REQUEST":                        10600,
	"S2C_TALENT_UP_RESPONSE":                       10601,
	"C2S_TALENT_LIST_REQUEST":                      10602,
	"S2C_TALENT_LIST_RESPONSE":                     10603,
	"C2S_TOWER_DATA_REQUEST":                       10700,
	"S2C_TOWER_DATA_RESPONSE":                      10701,
	"C2S_TOWER_RECORDS_INFO_REQUEST":               10702,
	"S2C_TOWER_RECORDS_INFO_RESPONSE":              10703,
	"C2S_TOWER_RECORD_DATA_REQUEST":                10704,
	"S2C_TOWER_RECORD_DATA_RESPONSE":               10705,
	"C2S_TOWER_RANKING_LIST_REQUEST":               10706,
	"S2C_TOWER_RANKING_LIST_RESPONSE":              10707,
	"C2S_DRAW_CARD_REQUEST":                        10800,
	"S2C_DRAW_CARD_RESPONSE":                       10801,
	"C2S_DRAW_DATA_REQUEST":                        10802,
	"S2C_DRAW_DATA_RESPONSE":                       10803,
	"C2S_TOUCH_GOLD_REQUEST":                       10900,
	"S2C_TOUCH_GOLD_RESPONSE":                      10901,
	"C2S_GOLD_HAND_DATA_REQUEST":                   10902,
	"S2C_GOLD_HAND_DATA_RESPONSE":                  10903,
	"C2S_SHOP_DATA_REQUEST":                        11000,
	"S2C_SHOP_DATA_RESPONSE":                       11001,
	"C2S_SHOP_BUY_ITEM_REQUEST":                    11002,
	"S2C_SHOP_BUY_ITEM_RESPONSE":                   11003,
	"C2S_SHOP_REFRESH_REQUEST":                     11004,
	"S2C_SHOP_REFRESH_RESPONSE":                    11005,
	"S2C_SHOP_AUTO_REFRESH_NOTIFY":                 11006,
	"C2S_RANK_LIST_REQUEST":                        11100,
	"S2C_RANK_LIST_RESPONSE":                       11101,
	"C2S_ARENA_DATA_REQUEST":                       11200,
	"S2C_ARENA_DATA_RESPONSE":                      11201,
	"C2S_ARENA_PLAYER_DEFENSE_TEAM_REQUEST":        11202,
	"S2C_ARENA_PLAYER_DEFENSE_TEAM_RESPONSE":       11203,
	"C2S_ARENA_MATCH_PLAYER_REQUEST":               11204,
	"S2C_ARENA_MATCH_PLAYER_RESPONSE":              11205,
	"S2C_ARENA_GRADE_REWARD_NOTIFY":                11206,
	"C2S_ACTIVE_STAGE_DATA_REQUEST":                11300,
	"S2C_ACTIVE_STAGE_DATA_RESPONSE":               11301,
	"C2S_ACTIVE_STAGE_BUY_CHALLENGE_NUM_REQUEST":   11302,
	"S2C_ACTIVE_STAGE_BUY_CHALLENGE_NUM_RESPONSE":  11303,
	"S2C_ACTIVE_STAGE_REFRESH_NOTIFY":              11304,
	"C2S_ACTIVE_STAGE_SELECT_ASSIST_ROLE_REQUEST":  11305,
	"S2C_ACTIVE_STAGE_SELECT_ASSIST_ROLE_RESPONSE": 11306,
	"C2S_ACTIVE_STAGE_ASSIST_ROLE_LIST_REQUEST":    11307,
	"S2C_ACTIVE_STAGE_ASSIST_ROLE_LIST_RESPONSE":   11308,
	"C2S_FRIEND_RECOMMEND_REQUEST":                 11400,
	"S2C_FRIEND_RECOMMEND_RESPONSE":                11401,
	"C2S_FRIEND_LIST_REQUEST":                      11402,
	"S2C_FRIEND_LIST_RESPONSE":                     11403,
	"C2S_FRIEND_ASK_REQUEST":                       11404,
	"S2C_FRIEND_ASK_RESPONSE":                      11405,
	"C2S_FRIEND_ASK_PLAYER_LIST_REQUEST":           11406,
	"S2C_FRIEND_ASK_PLAYER_LIST_RESPONSE":          11407,
	"S2C_FRIEND_ASK_PLAYER_LIST_ADD_NOTIFY":        11408,
	"C2S_FRIEND_AGREE_REQUEST":                     11409,
	"S2C_FRIEND_AGREE_RESPONSE":                    11410,
	"S2C_FRIEND_LIST_ADD_NOTIFY":                   11411,
	"C2S_FRIEND_REFUSE_REQUEST":                    11412,
	"S2C_FRIEND_REFUSE_RESPONSE":                   11413,
	"C2S_FRIEND_REMOVE_REQUEST":                    11414,
	"S2C_FRIEND_REMOVE_RESPONSE":                   11415,
	"C2S_FRIEND_GIVE_POINTS_REQUEST":               11416,
	"S2C_FRIEND_GIVE_POINTS_RESPONSE":              11417,
	"C2S_FRIEND_GET_POINTS_REQUEST":                11418,
	"S2C_FRIEND_GET_POINTS_RESPONSE":               11419,
	"C2S_FRIEND_SEARCH_BOSS_REQUEST":               11420,
	"S2C_FRIEND_SEARCH_BOSS_RESPONSE":              11421,
	"C2S_FRIENDS_BOSS_LIST_REQUEST":                11422,
	"S2C_FRIENDS_BOSS_LIST_RESPONSE":               11423,
	"C2S_FRIEND_BOSS_ATTACK_LIST_REQUEST":          11424,
	"S2C_FRIEND_BOSS_ATTACK_LIST_RESPONSE":         11425,
	"C2S_FRIEND_DATA_REQUEST":                      11426,
	"S2C_FRIEND_DATA_RESPONSE":                     11427,
	"S2C_FRIEND_BOSS_ATTACK_REWARD_NOTIFY":         11428,
	"C2S_FRIEND_SET_ASSIST_ROLE_REQUEST":           11429,
	"S2C_FRIEND_SET_ASSIST_ROLE_RESPONSE":          11430,
}

func (x MSGID) String() string {
	return proto.EnumName(MSGID_name, int32(x))
}
func (MSGID) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func init() {
	proto.RegisterEnum("msg.client_message_id.MSGID", MSGID_name, MSGID_value)
}

func init() { proto.RegisterFile("client_message_id.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 1717 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x7c, 0x58, 0x57, 0x93, 0x15, 0x45,
	0x14, 0x36, 0xa0, 0x65, 0x75, 0x95, 0xd6, 0xa5, 0x65, 0x41, 0x72, 0x4e, 0x0b, 0xa2, 0xe2, 0x2f,
	0xe8, 0x9d, 0xdb, 0x37, 0xd4, 0xce, 0xed, 0x99, 0x3b, 0xdd, 0xb3, 0x77, 0xe7, 0x69, 0xca, 0x40,
	0x51, 0x54, 0x89, 0x58, 0xc2, 0x4f, 0x30, 0x20, 0x22, 0x92, 0xc4, 0xac, 0x48, 0x46, 0x7d, 0x30,
	0xbc, 0x1a, 0xc8, 0x88, 0x92, 0x31, 0x00, 0xe2, 0xbb, 0xe1, 0xd5, 0x80, 0xb1, 0xac, 0x9e, 0x7b,
	0x4e, 0xcf, 0x74, 0xcf, 0x2c, 0x4f, 0x5b, 0x35, 0x5f, 0xe8, 0x73, 0xbb, 0xcf, 0x39, 0x7d, 0x7a,
	0xc9, 0xa4, 0x47, 0x1e, 0x5b, 0xb5, 0xe2, 0xf1, 0x75, 0xe9, 0xea, 0x15, 0x6b, 0xd7, 0x3e, 0xb4,
	0x72, 0x45, 0xba, 0xea, 0xd1, 0x65, 0x4f, 0x3c, 0xb9, 0x66, 0xdd, 0x1a, 0x3a, 0xb0, 0x7a, 0xed,
	0xca, 0x65, 0x25, 0x70, 0xf0, 0xec, 0x20, 0xb9, 0xad, 0x23, 0x9b, 0xed, 0x3a, 0xbd, 0x83, 0x8c,
	0x13, 0x81, 0xe0, 0xb5, 0x9b, 0xe8, 0x04, 0x52, 0xf3, 0x96, 0xcb, 0x54, 0x71, 0xa9, 0x52, 0x2f,
	0xe8, 0x74, 0x98, 0xa8, 0xd7, 0x6e, 0xa6, 0xe3, 0xc9, 0x9d, 0xfa, 0x6b, 0x8b, 0xb3, 0x48, 0x0d,
	0x71, 0xa6, 0x6a, 0xb7, 0x68, 0xa2, 0x5c, 0xee, 0xa5, 0x52, 0x31, 0xc5, 0x53, 0x11, 0xa8, 0x76,
	0x23, 0xa9, 0xdd, 0x4a, 0x27, 0x93, 0x01, 0x4d, 0xac, 0x33, 0xc5, 0x52, 0x99, 0x08, 0x2f, 0x8d,
	0x78, 0x37, 0xe6, 0x52, 0xd5, 0xc6, 0xd1, 0x89, 0x64, 0xbc, 0x86, 0xfc, 0xa0, 0xd9, 0x16, 0xe6,
	0xf3, 0x66, 0x41, 0x27, 0x11, 0xaa, 0x8d, 0xf0, 0xbb, 0x0c, 0x03, 0x21, 0x79, 0x6d, 0x8b, 0xa0,
	0x53, 0xc8, 0x80, 0x06, 0x02, 0xd5, 0xe2, 0x51, 0x1a, 0xfa, 0xcc, 0xe3, 0x7d, 0x52, 0x6d, 0xab,
	0xa0, 0x33, 0xc8, 0x64, 0x6d, 0x26, 0xb9, 0xcf, 0x3d, 0x95, 0x4a, 0x1e, 0x8d, 0xf0, 0xc8, 0x98,
	0x6e, 0x13, 0x74, 0x26, 0x99, 0x92, 0x45, 0xe7, 0xe0, 0x60, 0xbe, 0x5d, 0xd0, 0xa9, 0x64, 0xa2,
	0x36, 0xe0, 0x42, 0xf1, 0x28, 0x6d, 0xb2, 0x0e, 0x37, 0xea, 0x3d, 0x82, 0x4e, 0x23, 0x93, 0xb4,
	0xda, 0x02, 0x41, 0xba, 0x57, 0xd0, 0xb9, 0x64, 0x86, 0x83, 0x7a, 0x41, 0x27, 0xf4, 0x79, 0xbe,
	0x0f, 0xfb, 0x8c, 0xbf, 0xcf, 0xd9, 0x08, 0xb7, 0xfd, 0xf7, 0x1b, 0x7f, 0x0b, 0x04, 0xff, 0x03,
	0x82, 0x4e, 0x27, 0xf7, 0x68, 0x34, 0xf4, 0x59, 0xc2, 0xa3, 0xb4, 0x2d, 0x1a, 0x41, 0x0e, 0x1f,
	0x14, 0xb8, 0x8f, 0x51, 0xe0, 0x73, 0x69, 0x4c, 0x8f, 0x98, 0x7d, 0xc4, 0xef, 0x20, 0x38, 0x6a,
	0x56, 0xeb, 0x03, 0x5e, 0x8b, 0x89, 0xa6, 0x09, 0xf4, 0x98, 0x09, 0x54, 0xa3, 0x29, 0x53, 0x2a,
	0xca, 0x3d, 0x8f, 0x5b, 0x52, 0x03, 0x82, 0xf1, 0x89, 0x2c, 0x50, 0x23, 0xf5, 0xf9, 0x08, 0xf7,
	0xe3, 0xd0, 0x88, 0x4f, 0x66, 0x67, 0x64, 0xc4, 0x39, 0x0c, 0xf2, 0x2f, 0x32, 0x73, 0x23, 0x8f,
	0x98, 0x18, 0x2e, 0xa8, 0x4f, 0x99, 0x5d, 0xb0, 0x51, 0x10, 0x7f, 0x99, 0x1d, 0xb0, 0x11, 0xd7,
	0xb9, 0x3e, 0x80, 0x40, 0xe6, 0x7b, 0xfc, 0x95, 0xa0, 0xb3, 0xc8, 0x54, 0xa3, 0x2f, 0x12, 0xc0,
	0xe2, 0xb4, 0xbd, 0x7e, 0x23, 0x96, 0xed, 0x20, 0x4f, 0xcb, 0x33, 0xf6, 0xfa, 0x06, 0x05, 0xf1,
	0xd9, 0x2c, 0x39, 0xf3, 0xdf, 0x1e, 0x78, 0xc3, 0x46, 0x7a, 0x2e, 0xdb, 0xd2, 0xfc, 0x87, 0xf7,
	0x31, 0x10, 0x9e, 0x37, 0x99, 0x9b, 0x81, 0x2d, 0x26, 0xea, 0x43, 0x41, 0x90, 0x8b, 0x2f, 0x98,
	0xcc, 0x75, 0x71, 0x30, 0xb8, 0x98, 0xa5, 0x5f, 0x61, 0xd7, 0x1b, 0x4a, 0xfa, 0x81, 0x4a, 0x83,
	0x90, 0xe7, 0xd1, 0x5f, 0x12, 0x74, 0x1e, 0x99, 0x59, 0xd8, 0x7b, 0x9b, 0x04, 0x56, 0x5f, 0x0b,
	0x3a, 0x87, 0x4c, 0x37, 0x56, 0x81, 0xe0, 0xc3, 0x3c, 0x49, 0x79, 0x37, 0x6e, 0xe7, 0xe7, 0xf0,
	0x8d, 0xc9, 0xf6, 0x2a, 0x0e, 0x18, 0x7d, 0x9b, 0x2d, 0xe7, 0x1a, 0xc5, 0xc2, 0xb6, 0xfa, 0x4e,
	0xd0, 0xf9, 0x64, 0x96, 0x6b, 0x95, 0xb3, 0xc0, 0xec, 0xb2, 0xa0, 0x83, 0x64, 0x7e, 0xf9, 0x07,
	0x46, 0x5c, 0xc6, 0xbe, 0x4a, 0xa5, 0xae, 0x18, 0xb4, 0xbc, 0x22, 0xe8, 0x12, 0xb2, 0xa0, 0xfc,
	0x3b, 0x6d, 0x2e, 0x18, 0x5f, 0x15, 0x74, 0x29, 0x59, 0x38, 0xa6, 0xb1, 0xc7, 0x84, 0xc7, 0x7d,
	0x63, 0xfd, 0xbd, 0xa0, 0xf7, 0x92, 0x45, 0x63, 0x5a, 0x1b, 0x36, 0x98, 0x5f, 0x33, 0xe7, 0x3a,
	0xc4, 0x94, 0xf2, 0x39, 0xf2, 0xd0, 0xee, 0x57, 0x73, 0xae, 0x2e, 0x0e, 0x06, 0xbf, 0x95, 0x0d,
	0xbc, 0x20, 0xaa, 0x1b, 0x83, 0xdf, 0xcb, 0x06, 0x80, 0x83, 0xc1, 0x1f, 0x26, 0x31, 0x6c, 0x82,
	0xdf, 0x96, 0x79, 0x18, 0xd7, 0x4d, 0x62, 0x54, 0x92, 0xc0, 0xea, 0xcf, 0xec, 0xa4, 0xca, 0x56,
	0x75, 0x9e, 0xf5, 0x38, 0x34, 0xfb, 0x4b, 0xd0, 0x05, 0x64, 0x76, 0xd9, 0xcc, 0xd0, 0xc0, 0xee,
	0x6f, 0x41, 0x27, 0x93, 0x09, 0xfd, 0x6e, 0xad, 0x52, 0xc5, 0x59, 0xc7, 0x58, 0x5c, 0x0e, 0xb0,
	0xc9, 0x17, 0x20, 0x90, 0x5d, 0x09, 0xb0, 0xa3, 0xe9, 0xef, 0x85, 0xc6, 0x73, 0x35, 0xc0, 0x9e,
	0xd5, 0x56, 0xbc, 0x23, 0xed, 0x6b, 0xe6, 0x93, 0x90, 0xde, 0x4d, 0xee, 0xd2, 0xaa, 0x1c, 0xac,
	0x7d, 0x1a, 0xd2, 0x81, 0xfe, 0x6d, 0xd5, 0xff, 0x18, 0x87, 0x75, 0xa6, 0x78, 0xed, 0xb3, 0x10,
	0x5b, 0x80, 0xfe, 0xec, 0xb6, 0x80, 0xcf, 0x43, 0x6c, 0x01, 0x36, 0x0a, 0x51, 0x1c, 0x0a, 0xb1,
	0x05, 0x64, 0xb0, 0xe4, 0x7e, 0x9e, 0x3c, 0x87, 0x43, 0x6c, 0x01, 0x45, 0x0c, 0x84, 0x47, 0xc2,
	0x62, 0xf8, 0x4e, 0xbd, 0x1d, 0x0d, 0xb1, 0xe5, 0x5a, 0x20, 0x48, 0x8f, 0x85, 0xd8, 0x72, 0x33,
	0xd4, 0xad, 0xb0, 0xe3, 0x21, 0xb6, 0x5c, 0x07, 0xc6, 0x8e, 0xed, 0xc8, 0xc3, 0x66, 0xc4, 0xea,
	0xf9, 0x79, 0x9e, 0x74, 0xe4, 0x06, 0xc6, 0x8e, 0x1d, 0x62, 0x99, 0x67, 0x38, 0x16, 0xb0, 0xe3,
	0x72, 0x2a, 0xc4, 0x32, 0xaf, 0x66, 0x61, 0x07, 0x0f, 0x31, 0xdf, 0x3d, 0xd6, 0x09, 0x59, 0xbb,
	0x29, 0xfa, 0x33, 0x03, 0xda, 0xec, 0xec, 0x62, 0xbe, 0xbb, 0x38, 0x18, 0xbc, 0xd3, 0xa5, 0x0b,
	0xc9, 0x1c, 0xcb, 0x40, 0xdf, 0x6c, 0x71, 0x98, 0xb6, 0x85, 0x17, 0x14, 0xae, 0xdb, 0x5d, 0x5d,
	0xba, 0x88, 0xcc, 0xb5, 0x9c, 0x5c, 0x22, 0x58, 0xee, 0xee, 0xea, 0x76, 0x52, 0xc8, 0x7b, 0x9d,
	0x94, 0xc0, 0x35, 0x5a, 0x33, 0x25, 0x74, 0x75, 0x3b, 0x29, 0x64, 0x7f, 0x35, 0x19, 0xa7, 0x86,
	0x2e, 0x66, 0x4b, 0x87, 0xb5, 0xfd, 0x54, 0x72, 0x91, 0x97, 0xf6, 0x53, 0x11, 0x66, 0x4b, 0x11,
	0x03, 0xe1, 0xd3, 0x91, 0x25, 0xb4, 0xaa, 0xf9, 0x19, 0x5b, 0x68, 0x17, 0xf1, 0xb3, 0x11, 0x26,
	0x77, 0x06, 0xd6, 0xb9, 0xd2, 0x7f, 0x50, 0xba, 0x3e, 0xc2, 0xe4, 0xb6, 0x51, 0x10, 0x3f, 0x17,
	0xe9, 0x3d, 0x33, 0xe2, 0x26, 0x57, 0xfa, 0xf6, 0x67, 0x5e, 0x8b, 0xd7, 0xa1, 0x84, 0xd0, 0x68,
	0x43, 0x44, 0x17, 0x93, 0x79, 0xc6, 0xa8, 0x92, 0x09, 0xa6, 0xcf, 0xbb, 0x11, 0x59, 0xdd, 0x64,
	0xa3, 0x1b, 0x91, 0xdd, 0x44, 0x5e, 0x88, 0x74, 0x13, 0x41, 0x58, 0xa6, 0x82, 0xf7, 0x70, 0x86,
	0xd9, 0x64, 0xb6, 0x48, 0x31, 0x9f, 0x0b, 0x95, 0x16, 0xe6, 0x88, 0x1f, 0xcd, 0x16, 0x15, 0x31,
	0xf0, 0xfc, 0xc9, 0x04, 0x04, 0xa0, 0xb5, 0xbb, 0x3f, 0x9b, 0x80, 0x6c, 0x14, 0xc4, 0xbf, 0x44,
	0x58, 0xc6, 0x2a, 0xe8, 0xf1, 0xc8, 0xce, 0xde, 0xd3, 0x12, 0xcb, 0xd8, 0x02, 0x41, 0x7a, 0x46,
	0x62, 0xab, 0xee, 0xa3, 0xfd, 0xbe, 0x29, 0x71, 0xd2, 0xeb, 0x5b, 0x9c, 0x95, 0xd8, 0xaa, 0x2b,
	0x49, 0x60, 0x75, 0x4e, 0xe2, 0x1d, 0x5e, 0x64, 0xd9, 0xc1, 0x9c, 0x97, 0x78, 0x87, 0x57, 0x71,
	0xc0, 0xe8, 0x82, 0x1b, 0x13, 0x13, 0xc3, 0x6d, 0xd1, 0xb4, 0xb7, 0xe4, 0xa2, 0x1b, 0x93, 0x4d,
	0x02, 0xab, 0x4b, 0x12, 0xcf, 0xa3, 0x1e, 0xb1, 0x5e, 0xea, 0xb1, 0xc2, 0x35, 0xf6, 0x81, 0xc2,
	0xf3, 0x28, 0x62, 0x20, 0xfc, 0x50, 0x59, 0x42, 0xeb, 0x47, 0x7c, 0x64, 0x0b, 0xed, 0xe0, 0x3f,
	0x56, 0xf9, 0x59, 0xc4, 0x5e, 0x2b, 0x6d, 0x06, 0x7e, 0xbe, 0xe4, 0xf6, 0x38, 0x3f, 0x8b, 0x02,
	0x08, 0xd2, 0x97, 0x62, 0x9c, 0x24, 0xb3, 0xef, 0x7a, 0xe0, 0xb2, 0x17, 0xde, 0x11, 0xe3, 0x24,
	0x59, 0x22, 0x80, 0xc5, 0xcb, 0x31, 0x86, 0x2d, 0x5b, 0x41, 0x68, 0xab, 0xaf, 0xc7, 0x18, 0x76,
	0x11, 0xc3, 0x7b, 0x36, 0x36, 0xcf, 0x18, 0x0d, 0x0e, 0xc5, 0x49, 0xbf, 0x67, 0x9a, 0x0b, 0x36,
	0x36, 0xcf, 0x18, 0x07, 0xc7, 0x9b, 0x35, 0xc6, 0x86, 0x9e, 0x11, 0x22, 0xde, 0x88, 0xb8, 0x6c,
	0x19, 0xfd, 0x3f, 0x31, 0x36, 0x74, 0x07, 0x06, 0xf9, 0xbf, 0x31, 0x9d, 0x4d, 0xa6, 0x19, 0x9c,
	0xc5, 0x2a, 0x30, 0x24, 0xa8, 0xad, 0xff, 0xcc, 0x6f, 0xd3, 0x67, 0x6d, 0x67, 0xc3, 0xb5, 0x11,
	0x33, 0xe8, 0x16, 0x30, 0xf0, 0xfe, 0x61, 0x04, 0x8f, 0x84, 0x45, 0x5c, 0x30, 0x7b, 0x57, 0x0e,
	0xf5, 0xf0, 0x48, 0x2c, 0x10, 0xa4, 0x87, 0x7b, 0x38, 0x01, 0xf6, 0x51, 0x78, 0x07, 0xd5, 0x79,
	0x83, 0x0b, 0xc9, 0xed, 0x01, 0xe2, 0x48, 0x0f, 0x27, 0xc0, 0x1b, 0x71, 0xf1, 0x29, 0xd4, 0xc3,
	0x1c, 0xef, 0x93, 0x3b, 0x4c, 0x79, 0x2d, 0x94, 0xa0, 0xe3, 0xb1, 0x1e, 0xe6, 0x78, 0x25, 0x09,
	0xac, 0x8e, 0xf7, 0x74, 0xdd, 0xe5, 0x2c, 0xbc, 0xdd, 0x7a, 0x3a, 0xa1, 0x61, 0xef, 0x4e, 0xf4,
	0xb0, 0x36, 0x99, 0xa7, 0xda, 0x23, 0x5c, 0x3f, 0x95, 0x9b, 0xdc, 0xde, 0x89, 0x3d, 0xa3, 0x58,
	0x9b, 0x55, 0x1c, 0xbc, 0x3c, 0x46, 0xe9, 0x7d, 0x64, 0xb0, 0x64, 0xa4, 0xf3, 0xc1, 0x6b, 0x31,
	0xdf, 0xe7, 0xd9, 0x83, 0x2e, 0xce, 0x77, 0x65, 0xdf, 0x28, 0xbd, 0x9f, 0x2c, 0x29, 0xb9, 0x56,
	0x09, 0x60, 0x89, 0xfd, 0xa3, 0xe6, 0x57, 0x17, 0x15, 0x4e, 0x36, 0x1c, 0xc8, 0x7c, 0x4b, 0x81,
	0xc0, 0x23, 0x9b, 0x49, 0x99, 0xa5, 0x40, 0xf6, 0x60, 0x83, 0x48, 0x0e, 0x8e, 0xd2, 0x07, 0xc8,
	0xd2, 0x92, 0x6f, 0xa5, 0x02, 0x42, 0x79, 0x77, 0x94, 0x2e, 0x23, 0x8b, 0x4b, 0x8b, 0x14, 0xb9,
	0x56, 0x1a, 0xbe, 0x97, 0xed, 0x4e, 0x69, 0x89, 0x0a, 0x3e, 0x2c, 0xf0, 0xfe, 0xa8, 0x4e, 0x7b,
	0xbd, 0x40, 0x23, 0x6a, 0xf7, 0x2f, 0x5b, 0x2f, 0xe8, 0x74, 0x8a, 0x57, 0xf2, 0xfa, 0x04, 0x8f,
	0xb7, 0x82, 0x82, 0x77, 0x64, 0x82, 0xb7, 0x07, 0x70, 0xac, 0xa8, 0x36, 0x24, 0x78, 0x7b, 0xd8,
	0x28, 0xde, 0x85, 0x09, 0x96, 0x07, 0xc0, 0x4c, 0xe6, 0x8f, 0xc0, 0x8d, 0x09, 0x96, 0x87, 0x05,
	0xe2, 0x4d, 0x98, 0xe0, 0xe0, 0x53, 0x40, 0x21, 0x3f, 0xad, 0x10, 0x36, 0x25, 0x38, 0xf8, 0x8c,
	0x49, 0x04, 0xcb, 0x17, 0x13, 0x5d, 0x71, 0x37, 0x60, 0xb2, 0xba, 0xc9, 0xea, 0xcd, 0x09, 0xf6,
	0x1c, 0xe4, 0x36, 0x23, 0x9e, 0x1f, 0xf8, 0x96, 0x04, 0x7b, 0x8e, 0x03, 0xc3, 0x52, 0x5b, 0x13,
	0xec, 0x69, 0xc5, 0x7d, 0x29, 0xf8, 0x6f, 0x4b, 0xb0, 0x29, 0x9a, 0xad, 0x6f, 0xc4, 0x85, 0x97,
	0xfd, 0x76, 0xd7, 0xc0, 0xe0, 0xd8, 0xd1, 0xcb, 0x06, 0x9d, 0xa0, 0xf0, 0x68, 0xdc, 0x51, 0x36,
	0x00, 0x1c, 0xfb, 0x79, 0x82, 0x6d, 0x02, 0x08, 0x4d, 0x9d, 0x55, 0x61, 0xd0, 0x16, 0x2a, 0x9f,
	0x7b, 0x5e, 0x49, 0xb0, 0x60, 0x2a, 0x49, 0x60, 0xf5, 0x6a, 0x82, 0x2d, 0x00, 0x59, 0x5c, 0xb9,
	0x4e, 0xaf, 0x25, 0xd8, 0x02, 0xaa, 0x38, 0x60, 0xf4, 0xba, 0x1b, 0x93, 0xe4, 0x2c, 0xf2, 0x5a,
	0xe9, 0x50, 0x20, 0x73, 0xa7, 0x37, 0xdc, 0x98, 0x6c, 0x12, 0x58, 0xbd, 0xe9, 0xc4, 0x24, 0xfb,
	0xb8, 0x95, 0x3a, 0x6f, 0x39, 0x31, 0xd9, 0x1c, 0x30, 0x7a, 0x3b, 0xc1, 0x21, 0x11, 0x96, 0xcb,
	0x38, 0xd9, 0xf4, 0xe7, 0xdc, 0x14, 0x3b, 0x13, 0x1c, 0x12, 0xc7, 0x66, 0xe2, 0x58, 0xef, 0x56,
	0x95, 0xd5, 0x2e, 0x77, 0xb9, 0x55, 0x65, 0x37, 0xca, 0xdd, 0x37, 0x5a, 0xc7, 0x6e, 0xce, 0x7b,
	0xdc, 0x2a, 0xd2, 0xe3, 0x7b, 0x55, 0x07, 0xdb, 0xeb, 0x56, 0x51, 0x99, 0x08, 0xab, 0xef, 0x4b,
	0x1e, 0xbe, 0x3d, 0xfb, 0x77, 0xeb, 0x83, 0xff, 0x07, 0x00, 0x00, 0xff, 0xff, 0x81, 0xac, 0x55,
	0x6f, 0x89, 0x15, 0x00, 0x00,
}
