package model

import "github.com/jinzhu/gorm"

type Person struct {
  gorm.Model
  FirstName string
  ChatId int
  Active bool
}

type Restaurant struct {
  gorm.Model
  Name string
  Foods []Food
}

type Food struct {
  gorm.Model
  Name string
  Price float64
  RestaurantId uint
}

type Order struct {
  gorm.Model
  UserId uint
  RestaurantId uint
  FoodId uint
  Amount int64
}

type Orders []Order

var Db *gorm.DB

func (orders Orders) GroupByFoodName() map[string]int64 {
  res := make(map[string]int64)
  for _, order := range orders{
    var food Food
    Db.Where("id = ?", order.FoodId).First(&food)
    res[food.Name] += order.Amount
  }
  return res
}