package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode"

	sm "backend/stringMatching"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Penyakit struct {
	NamaPenyakit string
	DNA          string
}

type HasilPrediksi struct {
	TanggalPrediksi  string
	NamaPasien       string
	PenyakitPrediksi string
	SampelDNA 		 string
	TingkatKemiripan int
	Status           int
}

func getEnv(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}

func openDatabase() *sql.DB {
	// Open database connection.
	db, err := sql.Open("mysql",  getEnv("DATABASE_USERNAME")+":"+ getEnv("DATABASE_PASSWORD")+"@tcp("+ getEnv("DATABASE_PORT")+")/"+ getEnv("DATABASE_NAME"))

	// if there is an error opening the connection, handle it
	if err != nil {
		panic(err.Error())
	}
	return db
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func getDetailPrediction(res http.ResponseWriter, req *http.Request) {
	setupResponse(&res, req)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatalln(err)
	}

	//Convert the body to type string
	string_body := string(body)
	if strings.Split(string_body, ":")[0] != "" {
		// data := strings.Split(string_body, ":")[1]

		data := strings.Replace(string_body,"{", "", -1)
		data = strings.Replace(data,"}", "", -1)
		data = strings.Replace(data,`"`, "", -1)
		data = strings.Split(data,":")[1]
		
		data = strings.Map(func(r rune) rune {
			if unicode.IsPrint(r) {
				return r
			}
			return -1
		}, data)

		db := openDatabase()
		result := []HasilPrediksi{}

		checkData := strings.Split(data, " ")
		if (len(checkData) > 1) {
			db_result, err := db.Query("SELECT * FROM hasilprediksi WHERE TanggalPrediksi='" + checkData[0]+"'")
			if err != nil {
				panic(err.Error())
			}

			for db_result.Next() {
				var hasil HasilPrediksi

				// Get hasil for each row
				err = db_result.Scan(&hasil.TanggalPrediksi, &hasil.NamaPasien, &hasil.PenyakitPrediksi, &hasil.SampelDNA, &hasil.TingkatKemiripan, &hasil.Status)
				if err != nil {
					panic(err.Error()) // proper error handling instead of panic in your app
				}
				// Append hasil to result jika regex sesuai

				if sm.RegexSearch(strings.ToLower(checkData[1]), strings.ToLower(hasil.PenyakitPrediksi)) {
					result = append(result, hasil)
				}
			}
			
		} else {
			if sm.RegexTanggal(checkData[0]) {
				db_result, err := db.Query("SELECT * FROM hasilprediksi WHERE TanggalPrediksi='" + checkData[0]+"'")
				if err != nil {
					panic(err.Error())
				}

				for db_result.Next() {
					var hasil HasilPrediksi

					// Get hasil for each row
					err = db_result.Scan(&hasil.TanggalPrediksi, &hasil.NamaPasien, &hasil.PenyakitPrediksi, &hasil.SampelDNA, &hasil.TingkatKemiripan, &hasil.Status)
					if err != nil {
						panic(err.Error()) // proper error handling instead of panic in your app
					}

					// Append hasil to result

					result = append(result, hasil)
				}
			} else {
				db_result, err := db.Query("SELECT * FROM hasilprediksi")
				if err != nil {
					panic(err.Error())
				}

				for db_result.Next() {
					var hasil HasilPrediksi

					// Get hasil for each row
					err = db_result.Scan(&hasil.TanggalPrediksi, &hasil.NamaPasien, &hasil.PenyakitPrediksi, &hasil.SampelDNA, &hasil.TingkatKemiripan, &hasil.Status)
					if err != nil {
						panic(err.Error()) // proper error handling instead of panic in your app
					}
					// Append hasil to result jika regex sesuai

					if sm.RegexSearch(strings.ToLower(checkData[0]), strings.ToLower(hasil.PenyakitPrediksi)) {
						result = append(result, hasil)
					}
				}
			}
		}

		defer db.Close()

		marshal, err := json.Marshal(result)
		if err != nil {
			fmt.Println(err)
		}
		// Send to frontend
		res.Write(marshal)

	}
}

func getDiseasePrediction(res http.ResponseWriter, req *http.Request) {
	setupResponse(&res, req)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatalln(err)
	}

	string_body := string(body)

	if strings.Split(string_body, ":")[0] != "" {

		db := openDatabase()
		output := []HasilPrediksi{}

		s1 := strings.Split(string_body, "------WebKitFormBoundary")
		s1_dna := strings.Split(s1[1], "\n")
		DNA := s1_dna[4]
		DNA = strings.Map(func(r rune) rune {
			if unicode.IsPrint(r) {
				return r
			}
			return -1
		}, DNA)
		s1_nama := strings.Split(s1[2], "\n")
		Nama := s1_nama[3]
		Nama = strings.Map(func(r rune) rune {
			if unicode.IsPrint(r) {
				return r
			}
			return -1
		}, Nama)
		s1_penyakit := strings.Split(s1[3], "\n")
		Penyakit := s1_penyakit[3]
		Penyakit = strings.Map(func(r rune) rune {
			if unicode.IsPrint(r) {
				return r
			}
			return -1
		}, Penyakit)
		s1_tanggalsplit := strings.Split(s1[4], "\n")
		s1_tanggal := strings.Replace(s1_tanggalsplit[3], ",", "", -1)
		tanggalsplit := strings.Split(s1_tanggal, "/")
		year := tanggalsplit[2][:len(tanggalsplit[2])-1]
		Tanggal := year + "/" + tanggalsplit[0] + "/" + tanggalsplit[1]
		Tanggal = strings.Map(func(r rune) rune {
			if unicode.IsPrint(r) {
				return r
			}
			return -1
		}, Tanggal)

		
		if sm.Regex(DNA) {

			db_result, err := db.Query("SELECT DNA FROM jenispenyakit WHERE NamaPenyakit = '" + Penyakit + "'")
			if err != nil {
				panic(err.Error())
			}
			defer db_result.Close()

			empty := true

			var pDNA string

			for db_result.Next() {
				db_result.Scan(&pDNA)
				empty = false
			}

			if !empty {

				if sm.KMP(pDNA, DNA) && sm.BoyerMoore(DNA, pDNA) {
					db_result, err := db.Query("INSERT INTO hasilprediksi VALUES ('" + Tanggal + "','" + Nama + "','" + Penyakit + "','" + DNA + "',100,1)")

					if err != nil {
						outputisi := HasilPrediksi{
							NamaPasien:       Nama,
							PenyakitPrediksi: Penyakit,
							TanggalPrediksi:  Tanggal,
							TingkatKemiripan: 100,
							Status:           1,
						}
		
						output = append(output, outputisi)
		
						marshal, err := json.Marshal(output)
						if err != nil {
							fmt.Println(err)
						}
		
						res.Write(marshal)
						defer db.Close()
						return

					}
					outputisi := HasilPrediksi{
						NamaPasien:       Nama,
						PenyakitPrediksi: Penyakit,
						TanggalPrediksi:  Tanggal,
						TingkatKemiripan: 100,
						Status:           1,
					}

					defer db_result.Close()
	
					output = append(output, outputisi)
	
					marshal, err := json.Marshal(output)
					if err != nil {
						fmt.Println(err)
					}
	
					res.Write(marshal)
					defer db.Close()
					return
				}

				Percentage := sm.Lcs(DNA, pDNA)
				var stat int

				if Percentage > 80 {
					stat = 1
				} else {
					stat = 0
				}

				log.Printf("%s",DNA)
				db_result, err := db.Query("INSERT INTO hasilprediksi VALUES ('" + Tanggal + "','" + Nama + "','" + Penyakit + "','" + DNA + "','" + strconv.Itoa(Percentage) + "','" + strconv.Itoa(stat) + "')")
				if err != nil {
					log.Printf("here")
					outputisi := HasilPrediksi{
						NamaPasien:       Nama,
						PenyakitPrediksi: Penyakit,
						TanggalPrediksi:  Tanggal,
						TingkatKemiripan: Percentage,
						Status:           stat,
					}

					output = append(output, outputisi)

					marshal, err := json.Marshal(output)
					if err != nil {
						fmt.Println(err)
					}

					res.Write(marshal)
					defer db.Close()
					return
				}

				defer db_result.Close()

				outputisi := HasilPrediksi{
					NamaPasien:       Nama,
					PenyakitPrediksi: Penyakit,
					TanggalPrediksi:  Tanggal,
					TingkatKemiripan: Percentage,
					Status:           stat,
				}

				output = append(output, outputisi)

				marshal, err := json.Marshal(output)
				if err != nil {
					fmt.Println(err)
				}

				res.Write(marshal)
			} else {
				outputisi := HasilPrediksi{
					NamaPasien:       Nama,
					PenyakitPrediksi: Penyakit,
					TanggalPrediksi:  Tanggal,
					TingkatKemiripan: 0,
					Status:           -2,
				}

				output = append(output, outputisi)

				marshal, err := json.Marshal(output)
				if err != nil {
					fmt.Println(err)
				}

				res.Write(marshal)
			}

		} else {

			outputisi := HasilPrediksi{
				NamaPasien:       Nama,
				PenyakitPrediksi: Penyakit,
				TanggalPrediksi:  Tanggal,
				TingkatKemiripan: 0,
				Status:           -1,
			}

			output = append(output, outputisi)

			marshal, err := json.Marshal(output)
			if err != nil {
				fmt.Println(err)
			}

			res.Write(marshal)
		}

		defer db.Close()
	}
}

func submitDisease(res http.ResponseWriter, req *http.Request) {
	fmt.Println("HERE1")
	setupResponse(&res, req)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatalln(err)
	}

	//Convert the body to type string
	string_body := string(body)
	fmt.Println("HERE")
	if strings.Split(string_body, ":")[0] != "" {
		// remove curly bracket
		string_body = strings.Replace(string_body, "{", "", -1)
		string_body = strings.Replace(string_body, "}", "", -1)

		//remove quote mark
		string_body = strings.Replace(string_body, `"`, "", -1)
		arr := strings.Split(string_body, ",")

		namaPenyakit := strings.Split(arr[0], ":")[1]
		DNA := strings.Split(arr[1], ":")[1]
		fmt.Println(namaPenyakit)
		fmt.Println(DNA)

		if sm.Regex(DNA) {

			db := openDatabase()
			// query := ("INSERT INTO jenispenyakit VALUES(" + "'" + NamaPenyakit + "', '" + DNA + "')")
			query := ("INSERT INTO jenispenyakit VALUES('" + namaPenyakit + "', '" + DNA + "')")
			result, err := db.Query(query)
			if err != nil {
				// fmt.Fprintf(res, "failed duplicate")
				result1, _ := db.Query("SELECT DNA FROM jenispenyakit WHERE NamaPenyakit = '" + namaPenyakit + "'")
				result2, _ := db.Query("SELECT NamaPenyakit FROM jenispenyakit WHERE DNA = '" + DNA + "'")
				empty1, empty2 := true, true

				for result1.Next() {
					empty1 = false
					break
				}
				for result2.Next() {
					empty2 = false
					break
				}
				if empty1 && empty2 {
					fmt.Fprintf(res, "Unexpected Error")
				} else if !empty1 {
					fmt.Fprintf(res, "Nama peyakit sudah ada dalam database")
				} else if !empty2 {
					fmt.Fprintf(res, "untai DNA peyakit sudah ada dalam database")
				}

				defer db.Close()
				return

			}
			defer result.Close()
			defer db.Close()

			fmt.Fprintf(res, "Penyakit berhasil ditambahkan ke dalam database")
		} else {
			fmt.Fprintf(res, "Untai DNA tidak valid")
		}
	}

	// NamaPenyakit := data.NamaPenyakit
	// DNA := data.DNA
	// fmt.Print(NamaPenyakit)
	// fmt.Print(DNA)

}

func main() {

	fmt.Println(sm.BoyerMoore("a pattern matching algorithm", "rithm"))
	fmt.Println(sm.KMP("rithm", "a pattern matching algorithm"))

	fmt.Println(sm.BoyerMoore("abacaabadcabacabaabb", "abacab"))
	fmt.Println(sm.KMP("abacab", "abacaabadcabacabaabb"))

	fmt.Println(sm.RegexSearch("abcd", "abcdefghi"))

	// Server
	http.HandleFunc("/get-detailprediction", getDetailPrediction)
	http.HandleFunc("/get-diseaseprediction", getDiseasePrediction)
	http.HandleFunc("/submitDisease", submitDisease)

	fmt.Println("Starting server at port " + getEnv("BACKEND_PORT"))
	if err := http.ListenAndServe(":"+ getEnv("BACKEND_PORT"), nil); err != nil {
		log.Fatal(err)
	}
}
