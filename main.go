package main

import (
    "log"
    "os"
    "time"

    "github.com/tucnak/telebot"
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/sqlite"
    "R2D2/model"
    "fmt"
    "strings"
    "strconv"
)

func NotifyAll(db *gorm.DB, bot *telebot.Bot) {
    //从周一到周六的早上八点，给所有人发送点餐信息
    flag := true
    for {
        h, _, _ := time.Now().Clock();
        if (time.Now().Weekday() != 7) {   //周日不点
            if h == 15 && flag {
                var allRestaurant []model.Restaurant
                var allPeople []model.Person

                flag = false
                now := time.Now().Day()
                model.Db.Find(&allRestaurant)
                model.Db.Find(&allPeople)
                restaurant := allRestaurant[now % len(allRestaurant)]
                var foods []model.Food
                model.Db.Model(&restaurant).Related(&foods)
                Y, M, D := time.Now().Date()
                s := fmt.Sprintf("今天是 %d 年 %d 月 %d 日, 今天的早餐店是 %s\n",
                    Y, M, D, restaurant.Name)
                for index, food := range foods{
                    s += fmt.Sprintf("%d %s %.1f\n", index, food.Name, food.Price)
                }
                for _, person := range allPeople {
                    bot.SendMessage(&telebot.Chat{ID:int64(person.ChatId)}, s, nil)
                }
            } else if h == 10 && !flag{
                flag = true
            }
        }
        time.Sleep(10 * time.Second)
    }
}

func main() {
    bot, err := telebot.NewBot(os.Getenv("BOT_TOKEN"))
    if err != nil {
        log.Fatalln(err)
    }
    model.Db, err = gorm.Open("sqlite3", "test.db")
    if err != nil {
        log.Fatal(err)
        panic("failed to connect database")
    }
    defer model.Db.Close()

    messages := make(chan telebot.Message, 100)
    bot.Listen(messages, 1*time.Second)
    go NotifyAll(model.Db, bot)

    for message := range messages {
        if message.Text == "/ping" {
            //测试机器人是否还活着
            bot.SendMessage(message.Chat, "pong", nil)
        }

        if message.Text == "/login" {
            //新员工加入
            var person model.Person
            chatId := message.Sender.ID
            model.Db.Where("chat_id = ?", chatId).First(&person)

            if person.ID > 0 {
                person.Active = true
                model.Db.Save(&person)
            }else {
                firstName := message.Sender.FirstName
                model.Db.Create(&model.Person{ChatId:chatId, FirstName:firstName, Active:true})
            }
            bot.SendMessage(message.Chat, "欢迎加入黑湖早餐军团", nil)
        }

        if message.Text == "/logout" {
            var person model.Person
            chatId := message.Sender.ID
            model.Db.Where("chat_id = ?", chatId).First(&person)
            if person.ID > 0 {
                person.Active = false
                model.Db.Save(&person)
                bot.SendMessage(message.Chat, "还是要认真吃早餐哦", nil)
            }else {
                bot.SendMessage(message.Chat, "你还不在黑湖早餐军团哦", nil)
            }
        }

        if message.Text == "/order" {
            //看今日菜单，比如说周日加班，但是机器人没有发单，可以通过这个命令查看今日菜单
            now := time.Now().Day()
            var allRestaurant []model.Restaurant
            model.Db.Find(&allRestaurant)
            restaurant := allRestaurant[now % len(allRestaurant)]
            var foods []model.Food
            model.Db.Model(&restaurant).Related(&foods)
            s := fmt.Sprintf("今天的早餐店是 %s\n", restaurant.Name)
            for index, food := range foods{
                s += fmt.Sprintf("%d %s %.1f\n", index, food.Name, food.Price)
            }
            bot.SendMessage(message.Chat, s, nil)
        }

        if message.Text == "/query" {
            //最终点餐的人查看点餐总计，再去饿了么上面点餐
            var allRestaurant []model.Restaurant
            model.Db.Find(&allRestaurant)
            now := time.Now()
            restaurant := allRestaurant[now.Day() % len(allRestaurant)]
            var allOrders model.Orders
            yesterday := now.AddDate(0,0,-1)
            model.Db.Where("created_at > ?", yesterday).Find(&allOrders)
            s := fmt.Sprintf("今天的早餐店是 %s\n", restaurant.Name)
            for k,v := range allOrders.GroupByFoodName() {
                s += fmt.Sprintf("%s * %d\n", k, v)
            }
            bot.SendMessage(message.Chat, s, nil)
        }
        if message.Text == "/help" {
            s := "/help 查看帮助\n/order 查看今日早餐菜单\n/query 查看今日所有订单\n/login 新员工入职加入订餐\n/logout 员工退出或者放假啥的\n/ping 查看机器人是否活着"
            bot.SendMessage(message.Chat, s, nil)
        }
        if strings.HasPrefix(message.Text, "order") {
            //点餐命令，格式为 order 1*2 0*1, a*b 中 a为食物编号，b为数量
            //简化点餐命令， order 0 1 省略数量，默认为1
            now := time.Now().Day()
            var allRestaurant []model.Restaurant
            model.Db.Find(&allRestaurant)
            restaurant := allRestaurant[now % len(allRestaurant)]
            var foods []model.Food
            model.Db.Model(&restaurant).Related(&foods)
            msg := strings.Split(message.Text, " ")[1:]
            s := "订餐成功\n"
            for _, f := range msg {
                var food model.Food
                var amount int64
                if strings.Contains(f, "*") {
                    tmp := strings.Split(f, "*")
                    foodIdx, _ := strconv.ParseInt(tmp[0],0,32)
                    amount, _ = strconv.ParseInt(tmp[1],0,64)
                    food = foods[foodIdx]
                    model.Db.Create(&model.Order{UserId:1,RestaurantId:restaurant.ID,FoodId:food.ID,Amount:amount})
                } else {
                    foodIdx, _ := strconv.ParseInt(f,0,32)
                    food = foods[foodIdx]
                    amount = 1
                    model.Db.Create(&model.Order{UserId:1,RestaurantId:restaurant.ID,FoodId:food.ID,Amount:amount})
                }
                s += fmt.Sprintf("%s, %d\n", food.Name, amount)
            }
            bot.SendMessage(message.Chat, s, nil)
        }
    }
}
