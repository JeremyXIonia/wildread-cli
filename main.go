package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	dir := flag.String("dir", "./books", "书籍目录")
	dbPath := flag.String("db", "./novel-reader.db", "SQLite 数据库路径")
	flag.Parse()

	fmt.Printf("书籍目录: %s\n", *dir)
	fmt.Printf("数据库: %s\n", *dbPath)

	if _, err := os.Stat(*dir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "目录不存在: %s\n", *dir)
		os.Exit(1)
	}
}
