// MIT License
//
// Copyright (c) 2024 chaunsin
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//

package eapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/chaunsin/netease-cloud-music/api/types"
)

type V3SongDetailReq struct {
}

// V3SongDetailResp
// see: https://github.com/Binaryify/NeteaseCloudMusicApi/issues/1121#issuecomment-774438040
// name: String, 歌曲标题
// id: u64, 歌曲ID
// pst: 0，功能未知
// t: enum,
//
//	0: 一般类型
//	1: 通过云盘上传的音乐，网易云不存在公开对应
//	  如果没有权限将不可用，除了歌曲长度以外大部分信息都为null。
//	  可以通过 `/api/v1/playlist/manipulate/tracks` 接口添加到播放列表。
//	  如果添加到“我喜欢的音乐”，则仅自己可见，除了长度意外各种信息均为未知，且无法播放。
//	  如果添加到一般播放列表，虽然返回code 200，但是并没有效果。
//	  网页端打开会看到404画面。
//	  属于这种歌曲的例子: https://music.163.com/song/1345937107
//	2: 通过云盘上传的音乐，网易云存在公开对应
//	  如果没有权限则只能看到信息，但无法直接获取到文件。
//	  可以通过 `/api/v1/playlist/manipulate/tracks` 接口添加到播放列表。
//	  如果添加到“我喜欢的音乐”，则仅自己可见，且无法播放。
//	  如果添加到一般播放列表，则自己会看到显示“云盘文件”，且云盘会多出其对应的网易云公开歌曲。其他人看到的是其对应的网易云公开歌曲。
//	  网页端打开会看到404画面。
//	  属于这种歌曲的例子: https://music.163.com/song/435005015
//
// ar: Vec<Artist>, 歌手列表
// alia: Vec<String>,
//
//	别名列表，第一个别名会被显示作副标题
//	例子: https://music.163.com/song/536623501
//
// pop: 小数，常取[0.0, 100.0]中离散的几个数值, 表示歌曲热度
// st: 0: 功能未知
// rt: Option<String>, None、空白字串、或者类似`600902000007902089`的字符串，功能未知
// fee: enum,
//
//	0: 免费
//	1: 2元购买单曲
//	4: 购买专辑
//	8: 低音质免费
//
// v: u64, 常为[1, ?]任意数字, 功能未知
// crbt: Option<String>, None或字符串表示的十六进制，功能未知
// cf: Option<String>, 空白字串或者None，功能未知
// al: Album, 专辑，如果是DJ节目(dj_type != 0)或者无专辑信息(single == 1)，则专辑id为0
// dt: u64, 功能未知
// h: Option<Quality>, 高质量文件信息
// m: Option<Quality>, 中质量文件信息
// l: Option<Quality>, 低质量文件信息
// a: Option<?>, 常为None, 功能未知
// cd: Option<String>, None或如"04", "1/2", "3", "null"的字符串，表示歌曲属于专辑中第几张CD，对应音频文件的Tag
// no: u32, 表示歌曲属于CD中第几曲，0表示没有这个字段，对应音频文件的Tag
// rtUrl: Option<String(?)>, 常为None, 功能未知
// rtUrls: Vec<String(?)>, 常为空列表, 功能未知
// dj_id: u64,
//
//	0: 不是DJ节目
//	其他：是DJ节目，表示DJ ID
//
// copyright: u32, 0, 1, 2: 功能未知
// s_id: u64, 对于t == 2的歌曲，表示匹配到的公开版本歌曲ID
// mark: u64, 功能未知
// originCoverType: enum
//
//	0: 未知
//	1: 原曲
//	2: 翻唱
//
// originSongSimpleData: Option<SongSimpleData>, 对于翻唱曲，可选提供原曲简单格式的信息
// single: enum,
//
//	0: 有专辑信息或者是DJ节目
//	1: 未知专辑
//
// noCopyrightRcmd: Option<NoCopyrightRcmd>, None表示可以播，非空表示无版权
// mv: u64, 非零表示有MV ID
// rtype: 常为0，功能未知
// rurl: Option<String(?)>, 常为None，功能未知
// mst: u32, 偶尔为0, 常为9，功能未知
// cp: u64, 功能未知
// publish_time: i64, 毫秒为单位的Unix时间戳
type V3SongDetailResp struct {
	types.RespCommon[any]
	Playlist PlaylistRespList `json:"playlist"`
	Version  string           `json:"version"` // 时间戳1703557080686
}

// V3SongDetail 未知
// url: https://app.apifox.com/project/3870894
func (a *Api) V3SongDetail(ctx context.Context, req *V3SongDetailReq) (*V3SongDetailResp, error) {
	var (
		url   = "https://music.163.com/eapi/v3/song/detail"
		reply V3SongDetailResp
	)
	resp, err := a.client.Request(ctx, http.MethodPost, url, "eapi", req, &reply)
	if err != nil {
		return nil, fmt.Errorf("Request: %w", err)
	}
	_ = resp
	return &reply, nil
}
