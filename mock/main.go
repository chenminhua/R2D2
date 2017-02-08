package main

import (
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
  "log"
  "R2D2/model"
  "io/ioutil"
  "strings"
  "strconv"
)

func main() {
  db, err := gorm.Open("sqlite3", "../test.db")
  if err != nil {
    log.Fatal(err)
    panic("failed to connect database")
  }
  defer db.Close()
  db.DropTable("restaurants")
  db.DropTable("foods")
  db.AutoMigrate(&model.Person{})
  db.AutoMigrate(&model.Restaurant{})
  db.AutoMigrate(&model.Food{})
  db.AutoMigrate(&model.Order{})
  db.Model(&model.Person{}).AddUniqueIndex("idx_person_chatid", "chat_id")

  path := "../data"
  files, _ := ioutil.ReadDir(path)
  for _, file := range files{
    data, err := ioutil.ReadFile(path + "/" + file.Name())
    if err != nil{
      panic(err)
    }
    foods := strings.Split(string(data), "\n")
    var foodList []model.Food
    for _, fdstr := range foods{
      f := strings.Split(fdstr, " ")
      price, _ := strconv.ParseFloat(f[1], 64)
      food := model.Food{Name: f[0], Price: price}
      db.Create(&food)
      foodList = append(foodList, food)
    }
    db.Create(&model.Restaurant{Name:file.Name(), Foods:foodList})
  }
}

