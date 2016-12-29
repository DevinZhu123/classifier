package main 

import (
	"os"
	"fmt"
	"io"
	"encoding/gob"
	"math"
)

type BayesClassifier struct{
	Classes 	[]Class
	data 		map[Class]*BayesClassData
	globalData	map[string]int
	learned 	int
	seen 		int
}

type FormatBayesClassifier struct{
	Classes 	[]Class
	Data 		map[Class]*BayesClassData
	GlobalData	map[string]int
	Learned 	int
	Seen 		int
}

type BayesClassData struct{
	Freq		map[string]int
	Sum			int
}

//BayesLearnData() learns data from raw data, stores the classifier in file, 
//and return the pointer to the classifier.
func BayesLearnData(d RawData) (*BayesClassifier) {
	bc := new(BayesClassifier)
	bc.Classes = make([]Class, len(d.classes))
	bc.globalData = make(map[string]int)
	bc.learned = 0
	bc.seen = 0
	var classes []Class = d.classes
	bc.Classes = classes
	bc.GenerateData(d)
	bc.GenerateGlobalData(d)
	//bc.BCWriteToFile()

	return bc
}


func (bc *BayesClassifier) GenerateGlobalData(d RawData) {
	n := len(d.species)
	for i := 0; i < n; i++ {
		length := len(d.species[i].Words)
		for j := 0; j < length; j++ {
			bc.globalData[d.species[i].Words[j]]++
		}
	}
}


//GenerateData() generates the map of class to class data.
func (bc *BayesClassifier) GenerateData(d RawData) {
	length := len(d.classes)
	bc.data = make(map[Class]*BayesClassData, length)
	for _, class := range d.classes {
		bc.data[class] = newBayesClassData()
	}

	for i := 0; i < len(d.species); i++  {
		bc.UpdateData(d.species[i])
	}
}

func newBayesClassData() *BayesClassData {
	return &BayesClassData{
		Freq:	make(map[string]int),
		Sum:	0,
	}
}

// UpdateData() updates bayes class data based on a single species data.
func (bc *BayesClassifier) UpdateData(species *Species) {
	//var temp *BayesClassData = bc.data[species.class]

	for i := 0; i < len(species.Words); i++ {
		bc.data[species.Class].Freq[species.Words[i]]++
	}
	bc.data[species.Class].Sum++
	bc.learned++

}

//Store a Bayes classifier to a .gob file.
func (bc *BayesClassifier) BCWriteToFile() {
	file, err := os.Create("BayesClassifier.gob")
	if err != nil {
		fmt.Println("Error1: There is a problem when writing Bayes " +
					"classifier to a file!")
		os.Exit(1)
	}
	defer file.Close()
	bc.WriteTo(file)
}

func (bc *BayesClassifier) WriteTo(file io.Writer) {
	enc := gob.NewEncoder(file)
	err := enc.Encode(&FormatBayesClassifier{bc.Classes, bc.data, 
		bc.globalData, bc.learned, bc.seen})
	if err != nil {
		fmt.Println("Error2: There is a problem when writing Bayes " + 
					"classifier to a file!")
		os.Exit(1)
	}
	fmt.Println("Write file successfully!")
}

//Predict the class of sequence, based on existing Bayes classifier.
func (bc *BayesClassifier) BayesPredict(words []string) Class {
	//bc := LoadBCFromFile()
	length := len(bc.Classes)
	scores := make(map[Class]float64, length)
	for _, class := range bc.Classes {
		scores[class] = bc.getWordsProb(class, words)
	}
	predictClass := maxScore(scores)

	bc.seen++
	return predictClass
}

// Load existing Bayes classifier from file.
func LoadBCFromFile() *BayesClassifier {
	file, err := os.Open("BayesClassifier.gob")
	if err != nil {
		fmt.Println("Error: There is a problem when loading Bayes classifier "+
					"from file!")
		os.Exit(1)
	}
	dec := gob.NewDecoder(file)
	bc := new(FormatBayesClassifier)
	err = dec.Decode(bc)
	if err != nil {
		fmt.Println("Error: There is a problem when loading Bayes classifier "+
					"from file!")
		os.Exit(1)
	}
	if bc.Learned == 0 {
		fmt.Println("Error: Bayes classifier not initialized!")
		os.Exit(1)
	}
	return &BayesClassifier{bc.Classes, bc.Data, bc.GlobalData, 
		bc.Learned, bc.Seen}
}

// Get score of a sequence, based on a specific class.
func (bc *BayesClassifier) getWordsProb(class Class, words []string) float64 {

	score := 0.0
	for _, word := range words {
		score += math.Log(bc.wordProb(class, word))
	}
	return score
}

// Return the probability of a word existing in a sequence, based on a 
//sepcific class.
func (bc *BayesClassifier) wordProb(class Class, word string) float64 {
	tempData := bc.data[class]
	defaultProb := 1e-20
	priorProb := bc.wordPriorProb(word) / (float64(tempData.Sum) + 1.0)
	wordFreq, err := tempData.Freq[word]
	if !err {
		return defaultProb
	}

	return (float64(wordFreq) / float64(tempData.Sum) + 1.0) + priorProb
}

func (bc *BayesClassifier) wordPriorProb(word string) float64{
	sumOfWord := bc.globalData[word]
	return (float64(sumOfWord))/(float64(bc.learned))
}

// Find the class having the maximum score.
func maxScore(scores map[Class]float64) Class {
	var maxClass Class
	//fmt.Println(scores)
	var maxscore float64
	isInitial := true
	for class, score := range scores {
		if isInitial {
			maxClass = class
			maxscore = score
			isInitial = false
		}
		if maxscore < score {
			maxClass = class
			maxscore = score
		}
	}
	return maxClass
}

//Reset the Bayes classifier file.
func ResetBayesClassifier() {
	file, err := os.Create("BayesClassifier.gob")
	if err != nil {
		fmt.Println("Error1: There is a problem when resetting Bayes " +
					"classifier file!")
		os.Exit(1)
	}
	defer file.Close()
}






