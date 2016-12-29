package main

import(
	"fmt"
)

func ERT(bc *BayesClassifier, kc *KNNClassifier, d RawData, k int) {
	n := len(d.species)
	count := 1
	BCcount := 0
	KCcount := 0
	for _, spe := range d.species {
		if spe.Class == bc.BayesPredict(spe.Words) {
			BCcount++
		} 
		if spe.Class == kc.KNNPredict(spe.Words, k) {
			KCcount++
		}
		fmt.Println("# tested:", count, 
			"   NBC successful prediction #", BCcount, 
			"   KNN successful prediction #", KCcount)
		count++
	}
	BCrate := float64(BCcount)/float64(n)
	KCrate := float64(KCcount)/float64(n)
	fmt.Println("Based on test data, error rata of naïve Bayes "+
		"classifier is", 1 - BCrate, ";   error rate of kNN classifier" +
		" is", 1 - KCrate)
}