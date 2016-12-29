Dec 2, 2016
Zhefan Zhu

--------
Content
--------

This program implements a text-based naive Bayes classifier and a 
k-nearest-neighbor classifier for bacterial taxonomy. The 16s rRNA sequence is
 parsed into a set of 8-mers as the features of the species.


-------------
Instructions
-------------

1. To parse a .fasta file (e.g. SILVA_128_SSURef_tax_silva.fasta from 
https://www.arb-silva.de/no_cache/download/archive/release_128/Exports/), 
we can use command:
./classifier   TransformFile   OrignalFileName   NewDataSetName

2. To train naive Bayes classifier, we can use command:
./classifier   NBC   learn   TrainDataSetName

3. To predict a sequence with naive Bayes classifier, we can use command:
./classifier   NBC   predict   Sequence

4. To find optimal k for kNN classifier based on a specific training data set,
we can use command:
./classifier   KNN   crossvalidation   TrainDataSetName

5. To predict a sequence with kNN classifier, k could be arbitrary or achieved
from #4 command. We can use command:
./classifier   KNN    TrainDataSetName   k    Sequence

6. To run both classifier together (NBC should have been trained before), we 
can use command:
./classifier   NBKNN    k    Sequence    TrainDataSetName

7. To run error rate test for two classifiers with test data set, we can use 
command;
./classifier   ERT    TestDataSetName     TrainDataSetName    k




PS: 
1. In this package, there is a training data set (SortedData.txt) for training
classifier, and a testing data set (TestData.txt) for the use of error rate 
test. You can use other data set, but the sequence of a species should be 
linked in a single line.

3. The trained kNN classifier can not be stored in a .gob file, because the 
memory usage will increase significantly when writing file. This problem may 
be solved in future.


