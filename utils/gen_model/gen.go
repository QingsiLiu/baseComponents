// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	// 命令行参数
	dsn := flag.String("dsn", "", "MySQL连接字符串，例如：user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local")
	outPath := flag.String("outpath", "./model", "生成的模型代码输出目录")
	tables := flag.String("tables", "", "要生成的表名，多个表用逗号分隔，为空则生成所有表")
	packageName := flag.String("package", "model", "生成的模型代码包名")

	flag.Parse()

	// 验证参数
	if *dsn == "" {
		fmt.Println("必须提供MySQL连接字符串，使用 -dsn 参数")
		flag.Usage()
		os.Exit(1)
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(*dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 创建代码生成器
	g := gen.NewGenerator(gen.Config{
		OutPath:           *outPath,             // 输出目录
		ModelPkgPath:      *packageName,         // 模型包名
		Mode:              gen.WithDefaultQuery, // 默认查询模式
		FieldNullable:     true,                 // 支持数据库null值
		FieldWithTypeTag:  true,                 // 添加类型标签
		FieldWithIndexTag: true,                 // 添加索引标签
	})

	// 设置数据库连接
	g.UseDB(db)

	// 根据表名生成模型
	tableList := []string{}
	if *tables != "" {
		// 用户指定了表
		tableList = parseTableNames(*tables)
	} else {
		// 获取数据库中的所有表
		var dbTables []string
		if err := db.Raw("SHOW TABLES").Scan(&dbTables).Error; err != nil {
			log.Fatalf("获取数据库表失败: %v", err)
		}
		tableList = dbTables
	}

	// 生成模型
	if len(tableList) > 0 {
		log.Printf("开始生成以下表的模型: %v", tableList)
		for _, tableName := range tableList {
			g.GenerateModel(tableName)
		}
	} else {
		log.Println("未找到要生成的表")
	}

	// 执行代码生成
	g.Execute()

	log.Println("模型生成完成！输出目录:", *outPath)
}

// 解析表名列表
func parseTableNames(tables string) []string {
	if tables == "" {
		return []string{}
	}

	var result []string
	// 简单的逗号分隔解析
	current := ""
	for i := 0; i < len(tables); i++ {
		if tables[i] == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(tables[i])
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}
