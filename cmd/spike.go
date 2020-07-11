package main

import (
	"fmt"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_gin"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os"
	"shopping/controllers"
	"shopping/middleware"
	"shopping/models"
	"shopping/repositories"
	"shopping/services"
	"shopping/utils"
)

//基于hash环的分布式权限控制

var (
	//分布式集群地址
	hostList = []string{"127.0.0.1", "127.0.0.2", "127.0.0.3"}
	//端口
	port = "8081"
	//记录现在的秒杀商品的数量
	commodityCache []map[int]int

	//hash环
	consistent utils.ConsistentHashImp
)

func main() {
	consistent = utils.NewConsistent(20)
	for _, v := range hostList {
		consistent.Add(v)
	}
	//缓存所有需要秒杀的商品的库存
	models.Init()
	models.MysqlHandler.AutoMigrate(models.Order{})
	repository := &repositories.CommodityRepository{Db: models.MysqlHandler}
	service := &services.CommodityService{CommodityRepository: repository}
	commodityList, err := service.GetCommodityAll()
	if err != nil {
		utils.Log.WithFields(log.Fields{"errMsg": err.Error()}).Panic("缓存所有需要秒杀的商品的库存，获取库存失败")
		os.Exit(1)
		return
	}
	tmpInfo := make(map[int]int)
	for _, value := range *commodityList {
		tmpInfo[int(value.ID)] = value.Stock
	}

	app := gin.Default()
	ip, err := utils.GetIp()
	if err != nil {
		utils.Log.WithFields(log.Fields{"errMsg": err.Error()}).Panic("ip获取失败")
		os.Exit(1)
		return
	}

	simple := services.NewRabbitMQSimple("myxy99Shopping")

	spikeService := &services.SpikeService{
		Consistent:       consistent,
		LocalHost:        ip,
		HostList:         hostList,
		Port:             port,
		CommodityCache:   tmpInfo,
		RabbitMqValidate: simple,
	}

	spikeController := &controllers.SpikeController{SpikeService: spikeService} //, middleware.Auth()

	limiter := tollbooth.NewLimiter(1, nil)
	app.GET("/spike/:commodityId", middleware.Auth(), spikeController.Shopping)

	app.GET("/", tollbooth_gin.LimitHandler(limiter), func(context *gin.Context) {
		context.JSON(200, gin.H{"data": 1})
	})

	_ = app.Run(fmt.Sprintf(":%v", port))
}