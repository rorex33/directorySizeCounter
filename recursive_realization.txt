package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"time"
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

func dirSizeCalculation(path string) float64 {
	//Общий размер всех файлов и папок в данной директории
	var dirsize float64 = 0

	//Читаем директорию
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println("Ошибка при чтении папки:", err)
		return 0
	}

	//Проходим по всем файлам директории, если это папка - вызываем эту же фукнцию от данной папки
	//Иначе просто прибавляем размер файла к dirsize
	for _, file := range files {
		if file.IsDir() {
			dirsize += dirSizeCalculation(fmt.Sprintf("%s/%s", path, file.Name()))
		}
		dirsize += float64(file.Size())

	}

	//Возвращаем размер папки
	return dirsize
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
		fmt.Println("Ошибка при чтении файлов ROOT дериктории:", err)
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

func main() {
	timeCountingStart := time.Now()

	//Парсинг флагов
	ROOT := flag.String("ROOT", "", "Directory path")
	limit := flag.Float64("limit", -1, "The limit")
	sortType := flag.String("sort", "asc", "Type of sort")
	flag.Parse()
	if *ROOT == "" {
		fmt.Println("Отсутствует путь к дериктории")
		os.Exit(1)
	}
	if *limit < 0 {
		fmt.Println("Нежелательное значение лимита")
		os.Exit(1)
	}
	if *sortType != "asc" && *sortType != "desc" {
		fmt.Println("Неверный тип сортировки (только asc или desc )")
		os.Exit(1)
	}

	//Создаём срез структур nameSize, в котором будут храниться имена и размеры папок указанной директории
	nameSizeArray, err := arrayCreation(*ROOT)
	if err != nil {
		os.Exit(1)
	}

	//Сортировка полученного массива
	if *sortType == "asc" {
		sort.Sort(BySizeASC(nameSizeArray))
	} else {
		sort.Sort(BySizeDESC(nameSizeArray))
	}

	//Вывод
	err = output(nameSizeArray, *limit, *ROOT)
	if err != nil {
		os.Exit(1)
	}

	timeCountingStop := time.Since(timeCountingStart)
	fmt.Println(timeCountingStop)
}
