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
	MSGID_NONE                                    MSGID = 0
	MSGID_C2S_TEST_COMMAND                        MSGID = 1
	MSGID_C2S_HEARTBEAT                           MSGID = 2
	MSGID_S2C_STATE_NOTIFY                        MSGID = 3
	MSGID_C2S_DATA_SYNC_REQUEST                   MSGID = 4
	MSGID_C2S_LOGIN_REQUEST                       MSGID = 10000
	MSGID_S2C_LOGIN_RESPONSE                      MSGID = 10001
	MSGID_S2C_OTHER_PLACE_LOGIN                   MSGID = 10002
	MSGID_C2S_SELECT_SERVER_REQUEST               MSGID = 10003
	MSGID_S2C_SELECT_SERVER_RESPONSE              MSGID = 10004
	MSGID_C2S_ENTER_GAME_REQUEST                  MSGID = 10020
	MSGID_S2C_ENTER_GAME_RESPONSE                 MSGID = 10021
	MSGID_S2C_ENTER_GAME_COMPLETE_NOTIFY          MSGID = 10022
	MSGID_C2S_LEAVE_GAME_REQUEST                  MSGID = 10023
	MSGID_S2C_LEAVE_GAME_RESPONSE                 MSGID = 10024
	MSGID_S2C_PLAYER_INFO_RESPONSE                MSGID = 10025
	MSGID_C2S_ROLES_REQUEST                       MSGID = 10050
	MSGID_S2C_ROLES_RESPONSE                      MSGID = 10051
	MSGID_S2C_ROLES_CHANGE_NOTIFY                 MSGID = 10052
	MSGID_C2S_ROLE_LEVELUP_REQUEST                MSGID = 10053
	MSGID_S2C_ROLE_LEVELUP_RESPONSE               MSGID = 10054
	MSGID_C2S_ROLE_RANKUP_REQUEST                 MSGID = 10055
	MSGID_S2C_ROLE_RANKUP_RESPONSE                MSGID = 10056
	MSGID_C2S_ROLE_DECOMPOSE_REQUEST              MSGID = 10057
	MSGID_S2C_ROLE_DECOMPOSE_RESPONSE             MSGID = 10058
	MSGID_C2S_ROLE_FUSION_REQUEST                 MSGID = 10059
	MSGID_S2C_ROLE_FUSION_RESPONSE                MSGID = 10060
	MSGID_C2S_ROLE_LOCK_REQUEST                   MSGID = 10061
	MSGID_S2C_ROLE_LOCK_RESPONSE                  MSGID = 10062
	MSGID_C2S_ROLE_HANDBOOK_REQUEST               MSGID = 10063
	MSGID_S2C_ROLE_HANDBOOK_RESPONSE              MSGID = 10064
	MSGID_C2S_ROLE_LEFTSLOT_OPEN_REQUEST          MSGID = 10065
	MSGID_S2C_ROLE_LEFTSLOT_OPEN_RESPONSE         MSGID = 10066
	MSGID_C2S_ROLE_ONEKEY_EQUIP_REQUEST           MSGID = 10067
	MSGID_S2C_ROLE_ONEKEY_EQUIP_RESPONSE          MSGID = 10068
	MSGID_C2S_ROLE_ONEKEY_UNEQUIP_REQUEST         MSGID = 10069
	MSGID_S2C_ROLE_ONEKEY_UNEQUIP_RESPONSE        MSGID = 10070
	MSGID_C2S_BATTLE_RESULT_REQUEST               MSGID = 10100
	MSGID_S2C_BATTLE_RESULT_RESPONSE              MSGID = 10101
	MSGID_C2S_BATTLE_RECORD_REQUEST               MSGID = 10102
	MSGID_S2C_BATTLE_RECORD_RESPONSE              MSGID = 10103
	MSGID_C2S_BATTLE_RECORD_LIST_REQUEST          MSGID = 10104
	MSGID_S2C_BATTLE_RECORD_LIST_RESPONSE         MSGID = 10105
	MSGID_C2S_BATTLE_RECORD_DELETE_REQUEST        MSGID = 10106
	MSGID_S2C_BATTLE_RECORD_DELETE_RESPONSE       MSGID = 10107
	MSGID_C2S_SET_TEAM_REQUEST                    MSGID = 10200
	MSGID_S2C_SET_TEAM_RESPONSE                   MSGID = 10201
	MSGID_S2C_TEAMS_RESPONSE                      MSGID = 10202
	MSGID_C2S_ITEMS_SYNC_REQUEST                  MSGID = 10300
	MSGID_S2C_ITEMS_SYNC                          MSGID = 10301
	MSGID_S2C_ITEMS_UPDATE                        MSGID = 10302
	MSGID_C2S_ITEM_FUSION_REQUEST                 MSGID = 10303
	MSGID_S2C_ITEM_FUSION_RESPONSE                MSGID = 10304
	MSGID_C2S_ITEM_SELL_REQUEST                   MSGID = 10305
	MSGID_S2C_ITEM_SELL_RESPONSE                  MSGID = 10306
	MSGID_C2S_ITEM_EQUIP_REQUEST                  MSGID = 10307
	MSGID_S2C_ITEM_EQUIP_RESPONSE                 MSGID = 10308
	MSGID_C2S_ITEM_UNEQUIP_REQUEST                MSGID = 10309
	MSGID_S2C_ITEM_UNEQUIP_RESPONSE               MSGID = 10310
	MSGID_C2S_ITEM_UPGRADE_REQUEST                MSGID = 10311
	MSGID_S2C_ITEM_UPGRADE_RESPONSE               MSGID = 10312
	MSGID_C2S_CAMPAIGN_DATA_REQUEST               MSGID = 10400
	MSGID_S2C_CAMPAIGN_DATA_RESPONSE              MSGID = 10401
	MSGID_C2S_CAMPAIGN_HANGUP_INCOME_REQUEST      MSGID = 10402
	MSGID_S2C_CAMPAIGN_HANGUP_INCOME_RESPONSE     MSGID = 10403
	MSGID_C2S_BATTLE_SET_HANGUP_CAMPAIGN_REQUEST  MSGID = 10404
	MSGID_S2C_BATTLE_SET_HANGUP_CAMPAIGN_RESPONSE MSGID = 10405
	MSGID_C2S_MAIL_SEND_REQUEST                   MSGID = 10500
	MSGID_S2C_MAIL_SEND_RESPONSE                  MSGID = 10501
	MSGID_C2S_MAIL_LIST_REQUEST                   MSGID = 10502
	MSGID_S2C_MAIL_LIST_RESPONSE                  MSGID = 10503
	MSGID_C2S_MAIL_DETAIL_REQUEST                 MSGID = 10504
	MSGID_S2C_MAIL_DETAIL_RESPONSE                MSGID = 10505
	MSGID_C2S_MAIL_GET_ATTACHED_ITEMS_REQUEST     MSGID = 10506
	MSGID_S2C_MAIL_GET_ATTACHED_ITEMS_RESPONSE    MSGID = 10507
	MSGID_C2S_MAIL_DELETE_REQUEST                 MSGID = 10508
	MSGID_S2C_MAIL_DELETE_RESPONSE                MSGID = 10509
	MSGID_S2C_MAILS_NEW_NOTIFY                    MSGID = 10510
	MSGID_C2S_TALENT_UP_REQUEST                   MSGID = 10600
	MSGID_S2C_TALENT_UP_RESPONSE                  MSGID = 10601
	MSGID_C2S_TALENT_LIST_REQUEST                 MSGID = 10602
	MSGID_S2C_TALENT_LIST_RESPONSE                MSGID = 10603
	MSGID_C2S_TOWER_DATA_REQUEST                  MSGID = 10700
	MSGID_S2C_TOWER_DATA_RESPONSE                 MSGID = 10701
	MSGID_C2S_TOWER_RECORDS_INFO_REQUEST          MSGID = 10702
	MSGID_S2C_TOWER_RECORDS_INFO_RESPONSE         MSGID = 10703
	MSGID_C2S_TOWER_RECORD_DATA_REQUEST           MSGID = 10704
	MSGID_S2C_TOWER_RECORD_DATA_RESPONSE          MSGID = 10705
	MSGID_C2S_TOWER_RANKING_LIST_REQUEST          MSGID = 10706
	MSGID_S2C_TOWER_RANKING_LIST_RESPONSE         MSGID = 10707
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
	10053: "C2S_ROLE_LEVELUP_REQUEST",
	10054: "S2C_ROLE_LEVELUP_RESPONSE",
	10055: "C2S_ROLE_RANKUP_REQUEST",
	10056: "S2C_ROLE_RANKUP_RESPONSE",
	10057: "C2S_ROLE_DECOMPOSE_REQUEST",
	10058: "S2C_ROLE_DECOMPOSE_RESPONSE",
	10059: "C2S_ROLE_FUSION_REQUEST",
	10060: "S2C_ROLE_FUSION_RESPONSE",
	10061: "C2S_ROLE_LOCK_REQUEST",
	10062: "S2C_ROLE_LOCK_RESPONSE",
	10063: "C2S_ROLE_HANDBOOK_REQUEST",
	10064: "S2C_ROLE_HANDBOOK_RESPONSE",
	10065: "C2S_ROLE_LEFTSLOT_OPEN_REQUEST",
	10066: "S2C_ROLE_LEFTSLOT_OPEN_RESPONSE",
	10067: "C2S_ROLE_ONEKEY_EQUIP_REQUEST",
	10068: "S2C_ROLE_ONEKEY_EQUIP_RESPONSE",
	10069: "C2S_ROLE_ONEKEY_UNEQUIP_REQUEST",
	10070: "S2C_ROLE_ONEKEY_UNEQUIP_RESPONSE",
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
}
var MSGID_value = map[string]int32{
	"NONE":                                    0,
	"C2S_TEST_COMMAND":                        1,
	"C2S_HEARTBEAT":                           2,
	"S2C_STATE_NOTIFY":                        3,
	"C2S_DATA_SYNC_REQUEST":                   4,
	"C2S_LOGIN_REQUEST":                       10000,
	"S2C_LOGIN_RESPONSE":                      10001,
	"S2C_OTHER_PLACE_LOGIN":                   10002,
	"C2S_SELECT_SERVER_REQUEST":               10003,
	"S2C_SELECT_SERVER_RESPONSE":              10004,
	"C2S_ENTER_GAME_REQUEST":                  10020,
	"S2C_ENTER_GAME_RESPONSE":                 10021,
	"S2C_ENTER_GAME_COMPLETE_NOTIFY":          10022,
	"C2S_LEAVE_GAME_REQUEST":                  10023,
	"S2C_LEAVE_GAME_RESPONSE":                 10024,
	"S2C_PLAYER_INFO_RESPONSE":                10025,
	"C2S_ROLES_REQUEST":                       10050,
	"S2C_ROLES_RESPONSE":                      10051,
	"S2C_ROLES_CHANGE_NOTIFY":                 10052,
	"C2S_ROLE_LEVELUP_REQUEST":                10053,
	"S2C_ROLE_LEVELUP_RESPONSE":               10054,
	"C2S_ROLE_RANKUP_REQUEST":                 10055,
	"S2C_ROLE_RANKUP_RESPONSE":                10056,
	"C2S_ROLE_DECOMPOSE_REQUEST":              10057,
	"S2C_ROLE_DECOMPOSE_RESPONSE":             10058,
	"C2S_ROLE_FUSION_REQUEST":                 10059,
	"S2C_ROLE_FUSION_RESPONSE":                10060,
	"C2S_ROLE_LOCK_REQUEST":                   10061,
	"S2C_ROLE_LOCK_RESPONSE":                  10062,
	"C2S_ROLE_HANDBOOK_REQUEST":               10063,
	"S2C_ROLE_HANDBOOK_RESPONSE":              10064,
	"C2S_ROLE_LEFTSLOT_OPEN_REQUEST":          10065,
	"S2C_ROLE_LEFTSLOT_OPEN_RESPONSE":         10066,
	"C2S_ROLE_ONEKEY_EQUIP_REQUEST":           10067,
	"S2C_ROLE_ONEKEY_EQUIP_RESPONSE":          10068,
	"C2S_ROLE_ONEKEY_UNEQUIP_REQUEST":         10069,
	"S2C_ROLE_ONEKEY_UNEQUIP_RESPONSE":        10070,
	"C2S_BATTLE_RESULT_REQUEST":               10100,
	"S2C_BATTLE_RESULT_RESPONSE":              10101,
	"C2S_BATTLE_RECORD_REQUEST":               10102,
	"S2C_BATTLE_RECORD_RESPONSE":              10103,
	"C2S_BATTLE_RECORD_LIST_REQUEST":          10104,
	"S2C_BATTLE_RECORD_LIST_RESPONSE":         10105,
	"C2S_BATTLE_RECORD_DELETE_REQUEST":        10106,
	"S2C_BATTLE_RECORD_DELETE_RESPONSE":       10107,
	"C2S_SET_TEAM_REQUEST":                    10200,
	"S2C_SET_TEAM_RESPONSE":                   10201,
	"S2C_TEAMS_RESPONSE":                      10202,
	"C2S_ITEMS_SYNC_REQUEST":                  10300,
	"S2C_ITEMS_SYNC":                          10301,
	"S2C_ITEMS_UPDATE":                        10302,
	"C2S_ITEM_FUSION_REQUEST":                 10303,
	"S2C_ITEM_FUSION_RESPONSE":                10304,
	"C2S_ITEM_SELL_REQUEST":                   10305,
	"S2C_ITEM_SELL_RESPONSE":                  10306,
	"C2S_ITEM_EQUIP_REQUEST":                  10307,
	"S2C_ITEM_EQUIP_RESPONSE":                 10308,
	"C2S_ITEM_UNEQUIP_REQUEST":                10309,
	"S2C_ITEM_UNEQUIP_RESPONSE":               10310,
	"C2S_ITEM_UPGRADE_REQUEST":                10311,
	"S2C_ITEM_UPGRADE_RESPONSE":               10312,
	"C2S_CAMPAIGN_DATA_REQUEST":               10400,
	"S2C_CAMPAIGN_DATA_RESPONSE":              10401,
	"C2S_CAMPAIGN_HANGUP_INCOME_REQUEST":      10402,
	"S2C_CAMPAIGN_HANGUP_INCOME_RESPONSE":     10403,
	"C2S_BATTLE_SET_HANGUP_CAMPAIGN_REQUEST":  10404,
	"S2C_BATTLE_SET_HANGUP_CAMPAIGN_RESPONSE": 10405,
	"C2S_MAIL_SEND_REQUEST":                   10500,
	"S2C_MAIL_SEND_RESPONSE":                  10501,
	"C2S_MAIL_LIST_REQUEST":                   10502,
	"S2C_MAIL_LIST_RESPONSE":                  10503,
	"C2S_MAIL_DETAIL_REQUEST":                 10504,
	"S2C_MAIL_DETAIL_RESPONSE":                10505,
	"C2S_MAIL_GET_ATTACHED_ITEMS_REQUEST":     10506,
	"S2C_MAIL_GET_ATTACHED_ITEMS_RESPONSE":    10507,
	"C2S_MAIL_DELETE_REQUEST":                 10508,
	"S2C_MAIL_DELETE_RESPONSE":                10509,
	"S2C_MAILS_NEW_NOTIFY":                    10510,
	"C2S_TALENT_UP_REQUEST":                   10600,
	"S2C_TALENT_UP_RESPONSE":                  10601,
	"C2S_TALENT_LIST_REQUEST":                 10602,
	"S2C_TALENT_LIST_RESPONSE":                10603,
	"C2S_TOWER_DATA_REQUEST":                  10700,
	"S2C_TOWER_DATA_RESPONSE":                 10701,
	"C2S_TOWER_RECORDS_INFO_REQUEST":          10702,
	"S2C_TOWER_RECORDS_INFO_RESPONSE":         10703,
	"C2S_TOWER_RECORD_DATA_REQUEST":           10704,
	"S2C_TOWER_RECORD_DATA_RESPONSE":          10705,
	"C2S_TOWER_RANKING_LIST_REQUEST":          10706,
	"S2C_TOWER_RANKING_LIST_RESPONSE":         10707,
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
	// 1033 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x6c, 0x96, 0x59, 0x8f, 0x1c, 0x35,
	0x10, 0xc7, 0x39, 0x02, 0x42, 0x96, 0x40, 0x1d, 0x93, 0xdd, 0x65, 0x13, 0x72, 0x90, 0x84, 0x84,
	0x4b, 0x79, 0x08, 0x9f, 0xc0, 0xdb, 0x5d, 0x3b, 0xd3, 0xda, 0x1e, 0xdb, 0x63, 0x7b, 0x36, 0xda,
	0x27, 0x8b, 0x63, 0x15, 0x45, 0x22, 0x04, 0xb1, 0xf9, 0x08, 0xdc, 0x97, 0xb8, 0x1e, 0x78, 0xe4,
	0x3e, 0xbe, 0x03, 0x67, 0x0e, 0xc2, 0x91, 0x3b, 0x80, 0x80, 0x4f, 0xc0, 0xf1, 0xca, 0xcd, 0x4b,
	0xe4, 0x6e, 0x57, 0x75, 0xbb, 0x67, 0x9e, 0x56, 0xf2, 0xbf, 0xfe, 0xbf, 0xae, 0x29, 0x57, 0x95,
	0x97, 0x2d, 0x3c, 0xf4, 0xc8, 0x91, 0xf5, 0x47, 0x8f, 0xfb, 0xa3, 0xeb, 0x1b, 0x1b, 0x0f, 0x1c,
	0x5e, 0xf7, 0x47, 0x1e, 0x3e, 0xf0, 0xd8, 0xe3, 0xc7, 0x8e, 0x1f, 0xe3, 0x73, 0x47, 0x37, 0x0e,
	0x1f, 0x98, 0x12, 0xef, 0x79, 0x63, 0x9e, 0xdd, 0x30, 0xb2, 0x83, 0xb2, 0xe0, 0x37, 0xb1, 0x4d,
	0x52, 0x49, 0xc8, 0xae, 0xe1, 0x5b, 0x58, 0x96, 0x1f, 0xb4, 0xde, 0x81, 0x75, 0x3e, 0x57, 0xa3,
	0x91, 0x90, 0x45, 0x76, 0x2d, 0xdf, 0xcc, 0x6e, 0x0e, 0xa7, 0x43, 0x10, 0xc6, 0x2d, 0x81, 0x70,
	0xd9, 0x75, 0x21, 0xd0, 0x1e, 0xcc, 0xbd, 0x75, 0xc2, 0x81, 0x97, 0xca, 0x95, 0xcb, 0x6b, 0xd9,
	0xf5, 0x7c, 0x91, 0xcd, 0x85, 0xc0, 0x42, 0x38, 0xe1, 0xed, 0x9a, 0xcc, 0xbd, 0x81, 0xf1, 0x04,
	0xac, 0xcb, 0x36, 0xf1, 0x79, 0xb6, 0x39, 0x48, 0x95, 0x1a, 0x94, 0x92, 0x8e, 0x5f, 0x96, 0x7c,
	0x81, 0xf1, 0x00, 0xc2, 0x73, 0xab, 0x95, 0xb4, 0x90, 0xbd, 0x22, 0xf9, 0x56, 0x36, 0x17, 0x04,
	0xe5, 0x86, 0x60, 0xbc, 0xae, 0x44, 0x0e, 0x4d, 0x50, 0xf6, 0xaa, 0xe4, 0x3b, 0xd8, 0x62, 0x80,
	0x59, 0xa8, 0x20, 0x77, 0xde, 0x82, 0x59, 0x05, 0x43, 0xd0, 0xd7, 0x24, 0xdf, 0xc9, 0xb6, 0xd6,
	0xd9, 0xf5, 0xf4, 0x08, 0x7f, 0x5d, 0xf2, 0x6d, 0x6c, 0x3e, 0x00, 0x40, 0x3a, 0x30, 0x7e, 0x20,
	0x46, 0x40, 0xee, 0x77, 0x25, 0xbf, 0x9d, 0x2d, 0x04, 0x77, 0x22, 0x46, 0xeb, 0x7b, 0x92, 0xef,
	0x61, 0x3b, 0x7a, 0x6a, 0xae, 0x46, 0xba, 0x82, 0xb6, 0x0e, 0xef, 0x13, 0xbf, 0x02, 0xb1, 0x0a,
	0x29, 0xff, 0x03, 0xe2, 0x27, 0x62, 0xe4, 0x7f, 0x28, 0xf9, 0x76, 0x76, 0x5b, 0x50, 0x75, 0x25,
	0xd6, 0xc0, 0xf8, 0x52, 0x2e, 0xab, 0x56, 0xfe, 0x48, 0x62, 0x1d, 0x8d, 0xaa, 0xc0, 0x12, 0xf4,
	0x04, 0xd5, 0x11, 0xcf, 0xa3, 0xe1, 0x24, 0x7d, 0xad, 0x11, 0xf2, 0xa1, 0x90, 0x03, 0x4a, 0xf4,
	0x54, 0xfd, 0x35, 0xc4, 0xf9, 0x0a, 0x56, 0xa1, 0x9a, 0x68, 0xa2, 0x9e, 0xae, 0x0b, 0x8d, 0xe6,
	0x8e, 0x1c, 0xe1, 0x5f, 0xd6, 0x70, 0xb2, 0x1b, 0x21, 0x57, 0x3a, 0xee, 0x33, 0xf4, 0x53, 0x52,
	0x35, 0x9a, 0xbf, 0xaa, 0x6f, 0x89, 0xcc, 0x05, 0x84, 0x2a, 0x2a, 0xdb, 0x16, 0xea, 0x6b, 0xc9,
	0x77, 0xb1, 0x6d, 0xe4, 0xef, 0x06, 0x44, 0xc4, 0x37, 0xe9, 0xf7, 0x97, 0x27, 0xb6, 0x54, 0x6d,
	0x6f, 0x7d, 0x9b, 0x7e, 0x9f, 0xd4, 0x68, 0x3e, 0x5b, 0x77, 0x58, 0xfb, 0xdb, 0x55, 0xbe, 0x42,
	0xd6, 0x73, 0xf5, 0x05, 0xb6, 0x3f, 0xbc, 0xd1, 0xa2, 0xf1, 0x3c, 0xb5, 0x5f, 0x2d, 0x0e, 0x85,
	0x2c, 0x96, 0x94, 0x6a, 0xcd, 0x17, 0xa8, 0xfd, 0xfa, 0x7a, 0x04, 0x5c, 0xac, 0x7b, 0xa8, 0x53,
	0xf5, 0x65, 0x67, 0x2b, 0xe5, 0xbc, 0xd2, 0xd0, 0x66, 0x7f, 0x49, 0xf2, 0xbd, 0x6c, 0x67, 0xa7,
	0xf6, 0x69, 0x50, 0x44, 0x5d, 0x96, 0x7c, 0x37, 0xdb, 0x4e, 0x28, 0x25, 0x61, 0x05, 0xd6, 0x3c,
	0x8c, 0x27, 0x65, 0x7b, 0x0f, 0x57, 0xa8, 0x65, 0x67, 0xc5, 0x44, 0xd0, 0x77, 0xf5, 0xe7, 0xfa,
	0xa0, 0x89, 0x4c, 0x51, 0xdf, 0x4b, 0x7e, 0x27, 0xdb, 0xd5, 0x47, 0xb5, 0x51, 0x11, 0xf6, 0x03,
	0x55, 0x68, 0x49, 0x38, 0x57, 0xd5, 0x37, 0x36, 0xa9, 0x1c, 0x61, 0xfe, 0xa0, 0x0a, 0xf5, 0xf5,
	0x08, 0xf8, 0x73, 0x1a, 0x90, 0x2b, 0x53, 0x10, 0xe0, 0xaf, 0x69, 0x40, 0xd4, 0x23, 0xe0, 0x6f,
	0x2a, 0x71, 0x1a, 0x50, 0x95, 0xb6, 0x4d, 0xe3, 0x1f, 0x2a, 0xf1, 0xcc, 0xa0, 0x88, 0xfa, 0xb7,
	0xfe, 0xcd, 0xd3, 0xa8, 0x02, 0xea, 0x91, 0x47, 0xd8, 0x7f, 0x92, 0xef, 0x63, 0x77, 0x4c, 0xc3,
	0x28, 0x2c, 0xe2, 0xfe, 0x97, 0x7c, 0x91, 0x6d, 0x69, 0x96, 0x97, 0xf3, 0x0e, 0xc4, 0x88, 0x10,
	0x3f, 0x2a, 0xdc, 0x79, 0x1d, 0x29, 0xda, 0x7e, 0x52, 0x38, 0xe0, 0xe1, 0xbc, 0x33, 0xe0, 0x3f,
	0x2b, 0xdc, 0x35, 0xa5, 0x83, 0x91, 0x4d, 0xb7, 0xee, 0xc7, 0x9a, 0xdf, 0xca, 0x6e, 0x09, 0xae,
	0x56, 0xcc, 0x3e, 0xd1, 0x7c, 0xae, 0x59, 0xde, 0xcd, 0xe1, 0x44, 0x17, 0xc2, 0x41, 0xf6, 0xa9,
	0xc6, 0x61, 0x0a, 0xc7, 0xfd, 0x61, 0xfa, 0x4c, 0xe3, 0x30, 0xa5, 0x6a, 0xcc, 0xe2, 0x73, 0x8d,
	0xc3, 0x54, 0xcb, 0x16, 0xaa, 0x8a, 0xac, 0x5f, 0x68, 0x1c, 0xa6, 0xae, 0x16, 0x8d, 0x27, 0x74,
	0x37, 0xfd, 0x5e, 0xe7, 0x9e, 0xd4, 0xb8, 0xbc, 0x12, 0x31, 0x5a, 0x4f, 0x69, 0x5c, 0x5e, 0xb5,
	0xda, 0xef, 0xd5, 0xd3, 0x1a, 0x97, 0x57, 0x4f, 0xc6, 0xe5, 0xd5, 0xb3, 0xeb, 0x81, 0x11, 0x45,
	0x7b, 0x9f, 0x67, 0x7a, 0x76, 0x92, 0x71, 0x7d, 0x69, 0x6c, 0xd1, 0x5c, 0x8c, 0xb4, 0x28, 0x07,
	0xb2, 0x79, 0xf5, 0xd0, 0xff, 0xe6, 0x18, 0x5b, 0xb4, 0xaf, 0x47, 0xc0, 0x5b, 0x63, 0xbe, 0x9f,
	0xed, 0x4e, 0x00, 0x61, 0x37, 0x4f, 0xb4, 0x2f, 0x65, 0xae, 0x3a, 0x0f, 0xc6, 0xdb, 0x63, 0x7e,
	0x17, 0xdb, 0x93, 0x90, 0xfa, 0x81, 0x11, 0xf9, 0xce, 0x98, 0xdf, 0xcb, 0xf6, 0x75, 0x5a, 0x35,
	0xf4, 0x51, 0x8c, 0x25, 0x2f, 0xbd, 0x73, 0x63, 0x7e, 0x1f, 0xdb, 0xdf, 0x69, 0xd8, 0xd9, 0xc1,
	0xf8, 0xee, 0x8d, 0xf1, 0x82, 0x47, 0xa2, 0xac, 0xbc, 0x05, 0xd9, 0x4e, 0xe3, 0x13, 0x06, 0x2f,
	0xb8, 0xab, 0x45, 0xe3, 0x93, 0x26, 0x31, 0x26, 0x03, 0xf8, 0x54, 0x6a, 0x4c, 0xe7, 0xee, 0x69,
	0x83, 0xfd, 0x58, 0x8b, 0x05, 0xb8, 0xf0, 0x07, 0xad, 0xcf, 0x18, 0xec, 0xc7, 0x54, 0x8d, 0xe6,
	0x67, 0x4d, 0xa8, 0x19, 0x99, 0x07, 0xe0, 0xbc, 0x70, 0x4e, 0xe4, 0x43, 0x28, 0x62, 0xd7, 0x23,
	0xe8, 0x39, 0xc3, 0xef, 0x66, 0x7b, 0x09, 0x34, 0x33, 0x32, 0x42, 0x9f, 0xef, 0x67, 0x94, 0x2c,
	0x80, 0x17, 0xfa, 0x19, 0xa5, 0x73, 0xff, 0xa2, 0x09, 0x73, 0x8f, 0xb2, 0xf5, 0x12, 0x0e, 0xe1,
	0x2b, 0xfc, 0x12, 0x95, 0xc8, 0x89, 0x0a, 0xa4, 0xf3, 0x9d, 0x47, 0xf4, 0x17, 0x2a, 0x51, 0x57,
	0x8b, 0xcc, 0x5f, 0x29, 0xa1, 0x28, 0x26, 0xd5, 0xfd, 0x8d, 0x12, 0x4a, 0xd5, 0x68, 0xfe, 0xdd,
	0xe0, 0xe4, 0x39, 0x75, 0x08, 0x4c, 0xda, 0xbd, 0x67, 0x2d, 0x4e, 0x5e, 0x22, 0x46, 0xeb, 0x39,
	0x8b, 0xdb, 0xb5, 0x51, 0x9b, 0x55, 0x67, 0xf1, 0x7f, 0x95, 0x06, 0x71, 0xde, 0xe2, 0x76, 0x9d,
	0x19, 0x14, 0x51, 0x17, 0x2c, 0x3e, 0x60, 0xdd, 0xa8, 0x34, 0x99, 0x8b, 0x16, 0x1f, 0xb0, 0x59,
	0x31, 0x11, 0x74, 0xa9, 0x9f, 0x93, 0x90, 0x2b, 0xa5, 0x1c, 0xa4, 0x25, 0xb9, 0xdc, 0xcf, 0x29,
	0x0d, 0x8a, 0xa8, 0x2b, 0xf6, 0xc1, 0x1b, 0xeb, 0x7f, 0x9c, 0xef, 0xbf, 0x1a, 0x00, 0x00, 0xff,
	0xff, 0x9d, 0x9c, 0x87, 0xcc, 0x53, 0x0b, 0x00, 0x00,
}
