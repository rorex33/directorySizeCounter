package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rorex33/dirsizecalc"
)

// Необходимо для сортировки
type BySizeASC []dirsizecalc.NameSize

func (a BySizeASC) Len() int           { return len(a) }
func (a BySizeASC) Less(i, j int) bool { return a[i].Size < a[j].Size }
func (a BySizeASC) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type BySizeDESC []dirsizecalc.NameSize

func (a BySizeDESC) Len() int           { return len(a) }
func (a BySizeDESC) Less(i, j int) bool { return a[i].Size > a[j].Size }
func (a BySizeDESC) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

//
//

// Приведение вещественного числа к виду с заданным количеством чисел после запятой.
func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

// Проверка верности входных параметров.
func validation(rootPath string, limit float64, sortType string) error {
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		err := errors.New("validation fail: wrong root path")
		return err
	}

	if limit < 0 {
		err := errors.New("validation fail: wrong limit")
		return err
	}

	if sortType != "asc" && sortType != "desc" {
		err := errors.New("validation fail: wrong sort")
		return err
	}

	return nil

}

// Сортировка и вывод среза в терминал и выходной файл.
func output(outPutArray []dirsizecalc.NameSize, rootPath string, limit float64, sortType string) error {

	if sortType == "asc" {
		sort.Sort(BySizeASC(outPutArray))
	} else {
		sort.Sort(BySizeDESC(outPutArray))
	}

	//Создание выходного файла
	file, err := os.Create("/home/ivan/Desktop/githubProjects/directorySizeCounter/output.txt")
	if err != nil {
		return err
	}
	defer file.Close()
	// x, _ := json.Marshal(outPutArray)
	// fmt.Println(string(x))
	//Вывод данных в выходной файл (если размер больше лимита) и в терминал
	for _, nameSizeValue := range outPutArray {
		toWriteValue := fmt.Sprintf("%s/%s	%s", rootPath, nameSizeValue.Name, fmt.Sprintln(nameSizeValue.Size))
		if nameSizeValue.Size > limit {
			file.WriteString(toWriteValue)
		}
		fmt.Println(toWriteValue)
	}
	return nil
}

// Хендлер-функция для нашего запроса. Парсит параметры, запускает их валидацию, вычисления размеров директорий и вывод результата.
func startCalculation(w http.ResponseWriter, r *http.Request) {
	//Парсинг параметров
	queries := r.URL.Query()
	ROOT := queries["ROOT"][0]
	limit, _ := strconv.ParseFloat(queries["limit"][0], 32)
	sortType := strings.ToLower(queries["sort"][0])

	//Проверка верности указанных параметров
	err := validation(ROOT, limit, sortType)
	if err != nil {
		fmt.Println(err)
	} else {

		//Создаём срез, в котором будут храниться имена и размеры всех папок, находящихся в указанной директории
		nameSizeArray, err := dirsizecalc.ArrayCreation(ROOT)
		if err != nil {
		}

		//Выводим результат
		err = output(nameSizeArray, ROOT, limit, sortType)
		if err != nil {
			fmt.Println("Ошибка при создании выходного файла:", err)
		}
	}
	w.WriteHeader(http.StatusOK)

}

func main() {
	//Создаём роутер и добавляем его параметры
	r := mux.NewRouter()
	r.Path("/dirsize").
		Queries("ROOT", "{ROOT}").
		Queries("limit", "{limit}").
		Queries("sort", "{sort}").
		HandlerFunc(startCalculation)
	//Регистрируем хендлер
	http.Handle("/", r)

	//Запускаем сервер
	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}
