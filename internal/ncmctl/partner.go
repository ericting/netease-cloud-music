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

package ncmctl

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/chaunsin/netease-cloud-music/api"
	"github.com/chaunsin/netease-cloud-music/api/types"
	"github.com/chaunsin/netease-cloud-music/api/weapi"
	"github.com/chaunsin/netease-cloud-music/pkg/log"
	"github.com/chaunsin/netease-cloud-music/pkg/utils"

	"github.com/spf13/cobra"
)

type PartnerOpts struct {
	Star []int64
}

type Partner struct {
	root *Root
	cmd  *cobra.Command
	opts PartnerOpts
	l    *log.Logger
}

func NewPartner(root *Root, l *log.Logger) *Partner {
	c := &Partner{
		root: root,
		l:    l,
		cmd: &cobra.Command{
			Use:     "partner",
			Short:   "[need login] Executive music partner daily reviews",
			Example: `  ncmctl partner`,
		},
	}
	c.addFlags()
	c.cmd.Run = func(cmd *cobra.Command, args []string) {
		if err := c.execute(cmd.Context()); err != nil {
			cmd.Println(err)
		}
	}
	return c
}

func (c *Partner) addFlags() {
	c.cmd.PersistentFlags().Int64SliceVarP(&c.opts.Star, "star", "s", []int64{3, 4}, "star level range 1-5")
}

func (c *Partner) validate() error {
	if len(c.opts.Star) == 0 || len(c.opts.Star) > 5 {
		return fmt.Errorf("star level must be range 1-5")
	}
	if !utils.IsUnique(c.opts.Star) {
		return fmt.Errorf("star level must be unique")
	}
	return nil
}

func (c *Partner) Add(command ...*cobra.Command) {
	c.cmd.AddCommand(command...)
}

func (c *Partner) Command() *cobra.Command {
	return c.cmd
}

func (c *Partner) execute(ctx context.Context) error {
	if err := c.validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	if err := c.do(ctx); err != nil {
		c.cmd.Println("job:", err)
		return err
	}
	c.cmd.Printf("%s execute success\n", time.Now())
	return nil
}

func (c *Partner) do(ctx context.Context) error {
	cli, err := api.NewClient(c.root.Cfg.Network, c.l)
	if err != nil {
		return fmt.Errorf("NewClient: %w", err)
	}
	defer cli.Close(ctx)

	// 判断是否需要登录
	request := weapi.New(cli)
	if request.NeedLogin(ctx) {
		return fmt.Errorf("need login")
	}

	// 判断是否有音乐合伙人资格
	info, err := request.PartnerUserinfo(ctx, &weapi.PartnerUserinfoReq{ReqCommon: types.ReqCommon{}})
	if err != nil {
		return fmt.Errorf("PartnerUserinfo: %w", err)
	}
	if info.Code == 703 {
		return fmt.Errorf("您不是音乐合伙人不能进行测评 detail: %+v\n", info)
	}
	if info.Code != 200 {
		return fmt.Errorf("PartnerUserinfo err: %+v\n", info)
	}
	// TODO:状态需要明确
	if info.Data.Status == "ELIMINATED" {
		return fmt.Errorf("您没有测评资格或失去测评资格 status: %s\n", info.Data.Status)
	}

	// 获取每日基本任务5首歌曲列表并执行测评
	task, err := request.PartnerDailyTask(ctx, &weapi.PartnerTaskReq{ReqCommon: types.ReqCommon{}})
	if err != nil {
		return fmt.Errorf("PartnerDailyTask: %w", err)
	}
	for _, work := range task.Data.Works {
		// 判断任务是否执行过
		if work.Completed {
			log.Warn("task completed: %+v\n", work)
			continue
		}

		// 模拟听歌消耗得时间,随机15-25秒
		time.Sleep(time.Second * time.Duration(15+int(rand.Int31n(10))))

		// 随机一个分数,然后从对应分数组中取一个tag
		star := c.opts.Star[rand.Int31n(int32(len(c.opts.Star)))]
		group := weapi.PartnerTagsGroup[star]
		tags := group[rand.Int31n(int32(len(group)))]

		// 上报听歌事件

		// 执行测评
		var req = &weapi.PartnerEvaluateReq{
			ReqCommon:     types.ReqCommon{},
			TaskId:        task.Data.Id,
			WorkId:        work.Work.Id,
			Score:         star,
			Tags:          tags,
			CustomTags:    "[]",
			Comment:       "",
			SyncYunCircle: false,
			SyncComment:   true,               // ?
			Source:        "mp-music-partner", // 定死的值？
		}
		resp, err := request.PartnerEvaluate(ctx, req)
		if err != nil {
			return fmt.Errorf("PartnerEvaluate: %w", err)
		}
		switch resp.Code {
		case 200:
			// ok
		case 405:
			// 当前任务歌曲已完成评
		default:
			log.Error("PartnerEvaluate(%+v) err: %+v\n", req, resp)
			// return fmt.Errorf("PartnerEvaluate: %v", resp.Message)
		}
	}

	// 获取扩展任务列表并执行扩展任务测评 2024年10月21日推出的新功能测评
	var (
		taskId   = task.Data.Id
		extraNum = 2 + rand.Int31n(6) // 扩展歌曲每日7首歌会给一分
	)
	extraTask, err := request.PartnerExtraTask(ctx, &weapi.PartnerExtraTaskReq{ReqCommon: types.ReqCommon{}})
	if err != nil {
		return fmt.Errorf("PartnerExtraTask: %w", err)
	}
	for _, work := range extraTask.Data {
		// 判断任务是否执行过
		if work.Completed {
			log.Warn("extra task completed: %+v\n", work)
			continue
		}

		// 模拟听歌消耗得时间,随机15-25秒
		time.Sleep(time.Second * time.Duration(15+int(rand.Int31n(10))))

		// 随机一个分数,然后从对应分数组中取一个tag
		star := c.opts.Star[rand.Int31n(int32(len(c.opts.Star)))]
		group := weapi.PartnerTagsGroup[star]
		tags := group[rand.Int31n(int32(len(group)))]

		// 上报听歌事件

		// 上报
		var req = &weapi.PartnerExtraReportReq{
			ReqCommon:     types.ReqCommon{},
			WorkId:        work.Work.Id,
			ResourceId:    work.Work.ResourceId,
			BizResourceId: "",
			InteractType:  "PLAY_END",
		}
		resp, err := request.PartnerExtraReport(ctx, req)
		if err != nil {
			return fmt.Errorf("PartnerExtraReport: %w", err)
		}
		switch resp.Code {
		case 200:
			// ok
		default:
			log.Error("PartnerExtraReport(%+v) err: %+v\n", req, resp)
			continue
		}

		// 执行测评
		var evaluateReq = &weapi.PartnerEvaluateReq{
			ReqCommon:     types.ReqCommon{},
			TaskId:        taskId,
			WorkId:        work.Work.Id,
			Score:         star,
			Tags:          tags,
			CustomTags:    "[]",
			Comment:       "",
			SyncYunCircle: false,
			SyncComment:   true,               // ?
			Source:        "mp-music-partner", // 定死的值？
			ExtraResource: true,
		}
		evaluateResp, err := request.PartnerEvaluate(ctx, evaluateReq)
		if err != nil {
			return fmt.Errorf("PartnerEvaluate: %w", err)
		}
		switch evaluateResp.Code {
		case 200:
			extraNum--
			if extraNum <= 0 {
				goto end
			}
		case 405:
			// 当前任务歌曲已完成评
		default:
			log.Error("PartnerEvaluate(%+v) err: %+v\n", req, resp)
			// return fmt.Errorf("PartnerEvaluate: %v", resp.Message)
		}
	}
end:

	// 刷新token过期时间
	refresh, err := request.TokenRefresh(ctx, &weapi.TokenRefreshReq{})
	if err != nil || refresh.Code != 200 {
		log.Warn("TokenRefresh resp:%+v err: %s", refresh, err)
	}
	return nil
}
