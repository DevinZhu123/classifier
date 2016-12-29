package main 

import (
	"log"
	"fmt"
	"os"
	"io"
	"encoding/gob"
	"container/heap"
	"math/rand"
	"sort"
	"time"
)

var kNNDataBase string = "kNNClassifier.gob"

type KNNClassifier struct {
	Classes			[]Class
	data			map[string][]*Species
	learned			int
	seen 			int
}

type SerializedKNNClassifier struct {
	Classes 		[]Class
	Data 			map[string][]*Species
	Learned 		int
	Seen 			int
}



// An Item is something we manage in a priority queue.
type Item struct {
	value    *Species // The value of the item; a pointer to a Species.
	priority int    // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface 
	//methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use 
	//greater than here.
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func KNNLearnData(d RawData) *KNNClassifier {
	numOfClasses := len(d.classes)
	kc := &KNNClassifier{
		make([]Class, numOfClasses),
		make(map[string][]*Species),
		0,
		0,
	}
	copy(kc.Classes, d.classes)
	kc.LearnDataHelper(d)
	//kc.WritekNNToFile()
	//fmt.Println("Number of classes kNN Classifier learned:", len(kc.Classes))
	//fmt.Println("Number of species kNN Classifier learned:", kc.learned)
	return kc
}

func (kc *KNNClassifier) LearnDataHelper(d RawData) {
	numOfSpecies := len(d.species)
	for i := 0; i < numOfSpecies; i++ {
		numOfWords := len(d.species[i].Words) 
		kc.learned++
		for j := 0; j < numOfWords; j++ {
			word := d.species[i].Words[j]
			_, isExist := kc.data[word]
			if !isExist {
				kc.data[word] = make([]*Species, 0)
			}
			kc.data[word] = append(kc.data[word], d.species[i])
		}
	}
}

func (kc *KNNClassifier) WritekNNToFile() {
	file, err := os.Create("kNNClassifier.gob")
	if err != nil {
		log.Fatal("Error: There was a problem when creating kNN" +
			" Classifier file!")
	}
	defer file.Close()
	kc.WriteTo(file)
}

func (kc *KNNClassifier) WriteTo(file io.Writer) {
	enc := gob.NewEncoder(file)
	err := enc.Encode(&SerializedKNNClassifier{kc.Classes, kc.data, 
		kc.learned, kc.seen})
	if err != nil {
		log.Fatal("Error: There was a problem when encoding kNN" +
			" Classifier data!", err)
	}
	fmt.Println("Write kNN classifier data file successfully!")
}


func LoadKCFromFile() *KNNClassifier {
	file, err := os.Open(kNNDataBase)
	if err != nil {
		log.Fatal("Error: There was a problem when loading kNN" +
			" classifier data from local file!")
	}
	dec := gob.NewDecoder(file)
	skc := new(SerializedKNNClassifier)
	err = dec.Decode(skc)
	if err != nil {
		log.Fatal("Error: There was a problem when decoding local" +
			"kNN classifier data!", err)
	}
	return &KNNClassifier{skc.Classes, skc.Data, skc.Learned, skc.Seen}
}

func (kc *KNNClassifier) KNNPredict(words []string, k int) Class {
	speciesFreq := make(map[*Species]int)
	for _, word := range words {
		length := len(kc.data[word])
		temp_data := kc.data[word]
		for i := 0; i < length; i++  {
			temp_species := temp_data[i]
			speciesFreq[temp_species]++
		}
	}


	kMax := FindKMax(speciesFreq, k)
	//for spe, num := range speciesFreq {
	//	fmt.Println(spe.name, num)
	//}
	classMap := make(map[Class]int)
	for i := 0; i < k; i++ {
		classMap[kMax[i].Class]++
	}

	countClass := 0
	var maxClass Class 
	for class, num := range classMap {
		if num > countClass {
			maxClass = class 
		}
	}

	return maxClass
}

func FindKMax(s map[*Species]int, k int) []*Species {
	pq := make(PriorityQueue, 1)
	isFirst := true
	for value, priority := range s {
		if isFirst {
			pq[0] = &Item{
			value:		value,
			priority: 	priority,
			index:		0, 
			}
			heap.Init(&pq)
			isFirst = false
		} else {
			item := &Item{
				value:		value,
				priority:	priority,
			}
			if len(pq) < k {
				heap.Push(&pq, item)
			} else if len(pq) == k {
				oldItem := heap.Pop(&pq).(*Item)
				//fmt.Println(oldItem.value.class)
				if oldItem.priority > item.priority {
					heap.Push(&pq, oldItem)
				} else {
					heap.Push(&pq, item)
				}
			} else {
				log.Fatal("Error: There was an error when finding closest" +
					" k species!")
			}
		}
	}
	result := make([]*Species, k)
	for i := 0; i < k; i++ {
		result[i] = pq[i].value
		//fmt.Println(pq[i].value.name, pq[i].value.class)
	}
	return result
}

func CrossValidation(d RawData) int {
	rand.Seed(time.Now().UTC().UnixNano())
	correctRate := make(map[int]float64)
	for k := 1; k <= 10; k++ {
		temp := make([]float64, 20)
		average := 0.0
		for i := 0; i < 20; i++ {
			fmt.Println("Repeat times:", i+1)
			temp[i] = CrossValidationHelper(d, k)
			average += temp[i]
		}
		average /= 20
		correctRate[k] = average
	}
	cr := -1.0
	optimalK := 1
	for k, rate := range correctRate {
		//fmt.Println("k =", k, "   average correct rate:", rate)
		if rate > cr {
			optimalK = k
			cr = rate
		}
	}
	fmt.Println("Optimal k =", optimalK, "average correct rate:", cr)
	return optimalK
}

func CrossValidationHelper(d RawData, k int) float64 {
	n := len(d.species)
	//for _, spe := range d.species {
	//	fmt.Println("RawData:", spe.name)
	//}
	var numOfTest int = n / 1000
	//var numOfTrain int = n - numOfTest
	testSpecies := make([]*Species, numOfTest)
	rand := GenerateRandomNum(n, numOfTest)
	for i := 0; i < numOfTest; i++ {
		testSpecies[i] = d.species[rand[i]]
	}
	sort.Ints(rand)

	trainSpecies := make([]*Species, n)
	copy(trainSpecies, d.species) 
	for i := numOfTest - 1; i >= 0; i-- {
		if rand[i] != n-1 {
			trainSpecies = append(trainSpecies[:rand[i]], 
				trainSpecies[rand[i]+1:]...)
		} else {
			trainSpecies = trainSpecies[:rand[i]]
		}
	}
	/*
	for _, spe := range testSpecies {
		fmt.Println("test species: ", spe.name)
	}
	for _, spe := range trainSpecies {
		fmt.Println("train species: ", spe.name)
	}
	*/
	temp_trainData := RawData{
		classes:	make([]Class, 0),
		species:	trainSpecies,
		classMap:	make(map[Class]int),
	}
	kc := KNNLearnData(temp_trainData)
	count := 0
	for i := 0; i < numOfTest; i++ {
		//fmt.Println("testspecies:", testSpecies[i].name, testSpecies[i].class, 
		//		"predict result:", kc.KNNPredict(testSpecies[i].words, k))
		if testSpecies[i].Class == kc.KNNPredict(testSpecies[i].Words, k) {
			count++
		}
	}
	fmt.Println("k:", k, "   correct rate:", 
		float64(count) / float64(numOfTest))
	return float64(count) / float64(numOfTest)
}

func GenerateRandomNum(n, m int) []int {
	temp := make(map[int]int)
	for len(temp) < m {
		num := rand.Intn(n)
		temp[num]++
	}
	result := make([]int, m)
	i := 0
	for num, _ := range temp {
		result[i] = num
		i++
	}
	return result
}