package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/ximplez/wxread/notify"
	"github.com/ximplez/wxread/utils"
	"github.com/ximplez/wxread/utils/conv"
	"github.com/ximplez/wxread/utils/env_utils"
)

var (
	// 飞书机器人通知链接
	feishuBotUrl string
	// vid
	vid string
	// skey
	skey string

	statService          = NewReadingStatisticService()
	dailyLotteryService  = NewDailyLotteryService()
	readingRewardService = NewReadingRewardService()
)

const (
	envVid          = "VID"
	envFeishuBotUrl = "FEISHU_BOT_URL"
	envSkey         = "SKEY"
)

func main() {
	f := flag.Int64("f", 0, "执行方法: 1-统计截止当天阅读时长数据, 2-每日抽奖, 3-阅读奖励")
	flag.Parse()
	if f == nil || *f <= 0 {
		log.Fatalln("func 非法")
	}
	vid = env_utils.GetEnv(envVid)
	skey = env_utils.GetEnv(envSkey)
	feishuBotUrl = env_utils.GetEnv(envFeishuBotUrl)
	if vid == "" || skey == "" {
		log.Fatalln("vid 或 skey 为空")
	}
	ctx := context.Background()
	switch *f {
	case 1:
		if err := stat(ctx); err != nil {
			log.Fatalf("err: %v", err)
		}
	case 2:
		if err := dailyLottery(ctx); err != nil {
			log.Fatalf("err: %v", err)
		}
	case 3:
		if err := readingReward(ctx); err != nil {
			log.Fatalf("err: %v", err)
		}
	default:
		log.Fatalln("func 非法")
	}
}

// 统计截止当天阅读时长数据
func stat(ctx context.Context) (e error) {
	defer func() {
		if e != nil {
			notify.NotifyFeishu(feishuBotUrl, notify.NewFeishuMsg("微信读书", "❌ 统计截止当天阅读时长数据失败", e.Error(), ""))
		}
	}()
	log.Printf("📊 开始统计今天阅读时长数据")
	statistic, err := statService.GetReadingStatistic(ctx, vid, skey)
	if err != nil {
		return err
	}
	ts := make([]string, 0, len(statistic.ReadTimes))
	for k := range statistic.ReadTimes {
		ts = append(ts, k)
	}
	// 按时间戳倒排
	sort.Slice(ts, func(i, j int) bool {
		return conv.Str2Int64(ts[i]) > conv.Str2Int64(ts[j])
	})
	s := ""
	// 使用中国时区
	loc := time.FixedZone("Asia/Shanghai", 8*3600)
	now := time.Now().In(loc)
	yesterday := now.AddDate(0, 0, -1)

	// 将时间转换为日期字符串进行比较
	nowDateStr := now.Format("2006-01-02")
	yesterdayDateStr := yesterday.Format("2006-01-02")

	for _, k := range ts {
		v := statistic.ReadTimes[k]
		t := time.Unix(conv.Str2Int64(k), 0).In(loc)
		hours := v / 3600
		minutes := (v % 3600) / 60
		seconds := v % 60
		timeStr := ""
		if hours > 0 {
			timeStr += fmt.Sprintf("%d 小时 ", hours)
		}
		if minutes > 0 {
			timeStr += fmt.Sprintf("%d 分钟 ", minutes)
		}
		if seconds > 0 {
			timeStr += fmt.Sprintf("%d 秒", seconds)
		}

		// 将t转换为日期字符串进行比较
		tDateStr := t.Format("2006-01-02")

		if tDateStr == nowDateStr {
			s += fmt.Sprintf(`
	%s：%s`, notify.GreenText(fmt.Sprintf("今天（%s）", utils.FormatWeekdayCN(t))), notify.OrangeText(timeStr))
		} else if tDateStr == yesterdayDateStr {
			s += fmt.Sprintf(`
	%s：%s`, notify.GreenText(fmt.Sprintf("昨天（%s）", utils.FormatWeekdayCN(t))), notify.OrangeText(timeStr))
		} else {
			s += fmt.Sprintf(`
	%s（%s）：%s`, t.Format("2006-01-02"), utils.FormatWeekdayCN(t), notify.OrangeText(timeStr))
		}
	}
	msg := notify.BoldText(fmt.Sprintf(`📊 本周共阅读 %s:%s`, notify.BlueText(strconv.Itoa(statistic.ReadDays)+" 天"), s))

	notify.NotifyFeishu(feishuBotUrl, notify.NewFeishuMsg("微信读书", fmt.Sprintf("📊 %s 统计数据", nowDateStr), msg, ""))
	return nil
}

// 自动每日抽奖
func dailyLottery(ctx context.Context) (e error) {
	defer func() {
		if e != nil {
			notify.NotifyFeishu(feishuBotUrl, notify.NewFeishuMsg("微信读书", "❌ 每日抽奖失败", e.Error(), ""))
		}
	}()
	log.Printf("🎉 开始每日抽奖")
	lottery, err := dailyLotteryService.GetDailyLottery(ctx, vid, skey)
	if err != nil {
		return err
	}
	if lottery.GiftReceived.ReceiveTime > 0 {
		notifyGiftReceived(&lottery.GiftReceived)
		return nil
	}
	log.Printf("🎉 今日未抽,开始抽奖!")
	giftReceived, err := dailyLotteryService.ExecuteDailyLottery(ctx, vid, skey, lottery.Issue)
	if err != nil {
		return err
	}
	notifyGiftReceived(giftReceived)
	return nil
}

func readingReward(ctx context.Context) (e error) {
	totalNum := 0
	receivedNum := 0
	defer func() {
		if e != nil {
			msg := notify.BoldText(notify.PurpleText(fmt.Sprintf(`本次可领取 %d 天体验卡，领取成功 %d 天体验卡！`, totalNum, receivedNum)))
			notify.NotifyFeishu(feishuBotUrl, notify.NewFeishuMsg("微信读书", "❌ 领取阅读奖励失败",
				fmt.Sprintf(`%s
失败信息：%s`, msg, e.Error()), ""))
		}
	}()
	log.Printf("🎁 开始领取阅读奖励")
	reward, err := readingRewardService.GetReadingReward(ctx, vid, skey)
	if err != nil {
		return err
	}
	receivableAward := make([]*Award, 0)
	for _, award := range reward.ReadtimeAwards {
		if award.AwardStatus == awardStatus_Unreceived {
			receivableAward = append(receivableAward, award)
			totalNum += award.AwardChoices[0].AwardNum
		}
	}
	for _, award := range reward.ReaddayAwards {
		if award.AwardStatus == awardStatus_Unreceived {
			receivableAward = append(receivableAward, award)
			totalNum += award.AwardChoices[0].AwardNum
		}
	}
	log.Printf("🎁 【阅读奖励】本次可领取 %d 天体验卡", totalNum)
	if len(receivableAward) != 0 {
		for _, award := range receivableAward {
			err := readingRewardService.ExchangeReadingReward(ctx, vid, skey, award.AwardLevelId, award.AwardChoices[0].ChoiceType)
			if err != nil {
				return err
			}
			receivedNum += award.AwardChoices[0].AwardNum
		}
	}
	log.Printf(fmt.Sprintf(`🎁 【阅读奖励】领取成功 %d 天体验卡`, receivedNum))
	notify.NotifyFeishu(feishuBotUrl, notify.NewFeishuMsg("微信读书", "🎁 领取阅读奖励",
		notify.BoldText(notify.PurpleText(fmt.Sprintf(`本次可领取 %d 天体验卡，领取成功 %d 天体验卡`, totalNum, receivedNum))), ""))
	return nil
}

func notifyGiftReceived(giftReceived *GiftReceived) {
	gift := giftReceived.Name
	if giftReceived.Money > 0 {
		gift = strconv.Itoa(giftReceived.Money/100) + " " + gift
	}
	msg := notify.BoldText(fmt.Sprintf(`🎉 今日已抽 %s`, notify.BlueText(gift)))
	notify.NotifyFeishu(feishuBotUrl, notify.NewFeishuMsg("微信读书", "🎉 每日抽奖", msg, ""))
}
