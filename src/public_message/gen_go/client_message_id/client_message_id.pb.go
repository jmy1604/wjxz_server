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
	MSGID_C2S_LOGIN_REQUEST                       MSGID = 10000
	MSGID_S2C_LOGIN_RESPONSE                      MSGID = 10001
	MSGID_S2C_OTHER_PLACE_LOGIN                   MSGID = 10002
	MSGID_C2S_SELECT_SERVER_REQUEST               MSGID = 10003
	MSGID_S2C_SELECT_SERVER_RESPONSE              MSGID = 10004
	MSGID_C2S_ENTER_GAME_REQUEST                  MSGID = 11000
	MSGID_S2C_ENTER_GAME_RESPONSE                 MSGID = 11001
	MSGID_S2C_ENTER_GAME_COMPLETE_NOTIFY          MSGID = 11002
	MSGID_C2S_LEAVE_GAME_REQUEST                  MSGID = 11003
	MSGID_S2C_LEAVE_GAME_RESPONSE                 MSGID = 11004
	MSGID_S2C_PLAYER_INFO_RESPONSE                MSGID = 11005
	MSGID_S2C_ROLES_RESPONSE                      MSGID = 11050
	MSGID_S2C_ROLES_CHANGE_NOTIFY                 MSGID = 11151
	MSGID_C2S_ROLE_LEVELUP_REQUEST                MSGID = 11152
	MSGID_S2C_ROLE_LEVELUP_RESPONSE               MSGID = 11153
	MSGID_C2S_ROLE_RANKUP_REQUEST                 MSGID = 11154
	MSGID_S2C_ROLE_RANKUP_RESPONSE                MSGID = 11155
	MSGID_C2S_ROLE_DECOMPOSE_REQUEST              MSGID = 11156
	MSGID_S2C_ROLE_DECOMPOSE_RESPONSE             MSGID = 11157
	MSGID_C2S_ROLE_FUSION_REQUEST                 MSGID = 11158
	MSGID_S2C_ROLE_FUSION_RESPONSE                MSGID = 11159
	MSGID_C2S_BATTLE_RESULT_REQUEST               MSGID = 10100
	MSGID_S2C_BATTLE_RESULT_RESPONSE              MSGID = 10101
	MSGID_C2S_SET_TEAM_REQUEST                    MSGID = 10200
	MSGID_S2C_SET_TEAM_RESPONSE                   MSGID = 10201
	MSGID_S2C_TEAMS_RESPONSE                      MSGID = 10202
	MSGID_S2C_ITEMS_SYNC                          MSGID = 10300
	MSGID_S2C_ITEMS_UPDATE                        MSGID = 10301
	MSGID_C2S_ITEM_FUSION_REQUEST                 MSGID = 10302
	MSGID_S2C_ITEM_FUSION_RESPONSE                MSGID = 10303
	MSGID_C2S_ITEM_SELL_REQUEST                   MSGID = 10304
	MSGID_S2C_ITEM_SELL_RESPONSE                  MSGID = 10305
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
)

var MSGID_name = map[int32]string{
	0:     "NONE",
	1:     "C2S_TEST_COMMAND",
	2:     "C2S_HEARTBEAT",
	3:     "S2C_STATE_NOTIFY",
	10000: "C2S_LOGIN_REQUEST",
	10001: "S2C_LOGIN_RESPONSE",
	10002: "S2C_OTHER_PLACE_LOGIN",
	10003: "C2S_SELECT_SERVER_REQUEST",
	10004: "S2C_SELECT_SERVER_RESPONSE",
	11000: "C2S_ENTER_GAME_REQUEST",
	11001: "S2C_ENTER_GAME_RESPONSE",
	11002: "S2C_ENTER_GAME_COMPLETE_NOTIFY",
	11003: "C2S_LEAVE_GAME_REQUEST",
	11004: "S2C_LEAVE_GAME_RESPONSE",
	11005: "S2C_PLAYER_INFO_RESPONSE",
	11050: "S2C_ROLES_RESPONSE",
	11151: "S2C_ROLES_CHANGE_NOTIFY",
	11152: "C2S_ROLE_LEVELUP_REQUEST",
	11153: "S2C_ROLE_LEVELUP_RESPONSE",
	11154: "C2S_ROLE_RANKUP_REQUEST",
	11155: "S2C_ROLE_RANKUP_RESPONSE",
	11156: "C2S_ROLE_DECOMPOSE_REQUEST",
	11157: "S2C_ROLE_DECOMPOSE_RESPONSE",
	11158: "C2S_ROLE_FUSION_REQUEST",
	11159: "S2C_ROLE_FUSION_RESPONSE",
	10100: "C2S_BATTLE_RESULT_REQUEST",
	10101: "S2C_BATTLE_RESULT_RESPONSE",
	10200: "C2S_SET_TEAM_REQUEST",
	10201: "S2C_SET_TEAM_RESPONSE",
	10202: "S2C_TEAMS_RESPONSE",
	10300: "S2C_ITEMS_SYNC",
	10301: "S2C_ITEMS_UPDATE",
	10302: "C2S_ITEM_FUSION_REQUEST",
	10303: "S2C_ITEM_FUSION_RESPONSE",
	10304: "C2S_ITEM_SELL_REQUEST",
	10305: "S2C_ITEM_SELL_RESPONSE",
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
}
var MSGID_value = map[string]int32{
	"NONE":                                    0,
	"C2S_TEST_COMMAND":                        1,
	"C2S_HEARTBEAT":                           2,
	"S2C_STATE_NOTIFY":                        3,
	"C2S_LOGIN_REQUEST":                       10000,
	"S2C_LOGIN_RESPONSE":                      10001,
	"S2C_OTHER_PLACE_LOGIN":                   10002,
	"C2S_SELECT_SERVER_REQUEST":               10003,
	"S2C_SELECT_SERVER_RESPONSE":              10004,
	"C2S_ENTER_GAME_REQUEST":                  11000,
	"S2C_ENTER_GAME_RESPONSE":                 11001,
	"S2C_ENTER_GAME_COMPLETE_NOTIFY":          11002,
	"C2S_LEAVE_GAME_REQUEST":                  11003,
	"S2C_LEAVE_GAME_RESPONSE":                 11004,
	"S2C_PLAYER_INFO_RESPONSE":                11005,
	"S2C_ROLES_RESPONSE":                      11050,
	"S2C_ROLES_CHANGE_NOTIFY":                 11151,
	"C2S_ROLE_LEVELUP_REQUEST":                11152,
	"S2C_ROLE_LEVELUP_RESPONSE":               11153,
	"C2S_ROLE_RANKUP_REQUEST":                 11154,
	"S2C_ROLE_RANKUP_RESPONSE":                11155,
	"C2S_ROLE_DECOMPOSE_REQUEST":              11156,
	"S2C_ROLE_DECOMPOSE_RESPONSE":             11157,
	"C2S_ROLE_FUSION_REQUEST":                 11158,
	"S2C_ROLE_FUSION_RESPONSE":                11159,
	"C2S_BATTLE_RESULT_REQUEST":               10100,
	"S2C_BATTLE_RESULT_RESPONSE":              10101,
	"C2S_SET_TEAM_REQUEST":                    10200,
	"S2C_SET_TEAM_RESPONSE":                   10201,
	"S2C_TEAMS_RESPONSE":                      10202,
	"S2C_ITEMS_SYNC":                          10300,
	"S2C_ITEMS_UPDATE":                        10301,
	"C2S_ITEM_FUSION_REQUEST":                 10302,
	"S2C_ITEM_FUSION_RESPONSE":                10303,
	"C2S_ITEM_SELL_REQUEST":                   10304,
	"S2C_ITEM_SELL_RESPONSE":                  10305,
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
	// 699 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x6c, 0x95, 0x49, 0x6e, 0x14, 0x31,
	0x14, 0x86, 0x99, 0x85, 0x2c, 0x40, 0x15, 0x93, 0x81, 0x24, 0x10, 0x10, 0x41, 0x84, 0x49, 0x59,
	0x84, 0x13, 0x38, 0x55, 0x2f, 0xdd, 0x25, 0x5c, 0xb6, 0xdb, 0x76, 0x75, 0x2b, 0x2b, 0x8b, 0x21,
	0x8a, 0x22, 0x11, 0x82, 0x48, 0x8e, 0xc0, 0x3c, 0x08, 0x48, 0x02, 0x6c, 0x19, 0x37, 0x2c, 0x38,
	0x01, 0xe3, 0x2d, 0x80, 0x73, 0x80, 0xc4, 0x28, 0x21, 0x77, 0xdb, 0xae, 0x81, 0xac, 0x22, 0xf9,
	0x7f, 0xdf, 0x57, 0xae, 0xf7, 0x5e, 0xa5, 0xd1, 0xd0, 0xf9, 0x8b, 0x0b, 0x73, 0x97, 0x56, 0xcc,
	0xe2, 0xdc, 0xf2, 0xf2, 0xd9, 0xf9, 0x39, 0xb3, 0x70, 0x61, 0xf2, 0xf2, 0x95, 0xa5, 0x95, 0x25,
	0x3c, 0xb0, 0xb8, 0x3c, 0x3f, 0xf9, 0x5f, 0x78, 0xe2, 0xcd, 0x2e, 0xb4, 0x3d, 0x53, 0x8d, 0x34,
	0xc1, 0x3b, 0xd1, 0x36, 0xc6, 0x19, 0x44, 0x9b, 0x70, 0x3f, 0x8a, 0xe2, 0x29, 0x65, 0x34, 0x28,
	0x6d, 0x62, 0x9e, 0x65, 0x84, 0x25, 0xd1, 0x66, 0xdc, 0x87, 0x76, 0xdb, 0xd3, 0x26, 0x10, 0xa9,
	0xa7, 0x81, 0xe8, 0x68, 0x8b, 0x2d, 0x54, 0x53, 0xb1, 0x51, 0x9a, 0x68, 0x30, 0x8c, 0xeb, 0x74,
	0x66, 0x36, 0xda, 0x8a, 0x07, 0x51, 0x9f, 0x2d, 0xa4, 0xbc, 0x91, 0x32, 0x23, 0xa1, 0x95, 0x83,
	0xd2, 0xd1, 0x03, 0x86, 0x87, 0x10, 0xb6, 0xd5, 0xfe, 0x5c, 0x09, 0xce, 0x14, 0x44, 0x0f, 0x19,
	0x1e, 0x41, 0x03, 0x36, 0xe0, 0xba, 0x09, 0xd2, 0x08, 0x4a, 0x62, 0xe8, 0x15, 0x45, 0xab, 0x0c,
	0x8f, 0xa1, 0x61, 0x2b, 0x53, 0x40, 0x21, 0xd6, 0x46, 0x81, 0x6c, 0x83, 0x0c, 0xd2, 0x35, 0x86,
	0x0f, 0xa2, 0x91, 0xee, 0x15, 0x6a, 0xb9, 0x93, 0xaf, 0x33, 0x3c, 0x8a, 0x06, 0xad, 0x00, 0x98,
	0x06, 0x69, 0x1a, 0x24, 0x83, 0x40, 0xff, 0xc8, 0xf1, 0x7e, 0x34, 0x64, 0xe9, 0x4a, 0xe8, 0xd0,
	0x9f, 0x39, 0x1e, 0x47, 0x63, 0xb5, 0x34, 0xe6, 0x99, 0xa0, 0x50, 0xbc, 0xec, 0xaf, 0xdc, 0xfb,
	0x29, 0x90, 0x36, 0x54, 0xfd, 0xbf, 0x83, 0xbf, 0x12, 0x3a, 0xff, 0x9f, 0x1c, 0x1f, 0x40, 0xfb,
	0x6c, 0x2a, 0x28, 0x99, 0x05, 0x69, 0x52, 0x36, 0xc3, 0x8b, 0xf8, 0x6f, 0xee, 0xfb, 0x25, 0x39,
	0x05, 0x55, 0x04, 0xaf, 0xdb, 0xde, 0xda, 0x0b, 0xe2, 0x26, 0x61, 0x8d, 0x70, 0xa1, 0xfb, 0x1d,
	0x6b, 0xb5, 0x17, 0xb2, 0xa9, 0xa1, 0xd0, 0x06, 0x9a, 0x8b, 0x62, 0x0a, 0x1d, 0xdb, 0x50, 0x0f,
	0x97, 0x62, 0x3f, 0x8c, 0x8e, 0x95, 0x07, 0x5c, 0x12, 0x76, 0xa6, 0x44, 0xaf, 0x76, 0xfc, 0x95,
	0xab, 0xa9, 0x83, 0xd7, 0x3a, 0x76, 0x1a, 0x01, 0x4e, 0xc0, 0x76, 0x8b, 0xab, 0xa2, 0x21, 0xeb,
	0x1d, 0x7c, 0x08, 0x8d, 0x06, 0xbe, 0x5c, 0xe0, 0x14, 0x8f, 0xaa, 0xcf, 0x9f, 0xc9, 0x55, 0xca,
	0x8b, 0x1d, 0x7a, 0x5c, 0x7d, 0x7e, 0x48, 0x1d, 0xfc, 0xa4, 0xe3, 0xb7, 0x65, 0x9a, 0x68, 0x4d,
	0xbb, 0xda, 0x9c, 0xea, 0x80, 0x7f, 0x0b, 0xdb, 0x52, 0xcf, 0x9d, 0xe0, 0x3b, 0xc3, 0xc3, 0xa8,
	0xbf, 0xb7, 0x6e, 0xda, 0x68, 0x20, 0x59, 0x60, 0x3f, 0x73, 0xbf, 0xa5, 0xa5, 0xc8, 0x61, 0x5f,
	0xb8, 0x1f, 0x95, 0x3d, 0x2f, 0x8d, 0xea, 0x2b, 0xc7, 0x7b, 0xd1, 0x1e, 0x1b, 0xa4, 0x1a, 0x32,
	0x65, 0xd4, 0x2c, 0x8b, 0xa3, 0xb7, 0x02, 0x0f, 0xf4, 0x3e, 0x9b, 0xde, 0x61, 0x2e, 0x12, 0xa2,
	0x21, 0x7a, 0x27, 0xfc, 0x9b, 0xdb, 0xe3, 0xfa, 0x9b, 0xbf, 0x17, 0xfe, 0xcd, 0xab, 0xa9, 0x7b,
	0xd0, 0x07, 0x61, 0x6f, 0x17, 0x60, 0x05, 0x94, 0x06, 0xf4, 0xa3, 0xb0, 0x2b, 0x1a, 0x50, 0x97,
	0x39, 0xf0, 0x93, 0xf0, 0x2d, 0x8b, 0x49, 0x26, 0x48, 0xda, 0x60, 0x26, 0x21, 0x9a, 0x04, 0xf8,
	0x69, 0xcb, 0xb7, 0xac, 0x9e, 0x3b, 0xc1, 0xb3, 0x16, 0x9e, 0x40, 0x87, 0x2b, 0x02, 0xbb, 0x8f,
	0xb9, 0x30, 0x29, 0x8b, 0x79, 0xe9, 0x63, 0x78, 0xde, 0xc2, 0xc7, 0xd0, 0x78, 0xc5, 0x54, 0x2f,
	0x74, 0xca, 0x17, 0x2d, 0x7c, 0x12, 0x1d, 0x2d, 0x8d, 0xd1, 0x76, 0xdc, 0xd5, 0x06, 0xd6, 0x6b,
	0x5f, 0xb6, 0xf0, 0x29, 0x34, 0x51, 0x9a, 0xe9, 0xc6, 0xc5, 0x4e, 0xfd, 0xaa, 0xe5, 0xfb, 0x94,
	0x91, 0x94, 0x1a, 0x05, 0x2c, 0x09, 0xa6, 0xab, 0xd2, 0xf7, 0xa9, 0x9c, 0x39, 0xf0, 0x9a, 0xac,
	0x80, 0x34, 0x55, 0xc5, 0x5a, 0x5d, 0xaf, 0x82, 0x2e, 0x73, 0xe0, 0x0d, 0xe9, 0xc7, 0xda, 0x0d,
	0x13, 0xd0, 0xf6, 0x8f, 0x47, 0x6f, 0x4a, 0x3f, 0xd6, 0x6a, 0xea, 0xe0, 0x5b, 0xd2, 0xf6, 0x2c,
	0xc0, 0x0d, 0xd0, 0x86, 0x68, 0x4d, 0xe2, 0x26, 0x24, 0x6e, 0x79, 0xbc, 0xe8, 0xb6, 0xc4, 0xc7,
	0xd1, 0x91, 0x20, 0xda, 0xb0, 0xd2, 0x49, 0xef, 0xd4, 0x6f, 0xd4, 0xfd, 0x7f, 0xe6, 0x45, 0x77,
	0xeb, 0x37, 0x72, 0xa9, 0x83, 0xef, 0xc9, 0x73, 0x3b, 0xba, 0x3f, 0x27, 0xa7, 0xff, 0x05, 0x00,
	0x00, 0xff, 0xff, 0x21, 0xbc, 0x0a, 0x52, 0x69, 0x06, 0x00, 0x00,
}
