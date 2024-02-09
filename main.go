package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rorex33/dirsizecalc"
)

// Класс ASC сортировки
type BySizeASC []dirsizecalc.NameSize

func (a BySizeASC) Len() int           { return len(a) }
func (a BySizeASC) Less(i, j int) bool { return a[i].Size < a[j].Size }
func (a BySizeASC) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// Класс DESC сортировки
type BySizeDESC []dirsizecalc.NameSize

func (a BySizeDESC) Len() int           { return len(a) }
func (a BySizeDESC) Less(i, j int) bool { return a[i].Size > a[j].Size }
func (a BySizeDESC) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

//
//

// Возвращает текущий IP компьютера в сети (например: 192.168.1.0)
func getIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := strings.Split(conn.LocalAddr().String(), ":")

	return localAddr[0]
}

// Разбивает строку конфига на два строчные переменные: атрибут и его значение
func configStringParse(val string) (string, string) {
	i := 0
	for val[i] != ' ' {
		i++
	}
	name := val[0:i]
	for val[i] == ' ' || val[i] == '=' {
		i++
	}
	anch := i // "Якорь" для выделения части строки
	value := val[anch:]
	return name, value
}

// Функция считывает файл "server.config" и возвращает мапу из его атрибутов и их значений
func configRead() (map[string]string, error) {
	//Хранилище данных конфига в формате "атрибут":"значение"
	config := make(map[string]string)

	//Считываем конфиг
	file, err := os.Open("server.config")
	if err != nil {
		return config, err
	}
	defer file.Close()

	//Проходим по файлу, парсим его строки на атрибуты и их значения
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tempName, tempValue := configStringParse(scanner.Text())
		config[tempName] = tempValue
	}
	if err := scanner.Err(); err != nil {
		return config, err
	}

	return config, nil
}

// Приведение вещественного числа к виду с заданным количеством цифр после запятой.
func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

// Проверка верности входных параметров.
func validation(rootPath string, sortType string) error {
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		err := errors.New("Ошибка валидации: неверный путь " + rootPath)
		return err
	}

	if sortType != "asc" && sortType != "desc" {
		err := errors.New("Ошибка валидации: неверный тип сортировки")
		return err
	}

	return nil

}

// Сортировка и вывод в формате JSON
func output(dirsOutPutArray []dirsizecalc.NameSize, filesOutPutArray []dirsizecalc.NameSize, rootPath string, sortType string) string {
	//Сортировка папок
	if sortType == "asc" {
		sort.Sort(BySizeASC(dirsOutPutArray))
	} else {
		sort.Sort(BySizeDESC(dirsOutPutArray))
	}

	//Сортировка файлов
	if sortType == "asc" {
		sort.Sort(BySizeASC(filesOutPutArray))
	} else {
		sort.Sort(BySizeDESC(filesOutPutArray))
	}

	//Создание массива мап, хранящих параметры найденных папок
	mapOfDirs := make([]map[string]string, len(dirsOutPutArray))
	for i := range dirsOutPutArray {

		//Временная мапа для добавления в итоговый JSON
		term := make(map[string]string)

		//Атрибуты "name" и "size" - берутся из переданного функции массива структур с полями "имя" и "размер"
		term["name"] = dirsOutPutArray[i].Name
		//Размер приводится к виду с двумя цифрами после запятой
		term["size"] = fmt.Sprint(roundFloat(dirsOutPutArray[i].Size, 2))
		//Атрибут путь формируется из переданного функции путя к директории и названия папки
		term["path"] = fmt.Sprintf("%s/%s", rootPath, string(dirsOutPutArray[i].Name))

		//Временная мапа добавляется в массив
		mapOfDirs[i] = term
	}

	//Создание массива мап, хранящих параметры найденных файлов
	mapOfFiles := make([]map[string]string, len(filesOutPutArray))
	for i := range filesOutPutArray {

		//Временная мапа для добавления в итоговый JSON
		term := make(map[string]string)

		//Атрибуты "name" и "size" - берутся из переданного функции массива структур с полями "имя" и "размер"
		term["name"] = filesOutPutArray[i].Name
		//Размер приводится к виду с двумя цифрами после запятой
		term["size"] = fmt.Sprint(roundFloat(filesOutPutArray[i].Size, 2))

		//Временная мапа добавляется в массив
		mapOfFiles[i] = term
	}

	response := make(map[string]interface{})

	//Формирование ответа
	response["status"] = "1"
	response["error"] = "None"
	response["dirs"] = mapOfDirs
	response["files"] = mapOfFiles
	response["pathDirs"] = strings.Split(rootPath, "/")[1:]

	//Перевод ответа в json формат
	responseInJSON, _ := json.Marshal(response)
	//fmt.Printf("%s\n", responseInJSON)
	//fmt.Println()
	return fmt.Sprintf("%s\n", responseInJSON)
}

// Вывод в формате JSON при ошибке
func errOutPut(err error) string {
	response := make(map[string]interface{})

	//Формирование ответа
	response["status"] = "0"
	response["error"] = fmt.Sprint(err)
	response["dirs"] = "Something gone wrong"
	response["pathDirs"] = "Something gone wrong"

	//Перевод ответа в json формат
	responseInJSON, _ := json.Marshal(response)
	fmt.Printf("%s\n", responseInJSON)
	return fmt.Sprintf("%s\n", responseInJSON)
}

// Хендлер-функция для нашего запроса. Парсит параметры, запускает их валидацию, вычисления размеров директорий и вывод результата.
func startCalculation(w http.ResponseWriter, r *http.Request) {
	//Парсинг параметров
	queries := r.URL.Query()
	ROOT := queries["ROOT"][0]
	sortType := strings.ToLower(queries["sort"][0])

	//Ответ на запрос
	var response string = ""

	//Проверка верности указанных параметров
	err := validation(ROOT, sortType)
	if err != nil {
		response = errOutPut(err)
		fmt.Println(err)
	} else {

		//Создаём срез, в котором будут храниться имена и размеры всех папок, находящихся в указанной директории
		//Также создаём срез всех файлов указанной директории
		dirNameSizeArray, filesDirSizeArray, err := dirsizecalc.GetDirectories(ROOT)
		if err != nil {
			response = errOutPut(err)
		}

		//Выводим результат
		response = output(dirNameSizeArray, filesDirSizeArray, ROOT, sortType)
	}

	//Загрузка в хедер ответа параметров, без которых некоторые браузеры не примут ответ
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.WriteHeader(http.StatusOK)

	//Вывод ответа
	w.Write([]byte(response))
}

func main() {
	//Создаём роутер и добавляем его параметры
	r := mux.NewRouter()

	//
	r.HandleFunc("/dirsize", startCalculation).
		Queries("ROOT", "{ROOT}").
		Queries("sort", "{sort}").
		Methods("GET", "OPTIONS")

	//
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("."))))

	//Чтение конфига
	config, err := configRead()
	if err != nil {
		fmt.Println("Ошибка при чтении конфига ", err)
		config = map[string]string{}
	}

	//Создаём сервер
	serverAddress := fmt.Sprintf("%s:%s", getIP(), config["port"])
	server := &http.Server{
		Addr:         serverAddress,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	//Запускаем сервер
	log.Println("Listening on " + serverAddress)
	server.ListenAndServe()
}
