package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
)

type nameSize struct {
	name string
	size int64
}

type BySizeASC []nameSize

func (a BySizeASC) Len() int           { return len(a) }
func (a BySizeASC) Less(i, j int) bool { return a[i].size < a[j].size }
func (a BySizeASC) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type BySizeDESC []nameSize

func (a BySizeDESC) Len() int           { return len(a) }
func (a BySizeDESC) Less(i, j int) bool { return a[i].size > a[j].size }
func (a BySizeDESC) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func dirSizeCalculation(path string) int64 {
	var dirsize int64 = 0
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println("Ошибка при чтении папки:", err)
		return 0
	}
	for _, file := range files {
		if file.IsDir() {
			dirsize += dirSizeCalculation(fmt.Sprintf("%s/%s", path, file.Name()))
		}
		dirsize += file.Size()

	}
	return dirsize
}

func main() {
	//Парсинг флагов
	ROOT := flag.String("ROOT", "", "Directory path")
	limit := flag.Int64("limit", -1, "The limit")
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

	//Чтение указаной дериктории, поиск папок и вычисление их размера, занесении их в массив структур nameSize(название, размер)
	dirs, err := ioutil.ReadDir(*ROOT)
	if err != nil {
		fmt.Println("Ошибка при чтении файлов ROOT дериктории:", err)
		return
	}
	var nameSizeArray []nameSize
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		dirSize := dirSizeCalculation(fmt.Sprintf("%s/%s", *ROOT, dir.Name()))
		nameSizeValue := nameSize{dir.Name(), dirSize}
		nameSizeArray = append(nameSizeArray, nameSizeValue)

		err = os.Chdir("..")
		if err != nil {
			fmt.Println("..")
			return
		}
	}

	//Сортировка полученного массива
	//Найти более красивый способ
	if *sortType == "asc" {
		sort.Sort(BySizeASC(nameSizeArray))
	} else {
		sort.Sort(BySizeDESC(nameSizeArray))
	}

	//Открытие выходного файла
	file, err := os.Create("/home/ivan/Desktop/test/output.txt")
	if err != nil {
		fmt.Println("Ошибка при создании выходного файла:", err)
		os.Exit(1)
	}
	defer file.Close()

	//Вывод данных в выходной файл и в терминал согласно условию задачи
	for _, nameSizeValue := range nameSizeArray {
		if nameSizeValue.size > *limit {
			file.WriteString(fmt.Sprintf("%s/%s	%s", *ROOT, nameSizeValue.name, fmt.Sprintln(nameSizeValue.size)))
		}
		fmt.Println(fmt.Sprintf("%s/%s	%s", *ROOT, nameSizeValue.name, fmt.Sprint(nameSizeValue.size)))
	}

}
