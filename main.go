package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
)

type nameSize struct {
	name string
	size float64
}

// Необходимо для сортировки
type BySizeASC []nameSize

func (a BySizeASC) Len() int           { return len(a) }
func (a BySizeASC) Less(i, j int) bool { return a[i].size < a[j].size }
func (a BySizeASC) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type BySizeDESC []nameSize

func (a BySizeDESC) Len() int           { return len(a) }
func (a BySizeDESC) Less(i, j int) bool { return a[i].size > a[j].size }
func (a BySizeDESC) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func parametersCheck(limit float64, sortType string) error {
	err := errors.New("parametersCheck: wrong limit or sortType")
	if limit < 0 || sortType != "asc" && sortType != "desc" {
		return err
	}
	return nil

}

func dirSizeCalculation(path string) float64 {
	//Открываем канал sizes для передачи в него размеров вложенных дерикторий
	sizes := make(chan int64)

	//Данная функция считает размер всех файлов, которые не являются директориями, и отправляет результат канал sizes
	readSize := func(path string, file os.FileInfo, err error) error {
		if err != nil || file == nil {
			return nil
		}
		if !file.IsDir() {
			sizes <- file.Size()
		}
		return nil
	}

	//Каждая горутина считывает размер открытой для неё папки и отправляет результат в канал sizes
	//После завершения работы всег горутин канал закрывается
	go func() {
		filepath.Walk(path, readSize)
		close(sizes)
	}()

	//Суммируем всё, что находится в нашем канале sizes
	size := int64(0)
	for s := range sizes {
		size += s
	}

	//Возвращаем итоговый размер директории с учётом всех вложенных директорий
	return float64(size)
}

func output(outPutAttay []nameSize, limit float64, rootPath string) error {
	//Создание выходного файла
	file, err := os.Create("/home/ivan/Desktop/test/output.txt")
	if err != nil {
		fmt.Println("Ошибка при создании выходного файла:", err)
		return err
	}
	defer file.Close()

	//Вывод данных в выходной файл и в терминал согласно условию задачи
	for _, nameSizeValue := range outPutAttay {
		if nameSizeValue.size > limit {
			file.WriteString(fmt.Sprintf("%s/%s	%s", rootPath, nameSizeValue.name, fmt.Sprintln(nameSizeValue.size)))
		}
		fmt.Println(fmt.Sprintf("%s/%s	%s mb", rootPath, nameSizeValue.name, fmt.Sprint(roundFloat(nameSizeValue.size, 5))))
	}
	return nil
}

func arrayCreation(rootPath string) ([]nameSize, error) {
	//Срез структур с полями "имя" "размер" (для хранения имени и размера папок)
	var nameSizeArray []nameSize

	//Проходим по всем файлам указанной директории.
	dirs, err := ioutil.ReadDir(rootPath)
	if err != nil {
		fmt.Println("Ошибка при чтении файлов ROOT директории:", err)
		return nameSizeArray, err
	}

	//Если очередной файл явялется папкой, то вычисляем проводим необходимые действия
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		//Вычисляем размер папки
		dirSize := dirSizeCalculation(fmt.Sprintf("%s/%s", rootPath, dir.Name()))

		//Создаём переменную типа nameSize и добавления в срез nameSizeArray (размер папки переводится в мегабайты!)
		nameSizeValue := nameSize{dir.Name(), dirSize / 1024 / 1024}
		nameSizeArray = append(nameSizeArray, nameSizeValue)

		//Обработка возможной ошибки при вовзращении в родительскую директорию
		err = os.Chdir("..")
		if err != nil {
			fmt.Println("..")
			return nameSizeArray, err
		}
	}

	return nameSizeArray, nil
}

func startCalculation(w http.ResponseWriter, r *http.Request) {
	//Парсинг параметров
	vars := mux.Vars(r)
	//[1:len(...)] - очистка параметров от лишних символов
	ROOT := vars["ROOT"][1 : len(vars["ROOT"])-1]
	limit, _ := strconv.ParseFloat(vars["limit"][1:len(vars["limit"])-1], 32)
	sortType := vars["sort"][1 : len(vars["sort"])-1]

	//Проверка, что лимит и тип сортировки указаны верно
	err1 := parametersCheck(limit, sortType)
	if err1 != nil {
		fmt.Println(err1)
		//os.Exit(1)
	} else {
		//Создаём срез структур nameSize, в котором будут храниться имена и размеры папок указанной директории
		nameSizeArray, err := arrayCreation(ROOT)
		if err != nil {
			//os.Exit(1)
		}

		//Сортировка полученного массива
		if sortType == "asc" {
			sort.Sort(BySizeASC(nameSizeArray))
		} else {
			sort.Sort(BySizeDESC(nameSizeArray))
		}

		//Вывод
		err = output(nameSizeArray, limit, ROOT)
		if err != nil {
			//os.Exit(1)
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
