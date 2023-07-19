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
	"strings"

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

func dirSizeCalculation(path string, c chan<- float64) {
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
	c <- float64(size)
}

func output(outPutAttay []nameSize, limit float64, rootPath string) error {
	//Создание выходного файла
	file, err := os.Create("output.txt")
	if err != nil {
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
		c := make(chan float64)
		defer close(c)
		go dirSizeCalculation(fmt.Sprintf("%s/%s", rootPath, dir.Name()), c)
		dirSize := <-c
		dirSizeMb := dirSize / (1024 * 1024)
		//Создаём переменную типа nameSize и добавления в срез nameSizeArray (размер папки переводится в мегабайты!)
		nameSizeValue := nameSize{dir.Name(), dirSizeMb}
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
	queries := r.URL.Query()
	ROOT := queries["ROOT"][0]
	limit, _ := strconv.ParseFloat(queries["limit"][0], 32)
	sortType := strings.ToLower(queries["sort"][0])

	//Проверка, что лимит и тип сортировки указаны верно
	err := validation(ROOT, limit, sortType)
	if err != nil {
		fmt.Println(err)
	} else {

		//Создаём срез структур nameSize, в котором будут храниться имена и размеры папок указанной директории
		nameSizeArray, err := arrayCreation(ROOT)
		if err != nil {
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
