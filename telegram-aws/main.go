package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/yaml.v3"
)

type Config struct {
	AWS AwsInfo `yaml:"aws"`
	TG  TgInfo  `yaml:"tg"`
}
type AwsInfo struct {
	AwsAccessKeyId     string `yaml:"awsAccessKeyId"`
	AwsSecretAccessKey string `yaml:"awsSecretAccessKey"`
}
type TgInfo struct {
	Token string `yaml:"token"`
}

// NetworkUsageResponse 包含网络传输额度信息的结构
type NetworkUsageResponse struct {
	LigthsailName string  `json:"vpsName"`
	NetworkIn     float64 `json:"networkIn"`
	NetworkOut    float64 `json:"networkOut"`
	NetworkTotal  string  `json:"networkTotal"`
}

// 获取本月的第一天
func GetCurrentMonthFirstDayZeroTime() *time.Time {

	currentTime := time.Now()
	firstDayOfMonth := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())
	// return fmt.Sprintf(firstDayOfMonth.Format("2006-01-02"))
	return &firstDayOfMonth

}

// func getCurrentMonthLastDayLastTime() string {
// 	// 获取当前时间
// 	currentTime := time.Now()

// 	// 获取本月的最后一天
// 	lastDayOfMonth := time.Date(currentTime.Year(), currentTime.Month()+1, 0, 0, 0, 0, 0, currentTime.Location())

// 	return fmt.Sprintf(lastDayOfMonth.Format("2006-01-02"))
// }

// 获取实例列表
func ListInstance(client *lightsail.Client) []string {
	var InstanceNames []string
	instances, err := client.GetInstances(context.TODO(), &lightsail.GetInstancesInput{})
	if err != nil {
		fmt.Println("Error get instances: ", err)
	}
	for _, instance := range instances.Instances {
		InstanceNames = append(InstanceNames, *instance.Name)
	}
	return InstanceNames
}

// 获取实例的Metric数据
func GetInstanceDataUsage(client *lightsail.Client, vpsName string, metricType types.InstanceMetricName) float64 {
	input := &lightsail.GetInstanceMetricDataInput{
		EndTime:      aws.Time(time.Now()),
		InstanceName: aws.String(vpsName),
		MetricName:   metricType,
		Period:       aws.Int32(6 * 600 * 24),
		StartTime:    aws.Time(*GetCurrentMonthFirstDayZeroTime()),
		Statistics:   []types.MetricStatistic{types.MetricStatisticSum},
		Unit:         types.MetricUnitBytes,
	}

	result, err := client.GetInstanceMetricData(context.TODO(), input)
	if err != nil {
		fmt.Println("Error getting instance metric data:", err)
		return 0
	}
	var total float64
	for _, data := range result.MetricData {

		total = total + float64(*data.Sum)

	}
	return total

}

// 获取实例的网络流量使用情况
func GetLightsailNetworkUsage(awsAccessKeyId, awsSecretAccessKey string) NetworkUsageResponse {

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKeyId, awsSecretAccessKey, "")), config.WithRegion("ap-southeast-1"))
	if err != nil {
		log.Fatal("Error config: ", err)
	}
	client := lightsail.NewFromConfig(cfg)

	instanceNameList := ListInstance(client)

	// 获取networkIn networkOut
	networkIn := GetInstanceDataUsage(client, instanceNameList[0], types.InstanceMetricNameNetworkIn)
	networkOut := GetInstanceDataUsage(client, instanceNameList[0], types.InstanceMetricNameNetworkOut)
	networkTotal := fmt.Sprintf("%.1fG", (networkIn+networkOut)/1024/1024/1024)

	// 构建响应结构
	response := NetworkUsageResponse{
		LigthsailName: instanceNameList[0],
		NetworkIn:     networkIn,
		NetworkOut:    networkOut,
		NetworkTotal:  networkTotal,
	}

	return response
}

func GetConfig() (conf Config, err error) {
	config := &Config{}
	data, err := os.ReadFile("./config.yaml")
	if err != nil {
		fmt.Println("Error open the file: ", err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return *config, err
	}
	return *config, nil
}

func main() {

	conf, err := GetConfig()
	if err != nil {
		fmt.Println("err get config: ", err)
	}

	bot, err := tgbotapi.NewBotAPI(conf.TG.Token)
	if err != nil {
		panic(err)
	}
	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go handleUpdate(&wg, bot, updates, conf)
	}
	wg.Wait()

}

// 处理TG的信息返回
func handleUpdate(wg *sync.WaitGroup, bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, conf Config) {

	defer wg.Done()

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		chatID := update.Message.Chat.ID

		replyMsg := tgbotapi.NewMessage(chatID, "")
		switch update.Message.Command() {
		case "help":
			replyMsg.Text = "I understand /sayhi and /usage."
		case "sayhi":
			replyMsg.Text = "Hi :) I am your father!"
		case "usage":
			netWorkInfo := GetLightsailNetworkUsage(conf.AWS.AwsAccessKeyId, conf.AWS.AwsSecretAccessKey)
			log.Printf("API [GetLightsailNetworkUsage] 被调用")
			netWorkInfoText := fmt.Sprintf(" Name: %s\n NetworkIn: %.1f \n NetworkOut: %.1f \n Total: %s",
				netWorkInfo.LigthsailName, netWorkInfo.NetworkIn, netWorkInfo.NetworkOut, netWorkInfo.NetworkTotal)
			replyMsg.Text = netWorkInfoText
		default:
			replyMsg.Text = "I don't know that command"
		}

		_, _ = bot.Send(replyMsg)

	}

}
