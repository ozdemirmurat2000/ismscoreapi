package main

import (
	"context"
	"encoding/json"
	"ismscoreapi/myModels"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-collections/collections/set"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DATABASE BILGILERI
const (
	MongoDBHost = "mongodb+srv://ozdemirmurat5sj:100200345@cluster0.vfb3in6.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
	DBName      = "saves"
	Collection  = "saves_details"
)

type HaveRedCard struct {

	MatchID string
	PlayerName string
	
}

var client *mongo.Client
func main() {

	// DATABASE BAGLAN
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(MongoDBHost))
	if err != nil {
		log.Fatal(err)
	}

	// HTTP SUNUCU BASLAT
	app := fiber.New()

	app.Post("/add_match",saveDataForUser)
	app.Delete("/delete_match/:match_id/:device_id", deleteMatchHandler)

	port := ":3000"
	log.Printf("HTTP sunucusu %s portunda başlatıldı\n", port)

	// API URL
	apiURL := "https://apiv3.apifootball.com?action=get_events&APIkey=0bb2e1fcd01fe076d54ae77d3acfe2a57353820b668d5efb837e5167b7cb1f8d&match_live=1&timezone=Europe/Istanbul&withPlayerStats=1"



	// APIYE ATILAN ISTEK SURESI
	go func() {
		for {
			
			fetchDataFromAPI(apiURL)
		}
	}()

	// SERVERI SUREKLI ACIK TUT
	app.Listen("0.0.0.0"+port)

	select {}
}

// CANLI MAC VERILERINI AL
func fetchDataFromAPI(apiURL string) {

	log.Println("maclar alindi")



	// BIRINCI LISTEYI AL 
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Println("API'ye istek gönderilirken hata oluştu:", err)
		return
	}
	defer resp.Body.Close()

	
	var matches []myModels.SubMatch
	if err := json.NewDecoder(resp.Body).Decode(&matches); err != nil {

		log.Println("hata olustu",err)
	}	

	time.Sleep(1 * time.Second)

	// IKINCI LISTEYI AL
	resp2, err2 := http.Get(apiURL)
	if err2 != nil {
		log.Println("API'ye istek gönderilirken hata oluştu:", err2)
		return
	}
	defer resp2.Body.Close()

	
	var matches2 []myModels.SubMatch
	if err2 := json.NewDecoder(resp2.Body).Decode(&matches2); err2 != nil {

		log.Println("hata olustu",err2)
	}

	 listGoalNew := set.New()
	 listGoalOld := set.New()

	

	for _, v := range matches {
	
		if v.MatchStatus !="Finished" {
			for _, v1 := range v.Goalscorer {

				listGoalNew.Insert(DiffGoal{
					
					Time: v1.Time,
					HomeScorer: v1.HomeScorer,
					AwayScorer: v1.AwayScorer,
					MatchID: v.MatchID,
					HomeTeamName: v.MatchHometeamName,
					AwayTeamName: v.MatchAwayteamName,
					MatchStatus: v.MatchStatus,
					
				
				})
			}
	
		}

	}
	for _, v := range matches2 {
	
		if v.MatchStatus != "Finished" {
			for _, v1 := range v.Goalscorer {
				listGoalOld.Insert(DiffGoal{
					
					Time: v1.Time,
					HomeScorer: v1.HomeScorer,
					AwayScorer: v1.AwayScorer,
					MatchID: v.MatchID,
					HomeTeamName: v.MatchHometeamName,
					AwayTeamName: v.MatchAwayteamName,
					MatchStatus: v.MatchStatus,

					
				
				})
			}
		}


	}

	diff := listGoalNew.Difference(listGoalOld)

	diff.Do(func(i interface{}) {
		goal, ok := i.(DiffGoal)
		if !ok {
			// Dönüşüm başarısız ise hata mesajı yazdır
			log.Println("Hata: Goal modeline dönüştürme başarısız")
			return
		}

		if goal.MatchStatus == goal.Time {
			

			data, err := getDataByMatchID(goal.MatchID)
			if err != nil {
				log.Println("Veri alınamadı:", err)
				return
			}
			
			for _, item := range data {
				log.Println("Maç ID:", item.MatchID, "Device ID:", item.DeviceID,"------> BU CIHAZA BILDIRIM GONDERILDI")
			}
			
		}


	})
		

	
}
type DiffGoal struct {
	MatchStatus	  string
	MatchID 	  string
	HomeTeamName  string
	AwayTeamName  string
	Time          string
	HomeScorer    string 
	AwayScorer    string
}

type ControlModel struct {

	MatchID string
	GoalScorer []interface{}
	Cards []interface{}
}

func deleteMatchInList(model ControlModel,silinecekItem []myModels.SubMatch) {

	for index, result := range silinecekItem {

		if result.MatchID == model.MatchID {
			silinecekItem = append(silinecekItem[:index], silinecekItem[index+1:]...)

		}


		
	}



	
}

	/// VERIYI SIL  
func deleteMatchHandler(c *fiber.Ctx) error {
	// PARAMETRELERI AL
	matchID := c.Params("match_id")
	deviceID := c.Params("device_id")

	// DATABASE DEN VERILERI SIL
	collection := client.Database(DBName).Collection(Collection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := collection.DeleteOne(ctx, bson.M{"match_id": matchID, "device_id": deviceID})
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Veri silinirken hata oluştu: " + err.Error())
	}

	// SILINEN VERIYI GOSTER
	return c.JSON(result)
}
// DATABASEDEN UYUSAN VERIYI BUL

func getDataByMatchID(matchID string) ([]DbDATA, error) {
    // DATABASE'DEN VERIYI GETIR
    collection := client.Database(DBName).Collection(Collection)
    
    var data []DbDATA
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    cursor, err := collection.Find(ctx, bson.M{"match_id": matchID})
    if err != nil {
        log.Println("Veri bulunamadı:", err)
        return nil, err
    }
    defer cursor.Close(ctx)

    for cursor.Next(ctx) {
        var item DbDATA
        if err := cursor.Decode(&item); err != nil {
            log.Println("Veri okunamadı:", err)
            return nil, err
        }
        data = append(data, item)
    }

    return data, nil
}

// KULLANICIDAN GELEN VERIYI KAYDET

func saveDataForUser(c *fiber.Ctx) error {
	var data DbDATA
	if err := c.BodyParser(&data); err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Veri silinirken hata oluştu: " + err.Error())
	}

	// DATABASE VERI EKLE
	collection := client.Database(DBName).Collection(Collection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := collection.InsertOne(ctx, data)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Veri silinirken hata oluştu: " + err.Error())
	}

	// EKLENEN VERIYI GOSTER
	return c.JSON(result)
}


// ORNEK VERI 
type DbDATA struct {
	MatchID   string `json:"match_id,omitempty" bson:"match_id,omitempty"`
	DeviceID  string `json:"device_id,omitempty" bson:"device_id,omitempty"`
}

