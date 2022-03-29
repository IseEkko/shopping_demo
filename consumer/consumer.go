package main

import (
	"fmt"
	"imooc-product/common"
	rabbitmq2 "imooc-product/rabbitmq"
	"imooc-product/repositories"
	"imooc-product/services"
)

func main() {
	db, err := common.NewMysqlConn()
	if err != nil {
		fmt.Println(err)
	}
	//创建product数据库操作实例
	product := repositories.NewProductManager("product", db)
	//创建product service
	productService := services.NewProductService(product)
	//创建order数据库实例
	order := repositories.NewOrderMangerRepository("order", db)
	//创建order service
	orderService := services.NewOrderService(order)
	rabbitmqConsumeSimple := rabbitmq2.NewRabbitMQSimple("imoocProduct")
	rabbitmqConsumeSimple.ConsumeSimple(orderService, productService)

}
