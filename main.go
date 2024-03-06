package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"ismscoreapi/myModels"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
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

var client *mongo.Client



func main() {
	oldMatches := make(map[string]myModels.SubMatch)
	log.Println(oldMatches)
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

	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	}
	log.Printf("HTTP sunucusu %s portunda başlatıldı\n", port)

	// API URL
	apiURL := "https://apiv3.apifootball.com?action=get_events&APIkey=0bb2e1fcd01fe076d54ae77d3acfe2a57353820b668d5efb837e5167b7cb1f8d&match_live=1&timezone=Europe/Istanbul&withPlayerStats=1"

	

	// APIYE ATILAN ISTEK SURESI
	go func() {
		for {
			fetchDataFromAPI(apiURL,oldMatches)
		}
	}()

	

	// SERVERI SUREKLI ACIK TUT
	app.Listen("0.0.0.0"+port)
	
	select {}
}


// CANLI MAC VERILERINI AL
func fetchDataFromAPI(apiURL string,oldMatches map[string]myModels.SubMatch) {

	log.Println("maclar alindi")





	// // BIRINCI LISTEYI AL 
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


	for _, v := range matches {

		// MAC ILK LISTEDE VARMI KONTROL ET
		_,ok := oldMatches[v.MatchID]
		// YOKSA ESKI LISTEYE EKLE
		if !ok {
			// EGER MAC BITMISSE EKLEME

			if v.MatchStatus != "Finished" {
				oldMatches[v.MatchID] = v
				log.Println(v.MatchID, " eklendi")
			}
		
		}else{
			db,dberr :=	getDataByMatchID(v.MatchID)

			if dberr != nil {

				log.Println("mac yeni basladi db hata var",dberr)
				
			}
			// MAC YENI BASLADI 

			if v.MatchStatus == "0" || v.MatchStatus == "1" {

			
			// DB DEKI KULLANICILARI BILDIRIM GONDER
			for _, dbv := range db {
				lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
				if dbv.StartStatus == "0" {
					SendNotificationToUser(dbv.DeviceID,lang["mac_basladi"],bildirimText(v))
					updateDataByMatchIDAndDeviceID(v.MatchID,dbv.DeviceID,DbDATA{StartStatus: "1"})
					log.Println(dbv.DeviceID ,lang["mac_basladi"])

				}
				
			}
			

			}else if v.MatchStatus == "Half Time"{

				// ILK YARI BITMIS BILDIRIM GONDER 

				for _, dbv := range db {
					lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
					if dbv.HalfStatus == "0" {
						SendNotificationToUser(dbv.DeviceID,lang["ilk_yari_bitti"],bildirimText(v))
						updateDataByMatchIDAndDeviceID(v.MatchID,dbv.DeviceID,DbDATA{HalfStatus: "1"})
						log.Println("ilk yari bitti bildirim gonderildi =>",dbv.DeviceID , lang["ilk_yari_bitti"])

					}
					
				}
				

			}else{
				// IKINCI YARI BASLADI BILDIRIM GONDER
				for _, dbv := range db {

					if dbv.SecondStatus == "0" && dbv.HalfStatus == "1" {
						lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
						SendNotificationToUser(dbv.DeviceID,lang["ikinci_yari_basaladi"],bildirimText(v))
						updateDataByMatchIDAndDeviceID(v.MatchID,dbv.DeviceID,DbDATA{SecondStatus: "1"})
						log.Println("ikinci yari basladi bitti bildirim gonderildi =>",dbv.DeviceID,lang["ikinci_yari_basaladi"])


						
					}
					
				}

				// GOL OLAYLARINI TAKIP ET 
				newHomeGoal,_ := strconv.Atoi(checkScore(v.MatchHometeamScore))
				oldHomeGoal,_ := strconv.Atoi(checkScore(oldMatches[v.MatchID].MatchHometeamScore))
				newAwayGoal,_ := strconv.Atoi(checkScore(v.MatchAwayteamScore))
				oldAwayGoal,_ := strconv.Atoi(checkScore(oldMatches[v.MatchID].MatchAwayteamScore))

				if newHomeGoal > oldHomeGoal {
					log.Println("ev sahibi gol atti")
					for _, dbv := range db {
						log.Println(dbv.DeviceID)
						lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
						SendNotificationToUser(dbv.DeviceID, lang["gol"]+"'"+v.MatchStatus+" "+v.MatchHometeamName,bildirimText(v))
						log.Println("gol bildirimi gonderildi =>",dbv.DeviceID,lang["gol"])
					}
				}
				if newHomeGoal < oldHomeGoal {
					log.Println("ev sahibi gol İptal")
					for _, dbv := range db {
						lang,_ :=	getLanguage(dbv.Language)
						SendNotificationToUser(dbv.DeviceID,lang["gol_ipta"] +"'"+v.MatchStatus+" "+v.MatchHometeamName,bildirimText(v))
						log.Println("gol İptal bildirimi gonderildi =>",dbv.DeviceID,lang["gol_ipta"])
					}
				}
				if newAwayGoal > oldAwayGoal {
					
					log.Println("deplasman gol atti")
					for _, dbv := range db {
						log.Println(dbv.DeviceID)

						lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
						SendNotificationToUser(dbv.DeviceID, lang["gol"]+"'"+v.MatchStatus+" "+v.MatchAwayteamName,bildirimText(v))
						log.Println("gol bildirimi gonderildi =>",dbv.DeviceID,lang["gol"])
					}
				}
				if newAwayGoal < oldAwayGoal {
					log.Println("deplasman gol İptal")
					for _, dbv := range db {
						lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
						SendNotificationToUser(dbv.DeviceID,lang["gol_ipta"] +"'"+v.MatchStatus+" "+v.MatchAwayteamName,bildirimText(v))
						log.Println("gol İptal bildirimi gonderildi =>",dbv.DeviceID,lang["gol_ipta"])
					}
				}
				
			}
			// KART VARSA KONTROL ET

			if len(v.Cards) != 0 {


				if len(v.Cards) > len(oldMatches[v.MatchID].Cards) {
				//  KONTROL ET
					if v.Cards[len(v.Cards)-1].Info == "home" && v.Cards[len(v.Cards)-1].Card == "red card"{
						// EV SAHIBI KIRMIZI KART

						for _, dbv := range db {
							lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
							SendNotificationToUser(dbv.DeviceID,lang["kirmizi_kart"] + " " +v.MatchHometeamName + " '" + v.MatchStatus,bildirimText(v))

						}

					}else if v.Cards[len(v.Cards)-1].Info == "away" && v.Cards[len(v.Cards)-1].Card == "red card"{
						for _, dbv := range db {
							lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
							SendNotificationToUser(dbv.DeviceID,lang["kirmizi_kart"] + " " +v.MatchAwayteamName + " '" + v.MatchStatus,bildirimText(v))

						}
					}else{
						if v.Cards[len(v.Cards)-1].HomeFault != "" && v.Cards[len(v.Cards)-1].Card == "red card"{
							// EV SAHIBI KIRMIZI KART
	
							for _, dbv := range db {
								lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
								SendNotificationToUser(dbv.DeviceID,lang["kirmizi_kart"] + " " +v.MatchHometeamName + " " + v.Cards[len(v.Cards)-1].HomeFault + " '" + v.MatchStatus,bildirimText(v))
	
							}
	
						}else if v.Cards[len(v.Cards)-1].AwayFault != "" && v.Cards[len(v.Cards)-1].Card == "red card"{

							for _, dbv := range db {
								lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
								SendNotificationToUser(dbv.DeviceID,lang["kirmizi_kart"] + " " +v.MatchHometeamName + " " + v.Cards[len(v.Cards)-1].AwayFault + " '" + v.MatchStatus,bildirimText(v))
	
							}
						}
					}
				} else if len(v.Cards) < len(oldMatches[v.MatchID].Cards){
					if oldMatches[v.MatchID].Cards[len(oldMatches[v.MatchID].Cards)-1].Info == "home" && oldMatches[v.MatchID].Cards[len(oldMatches[v.MatchID].Cards)-1].Card == "red card"{
						// EV SAHIBI KIRMIZI KART

						for _, dbv := range db {
							lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
							SendNotificationToUser(dbv.DeviceID,lang["kirmizi_kart"] + " " +oldMatches[v.MatchID].MatchHometeamName + " '" + oldMatches[v.MatchID].MatchStatus,bildirimText(oldMatches[v.MatchID]))

						}

					}else if oldMatches[v.MatchID].Cards[len(oldMatches[v.MatchID].Cards)-1].Info == "away" && oldMatches[v.MatchID].Cards[len(oldMatches[v.MatchID].Cards)-1].Card == "red card"{
						for _, dbv := range db {
							lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
							SendNotificationToUser(dbv.DeviceID,lang["kirmizi_kart"] + " " +oldMatches[v.MatchID].MatchAwayteamName + " '" + oldMatches[v.MatchID].MatchStatus,bildirimText(oldMatches[v.MatchID]))

						}
					}else{
						if oldMatches[v.MatchID].Cards[len(oldMatches[v.MatchID].Cards)-1].HomeFault != "" && oldMatches[v.MatchID].Cards[len(oldMatches[v.MatchID].Cards)-1].Card == "red card"{
							// EV SAHIBI KIRMIZI KART
	
							for _, dbv := range db {
								lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
								SendNotificationToUser(dbv.DeviceID,lang["kirmizi_kart"] + " " +oldMatches[v.MatchID].MatchHometeamName + " " + oldMatches[v.MatchID].Cards[len(v.Cards)-1].HomeFault + " '" + oldMatches[v.MatchID].MatchStatus,bildirimText(oldMatches[v.MatchID]))
	
							}
	
						}else if oldMatches[v.MatchID].Cards[len(oldMatches[v.MatchID].Cards)-1].AwayFault != "" && oldMatches[v.MatchID].Cards[len(oldMatches[v.MatchID].Cards)-1].Card == "red card"{

							for _, dbv := range db {
								lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
								SendNotificationToUser(dbv.DeviceID,lang["kirmizi_kart"] + " " +oldMatches[v.MatchID].MatchHometeamName + " " + oldMatches[v.MatchID].Cards[len(oldMatches[v.MatchID].Cards)-1].AwayFault + " '" + v.MatchStatus,bildirimText(oldMatches[v.MatchID]))
	
							}
						}
					}
				}

			}
			
			// MAC BITTIYSE BILDIRIM GONDER
			if v.MatchStatus == "Finished" {

				for _, dbv := range db {
					lan :=	dbv.Language
				lang,_ :=	getLanguage(lan)
					if dbv.EndStatus == "0" {
						log.Println("mac bitti bildirim gonderildi",dbv.DeviceID)
					SendNotificationToUser(dbv.DeviceID,lang["mac_bitti"],bildirimText(v))
					updateDataByMatchIDAndDeviceID(v.MatchID,dbv.DeviceID,DbDATA{EndStatus: "1"})	
					log.Println(lang["mac_bitti"])
					}

					
				}
				// MACI MAPDEN CIKAR
				
				delete(oldMatches,v.MatchID)
				deleteDataByMatchID(v.MatchID)
				
				log.Println("mac silindi :",v.MatchStatus," ",v.MatchID)
				
			}else{
				// MAC VERISINI GUNCELLE 
				oldMatches[v.MatchID] = v
			}

		}

		
		
	}

	

	
}

func checkScore(s string)string  {

	if s == "" {
		return "0"
	}
	return s
	
}

func getLanguage(language string) (map[string]string, error) {
	// JSON dosyasını oku
	jsonFile, err := os.ReadFile("languages/"+language+".json")
	if err != nil {
		return nil, err
	}

	// JSON verisini bir harita olarak ayrıştır
	var jsonData map[string]string
	err = json.Unmarshal(jsonFile, &jsonData)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}




func bildirimText (model myModels.SubMatch) string{

	return model.MatchHometeamName + " " + model.MatchHometeamScore + " - " + model.MatchAwayteamScore + " " + model.MatchAwayteamName

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
    // DATABASE'DEN VERIYI SİL

func deleteDataByMatchID(matchID string) error {
    collection := client.Database(DBName).Collection(Collection)
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    filter := bson.M{"match_id": matchID}
    
    _, err := collection.DeleteMany(ctx, filter)
    if err != nil {
        log.Println("Veri silinemedi:", err)
        return err
    }

    return nil
}


// Belirli bir maç ID'sine sahip belgeleri güncelleyen fonksiyon
func updateDataByMatchIDAndDeviceID(matchID, deviceID string, newData DbDATA) error {
    // DATABASE'DEN VERIYI GÜNCELLE
    collection := client.Database(DBName).Collection(Collection)
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    filter := bson.M{"match_id": matchID, "device_id": deviceID}
    update := bson.M{"$set": newData}
    
    _, err := collection.UpdateOne(ctx, filter, update)
    if err != nil {
        log.Println("Veri güncellenemedi:", err)
        return err
    }

    return nil
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
		return c.Status(http.StatusInternalServerError).SendString("Veri eklenirken hata oluştu: " + err.Error())
	}

	// DATABASE VERI EKLE
	collection := client.Database(DBName).Collection(Collection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ekleme tarihi alanını belirle
	data.CreatedAt = time.Now()
	data.StartStatus = "0"
	data.EndStatus = "0"
	data.HalfStatus = "0"
	data.SecondStatus = "0"

	// Veriyi veritabanına ekle
	result, err := collection.InsertOne(ctx, data)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Veri eklenirken hata oluştu: " + err.Error())
	}

	// EKLENEN VERIYI GOSTER
	return c.JSON(result)
}


// ORNEK VERI 
type DbDATA struct {
	StartStatus string `json:"start_status,omitempty" bson:"start_status,omitempty"`
	SecondStatus string `json:"second_status,omitempty" bson:"second_status,omitempty"`
	HalfStatus string `json:"half_status,omitempty" bson:"half_status,omitempty"`
	EndStatus string	`json:"end_status,omitempty" bson:"end_status,omitempty"`
	MatchID   string `json:"match_id,omitempty" bson:"match_id,omitempty"`
	Language   string `json:"language,omitempty" bson:"language,omitempty"`
	DeviceID  string `json:"device_id,omitempty" bson:"device_id,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
}

func SendNotificationToUser(userID, title, message string) error {
    // OneSignal API bilgileri
    appID := "7808a914-b9b2-45d8-888d-cc4fca3513b2"
    apiKey := "ZWFjNGQ3ZmEtZGQ0ZC00ODY0LWEwN2EtNGI4NWZjZTU4OTI3"

    // OneSignal API endpoint URL'si
    url := "https://onesignal.com/api/v1/notifications"

    // JSON verisini oluştur
    jsonStr := []byte(fmt.Sprintf(`{
        "app_id": "%s",
        "include_player_ids": ["%s"],
        "headings": {"en": "%s"},
        "contents": {"en": "%s"}
    }`, appID, userID, title, message))

    // HTTP isteği oluştur
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
    if err != nil {
        return err
    }

    // İstek başlıklarını ayarla
    req.Header.Set("Content-Type", "application/json; charset=utf-8")
    req.Header.Set("Authorization", "Basic "+apiKey)

    // HTTP isteğini gönder
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // HTTP yanıtını kontrol et
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
    }

    return nil
}
