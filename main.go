package main 

import(
	"fmt"
	"os"
	"log"
	"io"
	"strings"
	"bufio"
	"strconv"
)

var count int = 0

//var dataSetName string = "sortedData.txt"

type Species struct{
	Id			string
	Taxonomy	string
	Sequence	string
	Name 		string
	Class 		Class 
	Words 		[]string
}

type RawData struct {
	classes 	[]Class
	species		[]*Species
	classMap	map[Class]int
}

type Class string

// Generate a smaller data set from full-sized SILVA_SSU.fasta.
// A side effect of this function is to return a slice of struct Species
// of all species in our new data set.
func GetNewDataSetFromFASTA(primaryFileName, dataSetName string) *[]Species {
	file, err := os.Open(primaryFileName)
	if err != nil {
		log.Fatal("Error: There is a problem when opening SILVA fasta file!")
	}
	defer file.Close()
	species := ReadFASTAFile(file, dataSetName)
	for _, s := range *species {
		s.Words = GenerateWords(s.Sequence)
	}
	return species
}

// This is a subroutine of GetNewDataSetFromFASTA(). It parses .fasta file,
// and stores new data set in a .txt file.
func ReadFASTAFile(file io.Reader, dataSetName string) *[]Species {
	scanner := bufio.NewScanner(file)

	outfile, err1 := os.Create(dataSetName)
	if err1 != nil {
		log.Fatal("Error: there is a problem when creating sorted data file!")
	}
	defer outfile.Close()

	species := make([]Species, 0)
	isFirst := true
	s := Species{"", "", "", "", "", make([]string, 0)}
	for scanner.Scan() {
		temp_string := scanner.Text()
		if strings.HasPrefix(temp_string, ">") {
			if isFirst {
				isFirst = false
			} else {
				if s.SortHelper() {
					s.WriteToFile(outfile)
					species = append(species, s)
				}
				s = Species{"", "", "", "", "", make([]string, 0)}
			}
			s.IdHelper(temp_string)
		} else {
			temp_sequence := []string{s.Sequence, temp_string}
			s.Sequence = strings.Join(temp_sequence, "")
		}
	}
	if s.SortHelper() {
		s.WriteToFile(outfile)
		species = append(species, s)
	}
	if scanner.Err() != nil {
		log.Fatal("Sorry: theree was some kind of error during the file"+ 
				"reading!")
	}
	return &species
}

// This is a subroutine of ReadFASTAFile(), and it helps to parse the identity
// line of the fasta format, and then stores it in struct Species.
func (s *Species) IdHelper(temp_string string) {
	var parts []string = strings.Split(temp_string, ";")
	length := len(parts)
	var part_1 []string = strings.Split(parts[0], " ")
	if len(part_1) != 2 {
		log.Fatal("Error: there was an error when parsing identity line!")
	}
	s.Id = part_1[0]
	temp_tax := append([]string{part_1[1]}, parts[1:]...)
	s.Taxonomy = strings.Join(temp_tax, ";")
	s.Class = Class(parts[length-2])
	s.Name = parts[length-1]
} 

// This is a subroutine of ReadFASTAFile(). It helps to select sequence data 
// we need. We can change the selecting standard here.
func (s *Species) SortHelper() bool {
	var parts []string = strings.Split(s.Taxonomy, ";")
	length := len(parts)
	if parts[0] == "Bacteria" && parts[length-2] != "uncultured" {
		count++
		if count % 100 == 0 {
			return true
		}
	}
	return false
}

// This function writes struct Species into a file.
func (s *Species) WriteToFile(outfile io.Writer) {
	fmt.Fprintln(outfile, s.Id, s.Taxonomy)
	fmt.Fprintln(outfile, s.Sequence)
	//fmt.Println("writing")
}

// This function generates 8-mers based on the sequence data of a species.
func GenerateWords(sequence string) []string {
	temp_words := make(map[string]int)
	n := len(sequence)
	for i := 0; i < n-7; i++ {
		temp := sequence[i:i+8]
		temp_words[temp]++
	}
	words := make([]string, 0)
	for word, _ := range temp_words {
		words = append(words, word)
	}
	return words
}

//
func LoadRawData(dataSetName string) *RawData {
	file, err := os.Open(dataSetName)
	if err != nil {
		log.Fatal("Error: There was an error when opening", dataSetName, "!")
	}
	scanner := bufio.NewScanner(file)
	species := make([]Species, 0)
	isFirst := true
	s := Species{"", "", "", "", "", make([]string, 0)}
	for scanner.Scan() {
		temp_string := scanner.Text()
		if strings.HasPrefix(temp_string, ">") {
			if isFirst {
				isFirst = false
			} else {
				species = append(species, s)
				s = Species{"", "", "", "", "", make([]string, 0)}
			}
			s.IdHelper(temp_string)
		} else {
			s.Sequence = temp_string
		}
	}
	species = append(species, s)
	if scanner.Err() != nil {
		log.Fatal("Sorry: theree was some kind of error during the file"+ 
				"reading!")
	}
	for i:= 0; i < len(species); i++ {
		species[i].Words = GenerateWords(species[i].Sequence)
	}
	return NewRawData(&species)
}

func NewRawData(species *[]Species) *RawData {
	d := &RawData{
		make([]Class, 0),
		make([]*Species, 0),
		make(map[Class]int),
	}
	for i := 0; i < len(*species); i++ {
		d.species = append(d.species, &(*species)[i])
		d.classMap[(*species)[i].Class]++
	}

	for class, _ := range d.classMap {
		d.classes = append(d.classes, class)
	}
	return d
}


func main() {
	if len(os.Args) < 3 {
		log.Fatal("Error: there were not enough parameters!")
	}
	if os.Args[1] == "ParseFile" {
		if len(os.Args) != 4 {
			log.Fatal("Error: wrong number of parameters for parsing file!")
		}
		primaryFileName := os.Args[2]
		newDataSetName := os.Args[3]
		GetNewDataSetFromFASTA(primaryFileName, newDataSetName)
	} else if os.Args[1] == "NBC" {
		if len(os.Args) != 4 {
			log.Fatal("Error: wrong number of parameters for running naïve"+
			 " Bayes classifier!")
		}
		if os.Args[2] == "learn" {
			dataSetName := os.Args[3]
			d := LoadRawData(dataSetName)
			bc := BayesLearnData(*d)
			fmt.Println("Number of classes naive Bayes classifier learned:", 
						len(bc.Classes))
			fmt.Println("Number of species naive Bayes classifier learned:", 
						bc.learned)
			bc.BCWriteToFile()
			//fmt.Println("Write naïve Bayes classifier data into file "+
						//"successfully!")
		} else if os.Args[2] == "predict" {
			s := os.Args[3]
			bc := LoadBCFromFile()
			words := GenerateWords(s)
			class := bc.BayesPredict(words)
			fmt.Println("Naïve Bayes classifier prediction:", class)
		} else {
			log.Fatal("Error: wrong command for naïve Bayes classifier!")
		}
	} else if os.Args[1] == "KNN" {
		if os.Args[2] == "crossvalidation" {
			if len(os.Args) != 4 {
				log.Fatal("Error: wrong number of parameters for running "+
						"kNN classifier!")
			}
			dataSetName := os.Args[3]
			d := LoadRawData(dataSetName)
			k := CrossValidation(*d)
			fmt.Println("Optimal k based on current data set:", k)
		} else {
			if len(os.Args) != 5 {
				log.Fatal("Error: wrong number of parameters for running "+
						"kNN classifier!")
			}
			dataSetName := os.Args[2]
			d := LoadRawData(dataSetName)
			kc := KNNLearnData(*d)
			fmt.Println("Number of classes kNN Classifier learned:", 
						len(kc.Classes))
			fmt.Println("Number of species kNN Classifier learned:", 
						kc.learned)
			s := os.Args[4]
			words := GenerateWords(s)
			k, err := strconv.Atoi(os.Args[3])
			if err != nil {
				log.Fatal("Error: wrong k for running kNN classifier"+
				 		" prediction!")
			}
			class := kc.KNNPredict(words, k)
			fmt.Println("kNN classifier prediction:", class)
		} 
	} else if os.Args[1] == "ERT" {
		if len(os.Args) != 5 {
			log.Fatal("Error: wrong number of parameters for running" + 
				"error rate test for both classifier!")
		}
		dataSetName := os.Args[2]
		trainDataSet := os.Args[3]
		k, err := strconv.Atoi(os.Args[4])
		if err != nil {
			log.Fatal("Error: wrong k for running kNN classifier"+
			 		" prediction!")
		}
		d := LoadRawData(dataSetName)
		dt := LoadRawData(trainDataSet)
		bc := LoadBCFromFile()
		kc := KNNLearnData(*dt)
		ERT(bc, kc, *d, k)
	} else if os.Args[1] == "NBKNN" {
		if len(os.Args) != 5 {
			log.Fatal("Error: wrong number of parameters for running" +
				"both classifiers!")
		}
		s := os.Args[3]
		trainDataSet := os.Args[4]
		dt := LoadRawData(trainDataSet)
		words := GenerateWords(s)
		k, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal("Error: wrong k for running kNN classifier"+
			 		" prediction!")
		}
		bc := LoadBCFromFile()
		class1 := bc.BayesPredict(words)
		fmt.Println("Naïve Bayes classifier prediction:", class1)
		kc := KNNLearnData(*dt)
		class2 := kc.KNNPredict(words, k)
		fmt.Println("kNN classifier prediction:", class2)
	} else {
		log.Fatal("Wrong command!")
	}

}



