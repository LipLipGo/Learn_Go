package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{}) // 初始化db

	if err != nil {
		panic(err)
	}

	db = db.Debug() // 可输出语句

	// 初始化你的表结构
	db.AutoMigrate(&Product{}) // 根据 Product 结构生成语句

	// create 新增数据
	db.Create(&Product{Code: "D42", Price: 65})

	// Read 读取数据

	var product Product
	db.First(&product, 1)                 // 根据整型主键查找
	db.First(&product, "code = ?", "D42") // 查找 code 字段值为 “D42” 的记录

	// Update 修改数据

	db.Model(&product).Update("Price", 200)
	// 更新多个字段
	db.Model(&product).Updates(Product{Code: "F42", Price: 100}) // 仅更新非零字段	接收数据类型为接口，传入不同数据，行为不同，灵活，但可读性差
	db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete 删除
	db.Delete(&product, 1)

}
