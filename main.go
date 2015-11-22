package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MyJsonName2 struct {
	Prices []struct {
		Minimum      int     `json:"minimum"`
		Low_Estimate float64 `json:"low_estimate"`
		ProductID    string  `json:"product_id" bson:"product_id"`
		Distance     float64 `json:"distance"`
		Duration     int     `json:"duration"`
		VisitedFlag  bool
	} `json:"prices"`
}

type MyJsonNameArray struct {
	structArray []MyJsonName
}

type MyInput struct {
	Starting_from_location_id string   `json:"starting_from_location_id"`
	Location_ids              []string `json:"location_ids"`
}
type UberRequest struct {
	Id                           bson.ObjectId   `json:"_id" bson:"_id"`
	Status                       string          `json:"status"`
	Starting_from_location_id    bson.ObjectId   `json:"Starting_from_location_id" bson:"Starting_from_location_id"`
	Next_destination_location_id bson.ObjectId   `json:"Next_destination_location_id" bson:"Next_destination_location_id"`
	Best_route_location_ids      []bson.ObjectId `json:"Best_route_location_ids"`
	Total_uber_costs             float64         `json:"Total_uber_costs"`
	Total_uber_duration          int             `json:"Total_uber_duration"`
	Total_distance               float64         `json:"Total_distance"`
	Uber_wait_time_eta           int             `json:"uber_wait_time_eta"`
	CurrentIndex                 int             `json:"CurrentIndex"`
}

type JsonCoordinates struct {
	ProductID string  `json:"product_id" bson:"product_id"`
	StartLat  float64 `json:"start_latitude" bson:"start_latitude"`
	StartLng  float64 `json:"start_longitude" bson:"start_longitude"`
	EndLat    float64 `json:"end_latitude" bson:"end_latitude"`
	EndLng    float64 `json:"end_longitude" bson:"end_longitude"`
}

type EstimateTime struct {
	Eta int `json:"eta"`
}

type MyJsonResult struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string   `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

type MyJsonName struct {
	Id         bson.ObjectId `json:"id" bson:"_id"`
	Name       string        `json:"name"`
	Address    string        `json:"address"`
	City       string        `json:"city"`
	State      string        `json:"state"`
	Zip        string        `json:"zip"`
	Coordinate struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"coordinate"`
}

func GetTrips(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	fmt.Println("Hello world")
	session, err := mgo.Dial("mongodb://user1:pass1@ds045054.mongolab.com:45054/mydatabase")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("mydatabase").C("UberResult")

	id := p.ByName("name")
	Oid := bson.ObjectIdHex(id)
	var UberGetResult UberRequest
	c.FindId(Oid).One(&UberGetResult)

	if err != nil {
		log.Fatal(err)
	}

	b2, err := json.Marshal(UberGetResult)
	if err != nil {
	}
	rw.WriteHeader(http.StatusOK)

	fmt.Fprintf(rw, string(b2))
}

func Postlocations(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	var myjson3 MyJsonName
	s3 := json.NewDecoder(req.Body)
	err := s3.Decode(&myjson3)
	StartQuery := "http://maps.google.com/maps/api/geocode/json?address="
	WhereQuery := myjson3.Address + " " + myjson3.City + " " + myjson3.State
	WhereQuery = strings.Replace(WhereQuery, " ", "+", -1)
	EndQuery := "&sensor=false"
	Url1 := StartQuery + WhereQuery + EndQuery
	fmt.Println("Published URL: " + Url1)
	res, err := http.Get(Url1)
	if err != nil {
		log.Fatal(err)
	}
	robots, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	var myjsonresult1 MyJsonResult
	err = json.Unmarshal(robots, &myjsonresult1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(myjsonresult1.Results[0].Geometry.Location.Lat)
	fmt.Println(myjsonresult1.Results[0].Geometry.Location.Lng)

	myjson3.Id = bson.NewObjectId()

	myjson3.Coordinate.Lat = myjsonresult1.Results[0].Geometry.Location.Lat
	myjson3.Coordinate.Lng = myjsonresult1.Results[0].Geometry.Location.Lng

	if err != nil {
	}

	session, err := mgo.Dial("mongodb://user1:pass1@ds045054.mongolab.com:45054/mydatabase")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	c := session.DB("mydatabase").C("people")

	err = c.Insert(myjson3)
	if err != nil {
		log.Fatal(err)
	}

	result := MyJsonName{}
	id := myjson3.Id.Hex()
	oid := bson.ObjectIdHex(id)
	c.FindId(oid).One(&result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("New Name:", result.Name)
	fmt.Println("Address:", result.Address)
	fmt.Println("Id2:", result.Id.String())
	oid = bson.ObjectId(result.Id)

	b2, err := json.Marshal(result)
	if err != nil {
	}
	rw.WriteHeader(http.StatusCreated)

	fmt.Fprintf(rw, string(b2))
}

func PostTrips(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	var Myinput1 MyInput
	s3 := json.NewDecoder(req.Body)
	err := s3.Decode(&Myinput1)
	fmt.Println(Myinput1)
	fmt.Println(Myinput1.Starting_from_location_id)
	fmt.Println(Myinput1.Location_ids[1])
	fmt.Println(len(Myinput1.Location_ids))
	//	session, err := mgo.Dial("localhost")
	session, err := mgo.Dial("mongodb://user1:pass1@ds045054.mongolab.com:45054/mydatabase")
	if err != nil {
		panic("Connection error")
	}
	defer session.Close()

	fmt.Println("one")

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	c := session.DB("mydatabase").C("people")
	UberSession := session.DB("mydatabase").C("UberResult")

	//get originaldata
	var Arraybest_route_location_ids []bson.ObjectId
	Arraybest_route_location_ids = make([]bson.ObjectId, len(Myinput1.Location_ids))
	fmt.Println("point1")

	var StartingPoint MyJsonName
	LocationResultarray := make([]MyJsonName, len(Myinput1.Location_ids))
	UberResultarray := make([]MyJsonName2, len(Myinput1.Location_ids))

	Total_uber_costs := 0.0
	Total_uber_duration := 0
	Total_distance := 0.0

	var MinCount float64
	MinCount = 100
	MinCountIndex := 0
	fmt.Println("point2")
	Totalleft := len(Myinput1.Location_ids)
	//Start of outer loop
	for OuterIndex := 0; OuterIndex < len(Myinput1.Location_ids); OuterIndex++ {
		fmt.Println("MyUberArrayOuterLoop :")
		fmt.Println(OuterIndex)

		if OuterIndex == 0 {
			id := Myinput1.Starting_from_location_id
			oid := bson.ObjectIdHex(id)
			c.FindId(oid).One(&StartingPoint)
		} else {
			oid := Arraybest_route_location_ids[OuterIndex-1]
			c.FindId(oid).One(&StartingPoint)
		}

		fmt.Println("StartingPointttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttt: ")
		fmt.Println(OuterIndex)
		fmt.Println(StartingPoint)
		MinCount = 100
		MinCountIndex = 0

		//Start of inner loop
		for InnerIndex := 0; InnerIndex < Totalleft; InnerIndex++ {
			fmt.Println("Here1")
			id := Myinput1.Location_ids[InnerIndex]
			fmt.Println("Here2")
			oid := bson.ObjectIdHex(id)
			fmt.Println("Here3")
			c.FindId(oid).One(&LocationResultarray[InnerIndex])
			fmt.Println(LocationResultarray[InnerIndex])

			var ServerToken = "ZfFnTv1EKoy6SwKeHMuecmxy2IL8coZe-n5zC6No"
			Url2 := "https://sandbox-api.uber.com/v1/estimates/price?"
			Appendwhere := "start_latitude=" + strconv.FormatFloat(StartingPoint.Coordinate.Lat, 'f', 6, 64) + "&start_longitude=" + strconv.FormatFloat(StartingPoint.Coordinate.Lng, 'f', 6, 64) + "&end_latitude=" + strconv.FormatFloat(LocationResultarray[InnerIndex].Coordinate.Lat, 'f', 6, 64) + "&end_longitude=" + strconv.FormatFloat(LocationResultarray[InnerIndex].Coordinate.Lng, 'f', 6, 64) + "&server_token=" + ServerToken
			Url2 += Appendwhere
			fmt.Println("Published URL: " + Url2)
			res, err := http.Get(Url2)
			if err != nil {
				log.Fatal(err)
			}
			robots, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
			//var MyUberResultFirst MyJsonName2
			err = json.Unmarshal(robots, &UberResultarray[InnerIndex])

			fmt.Println("MyUberArray :")
			fmt.Println(InnerIndex)
			fmt.Println(UberResultarray[InnerIndex])
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("UberResultarray[InnerIndex].Prices[0].Low_Estimate")
			fmt.Println(UberResultarray[InnerIndex].Prices[0].Low_Estimate)

			fmt.Println("End of iteration 1")

			fmt.Println("MinCount: ")
			fmt.Println(MinCount)

			fmt.Println("UberResultarray[index].Prices[0].Low_Estimate")
			fmt.Println(UberResultarray[InnerIndex].Prices[0].Low_Estimate)

			if MinCount > UberResultarray[InnerIndex].Prices[0].Low_Estimate {
				MinCount = UberResultarray[InnerIndex].Prices[0].Low_Estimate
				MinCountIndex = InnerIndex
				fmt.Println("MinCount: ")
				fmt.Println(MinCount)
				fmt.Println("MinCountIndex: ")
				fmt.Println(MinCountIndex)
			}

			fmt.Println("End of iteration 2")

		} //End of innner loop

		Total_uber_costs += UberResultarray[MinCountIndex].Prices[0].Low_Estimate
		Total_uber_duration += UberResultarray[MinCountIndex].Prices[0].Duration
		Total_distance += UberResultarray[MinCountIndex].Prices[0].Distance

		fmt.Println("LocationResultarray+++++++++++++++++++++++++++++++++++++++++++++++")
		fmt.Println(OuterIndex)
		fmt.Println(LocationResultarray)

		Arraybest_route_location_ids[OuterIndex] = LocationResultarray[MinCountIndex].Id

		LocationResultarray[MinCountIndex] = LocationResultarray[len(LocationResultarray)-(OuterIndex+1)]
		UberResultarray[MinCountIndex] = UberResultarray[len(LocationResultarray)-(OuterIndex+1)]
		//LocationResultarray[MinCountIndex] = LocationResultarray[Totalleft-1]
		Myinput1.Location_ids[MinCountIndex] = Myinput1.Location_ids[len(LocationResultarray)-(OuterIndex+1)]
		fmt.Println("Totalleft:")
		fmt.Println(Totalleft)
		fmt.Println("MinCountIndex:")
		fmt.Println(MinCountIndex)

		Totalleft = Totalleft - 1
		MinCountIndex = 0

	} //End of Outer loop
	var LastPoint MyJsonName
	c.FindId(Arraybest_route_location_ids[len(Arraybest_route_location_ids)-1]).One(&StartingPoint)
	c.FindId(bson.ObjectIdHex(Myinput1.Starting_from_location_id)).One(&LastPoint)

	var ServerToken = "ZfFnTv1EKoy6SwKeHMuecmxy2IL8coZe-n5zC6No"
	Url2 := "https://sandbox-api.uber.com/v1/estimates/price?"
	Appendwhere := "start_latitude=" + strconv.FormatFloat(StartingPoint.Coordinate.Lat, 'f', 6, 64) + "&start_longitude=" + strconv.FormatFloat(StartingPoint.Coordinate.Lng, 'f', 6, 64) + "&end_latitude=" + strconv.FormatFloat(LastPoint.Coordinate.Lat, 'f', 6, 64) + "&end_longitude=" + strconv.FormatFloat(LastPoint.Coordinate.Lng, 'f', 6, 64) + "&server_token=" + ServerToken
	Url2 += Appendwhere
	res, err := http.Get(Url2)
	if err != nil {
		log.Fatal(err)
	}
	robots, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	var MyNewUberResultFirst MyJsonName2
	err = json.Unmarshal(robots, &MyNewUberResultFirst)

	Total_uber_costs = Total_uber_costs + MyNewUberResultFirst.Prices[0].Low_Estimate
	Total_uber_duration += MyNewUberResultFirst.Prices[0].Duration
	Total_distance += MyNewUberResultFirst.Prices[0].Distance

	var MyUberResult UberRequest

	MyUberResult.Total_uber_costs = Total_uber_costs

	MyUberResult.Total_uber_duration = Total_uber_duration
	MyUberResult.Total_distance = Total_distance
	MyUberResult.Best_route_location_ids = Arraybest_route_location_ids
	MyUberResult.Status = "Planning"
	MyUberResult.Starting_from_location_id = bson.ObjectIdHex(Myinput1.Starting_from_location_id)
	MyUberResult.Id = bson.NewObjectId()
	MyUberResult.Next_destination_location_id = MyUberResult.Best_route_location_ids[0]

	err = UberSession.Insert(MyUberResult)

	if err != nil {
		fmt.Println("Error while inserting record")
		fmt.Println(err)
	} else {
		fmt.Println("Inserted Successfully")
	}
	Oid2 := MyUberResult.Id
	fmt.Println("Oid2")
	fmt.Println(Oid2)
	var UberGetResult UberRequest
	c.FindId(Oid2).One(&UberGetResult)

	b2, err := json.Marshal(MyUberResult)
	if err != nil {
	}
	rw.WriteHeader(http.StatusOK)

	fmt.Fprintf(rw, string(b2))

}

func PutTrips(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	session, err := mgo.Dial("mongodb://user1:pass1@ds045054.mongolab.com:45054/mydatabase")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("mydatabase").C("UberResult")
	Cpeople := session.DB("mydatabase").C("people")

	id := p.ByName("name")
	Oidz := bson.ObjectIdHex(id)
	var UberGetResult UberRequest
	c.FindId(Oidz).One(&UberGetResult)

	if err != nil {
		log.Fatal(err)
	}

	if UberGetResult.Next_destination_location_id == bson.ObjectId("") {
		UberGetResult.Next_destination_location_id = UberGetResult.Best_route_location_ids[0]
		fmt.Println("Setting: ", UberGetResult.Next_destination_location_id.String())

	} else {
		for i := 0; i < len(UberGetResult.Best_route_location_ids)-1; i++ {
			if UberGetResult.Best_route_location_ids[0] == UberGetResult.Next_destination_location_id {
				UberGetResult.Next_destination_location_id = UberGetResult.Best_route_location_ids[i]
			} //if
		} //for
	} //else

	var Starting MyJsonName
	var NextPoint1 MyJsonName
	fmt.Println("Printing id")
	if UberGetResult.CurrentIndex == 0 {
		fmt.Println(UberGetResult.Starting_from_location_id)
		Cpeople.FindId(UberGetResult.Starting_from_location_id).One(&Starting)
		Cpeople.FindId(UberGetResult.Best_route_location_ids[0]).One(&NextPoint1)
		UberGetResult.Status = "Planning"
		UberGetResult.Next_destination_location_id = UberGetResult.Best_route_location_ids[0]
	} else if UberGetResult.CurrentIndex == len(UberGetResult.Best_route_location_ids) {
		Cpeople.FindId(UberGetResult.Best_route_location_ids[UberGetResult.CurrentIndex-1]).One(&Starting)
		Cpeople.FindId(UberGetResult.Starting_from_location_id).One(&NextPoint1)
		UberGetResult.Next_destination_location_id = UberGetResult.Starting_from_location_id
		UberGetResult.Status = "Finished"
	} else {
		Cpeople.FindId(UberGetResult.Best_route_location_ids[UberGetResult.CurrentIndex-1]).One(&Starting)
		Cpeople.FindId(UberGetResult.Best_route_location_ids[UberGetResult.CurrentIndex]).One(&NextPoint1)
		UberGetResult.Next_destination_location_id = UberGetResult.Best_route_location_ids[UberGetResult.CurrentIndex]
		UberGetResult.Status = "Planning"
	}

	var ServerToken = "ZfFnTv1EKoy6SwKeHMuecmxy2IL8coZe-n5zC6No"
	Url2 := "https://sandbox-api.uber.com/v1/estimates/price?"
	Appendwhere := "start_latitude=" + strconv.FormatFloat(Starting.Coordinate.Lat, 'f', 6, 64) + "&start_longitude=" + strconv.FormatFloat(Starting.Coordinate.Lng, 'f', 6, 64) + "&end_latitude=" + strconv.FormatFloat(NextPoint1.Coordinate.Lat, 'f', 6, 64) + "&end_longitude=" + strconv.FormatFloat(NextPoint1.Coordinate.Lng, 'f', 6, 64) + "&server_token=" + ServerToken
	Url2 += Appendwhere
	res, err := http.Get(Url2)
	if err != nil {
		log.Fatal(err)
	}
	robots, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	var MyUberResultFirst MyJsonName2
	err = json.Unmarshal(robots, &MyUberResultFirst)

	apiUrl := "https://sandbox-api.uber.com/v1/requests"

	JsonParam := JsonCoordinates{}
	JsonParam.ProductID = MyUberResultFirst.Prices[0].ProductID
	JsonParam.StartLat = Starting.Coordinate.Lat
	JsonParam.StartLng = Starting.Coordinate.Lng
	JsonParam.EndLat = NextPoint1.Coordinate.Lat
	JsonParam.EndLng = NextPoint1.Coordinate.Lng

	JsonStr, err := json.Marshal(JsonParam)
	if err != nil {
		fmt.Println("UBER Error")
	}

	req, err = http.NewRequest("POST", apiUrl, bytes.NewBuffer(JsonStr))
	if err != nil {
		fmt.Println("UBER Error")
	}

	req.Header.Add("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzY29wZXMiOlsicmVxdWVzdCJdLCJzdWIiOiJmNzZhYzhlMy03ZDhkLTQwMmEtODhkOC05ODEyMGM5YmYwMzkiLCJpc3MiOiJ1YmVyLXVzMSIsImp0aSI6IjRhMjFjNjkzLWNjNzItNGI0NC1hYTIwLWExY2I0ZDc5OWVmZiIsImV4cCI6MTQ1MDMzOTY2MSwiaWF0IjoxNDQ3NzQ3NjYxLCJ1YWN0IjoiZDZmVUZjMXlyOVVLUmxLWW81Z0JETDFHSXpSMXJtIiwibmJmIjoxNDQ3NzQ3NTcxLCJhdWQiOiJtaVJpS0FUMmxDcE54akRfUVRWSTFndUhObHhKby1LXyJ9.oIdJ9SYSZcF6xHotRYMKk_Bs1H6R_F0L7hiE13sCHZtGeZ4LuJGz7XC-sjOSkhh8i2c5TBakqu78S_t2APYjsIvTXUvnTY3UnKa33_23LMe6_9oMtjvPBRyaSK1sAiY0d7ig-6LbX80t8IsrO5Iy9yQ7ckeRlU7WWO1MG7Xk8hXsQ0xnVxC7m88_DZza6f46I94GGlBU6YaS8nYgK5OL5LWoYPR0bHCBXowfZQI4VflsgHPgkdFzG2uI8v5-RZkNeTnY5UEokwWbBWGMe5JozXv-JFybePdNR9jwSqKfu_ojTrkCe07fOEI4Qtgjk_w73OTux6pjBY16ZdZ2o5yHKg")

	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("UBER Error")
	}

	bodys, errs := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if errs != nil {
		fmt.Println("UBER Error")
	}

	var TimeEst EstimateTime
	errs = json.Unmarshal(bodys, &TimeEst)
	if errs != nil {
		fmt.Println("UBER Decoding error Partially unmarshalled")
	}
	UberGetResult.Uber_wait_time_eta = TimeEst.Eta

	UberGetResult.CurrentIndex = UberGetResult.CurrentIndex + 1
	if UberGetResult.CurrentIndex > len(UberGetResult.Best_route_location_ids) {
		UberGetResult.CurrentIndex = 0
	}

	c.UpdateId(Oidz, UberGetResult)

	b2, err := json.Marshal(UberGetResult)
	if err != nil {
	}
	rw.WriteHeader(http.StatusCreated)

	fmt.Fprintf(rw, string(b2))

}

func Getlocations(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	session, err := mgo.Dial("mongodb://user1:pass1@ds045054.mongolab.com:45054/mydatabase")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("mydatabase").C("people")
	id := p.ByName("name")
	oid := bson.ObjectIdHex(id)
	var result MyJsonName
	c.FindId(oid).One(&result)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Id2:", result.Id.String())
	oid = bson.ObjectId(result.Id)

	b2, err := json.Marshal(result)
	if err != nil {
	}
	rw.WriteHeader(http.StatusOK)

	fmt.Fprintf(rw, string(b2))
	fmt.Println("Method Name: " + req.Method)
}

func main() {

	mux := httprouter.New()
	mux.GET("/locations/:name", Getlocations)
	mux.GET("/trips/:name", GetTrips)
	mux.POST("/trips", PostTrips)
	mux.PUT("/trips/:name", PutTrips)
	mux.POST("/locations", Postlocations)
	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
