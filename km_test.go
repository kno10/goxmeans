package goxmeans

import (
	"bufio"
	"fmt"
	"os"
	"math"
	"testing"
	"github.com/bobhancock/gomatrix/matrix"
	"goxmeans/matutil"
)

var DATAPOINTS = matrix.MakeDenseMatrix([]float64{3.0,2.0,
	-3.0,2.0,
	0.355083,-3.376585,
	1.852435,3.547351,
	-2.078973,2.552013,
	-0.993756,-0.884433,
	2.682252,4.007573,
	-3.087776,2.878713,
	-1.565978,-1.256985,
	2.441611,0.444826,
	10.29,20.6594,
	12.93,23.3988}, 12, 2)

var CENTROIDS = matrix.MakeDenseMatrix([]float64{ 4.5,   11.3,
    6.1,  12.0,
    12.1,   9.6}, 3, 2)

var DATAPOINTS_D = matrix.MakeDenseMatrix( []float64{2,3, 3,2, 3,4, 4,3, 8,7, 9,6, 9,8, 10,7}, 8,2)
var CENTROIDS_D = matrix.MakeDenseMatrix([]float64{6,7}, 1,2)

var DATAPOINTS_D0 = matrix.MakeDenseMatrix( []float64{2,3, 3,2, 3,4, 4,3}, 4,2)
var CENTROID_D0 =  matrix.MakeDenseMatrix([]float64{3,3}, 1,2) 

var DATAOINTS_D1 = matrix.MakeDenseMatrix( []float64{8,7, 9,6, 9,8, 10,7}, 4,2)
var CENTROID_D1 =  matrix.MakeDenseMatrix([]float64{9,7}, 1,2) 

func makeClusterAssessment(datapoints, centroids *matrix.DenseMatrix) *matrix.DenseMatrix {
	r, c := datapoints.GetSize()
	clusterAssessment := matrix.Zeros(r, c)

	done := make(chan int)
	jobs := make(chan PairPointCentroidJob, r)
	results := make(chan PairPointCentroidResult, minimum(1024, r))
	var ed matutil.EuclidDist

	go addPairPointCentroidJobs(jobs, datapoints, centroids, clusterAssessment, ed, results)
		
	for i := 0; i < r; i++ {
		go doPairPointCentroidJobs(done, jobs)
	}
	go awaitPairPointCentroidCompletion(done, results)
	
	clusterChanged := assessClusters(clusterAssessment, results)
	
	if clusterChanged == true || clusterChanged == false {
	}
	//fmt.Printf("clusterchanged=%v\n", clusterChanged)
	return clusterAssessment
}

func TestAtof64Invalid(t *testing.T) {
	s := "xyz"
	if _, err := Atof64(s); err == nil {
		t.Errorf("err == nil with invalid input %s.", s)
	}
}

func TestAtof64Valid(t *testing.T) {
	s := "1234.5678"
	if f64, err := Atof64(s); err != nil {
		t.Errorf("err != nil with input %s. Returned f64=%f,err= %v.", s, f64, err)
	}
}

func TestFileNotExistsLoad(t *testing.T) {
	f := "filedoesnotexist"
	if _, err := Load(f); err == nil {
		t.Errorf("err == nil with file that does not exist.  err=%v.", err)
	}
}

func createtestfile(fname, record string) (int, error) {
	fp, err := os.Create(fname)
	if err != nil {
		return 0, err
	}
	defer fp.Close()

	w := bufio.NewWriter(fp)
	i, err := w.WriteString(record)
	if err != nil {
		return i, err
	}
	w.Flush()

	return i, err
}

// Does the input line contain < 2 elements
func TestInputInvalid(t *testing.T) {
	fname := "inputinvalid"
	_, err := createtestfile(fname, "123\n")
	if err != nil {
		t.Errorf("Could not create test file. err=%v", err)
	}
	defer os.Remove(fname)

	if _, err := Load(fname); err == nil {
		t.Errorf("err: %v", err)
	}
}

func TestValidReturnLoad(t *testing.T) {
	fname := "inputvalid"
	record := fmt.Sprintf("123\t456\n789\t101")
	_, err := createtestfile(fname, record)
	if err != nil {
		t.Errorf("Could not create test file %s err=%v", fname, err)
	}
	defer os.Remove(fname)

	if _, err := Load(fname); err != nil {
		t.Errorf("Load(%s) err=%v", fname, err)
	}
}

/* Test fails
func TestRandCentroids(t *testing.T) {
	rows := 3
	cols := 3
	k := 2
	data := []float64{1, 2.0, 3, -4.945, 5, -6.1, 7, 8, 9}
	mat := matrix.MakeDenseMatrix(data, rows, cols)
	choosers := []CentroidChooser{RandCentroids{}, DataCentroids{}, EllipseCentroids{0.5}}
	for _, cc := range choosers{
		centroids := cc.ChooseCentroids(mat, k)

		r, c := centroids.GetSize()
		if r != k || c != cols {
			t.Errorf("Returned centroid was %dx%d instead of %dx%d", r, c, rows, cols)
		}
	}
}
*/


func TestComputeCentroid(t *testing.T) {
	empty := matrix.Zeros(0, 0)
	_, err := ComputeCentroid(empty)
	if err == nil {
		t.Errorf("Did not raise error on empty matrix")
	}
	twoByTwo := matrix.Ones(2, 2)
	centr, err := ComputeCentroid(twoByTwo)
	if err != nil {
		t.Errorf("Could not compute centroid, err=%v", err)
	}
	expected := matrix.MakeDenseMatrix([]float64{1.0, 1.0}, 1, 2)
	if !matrix.Equals(centr, expected) {
		t.Errorf("Incorrect centroid: was %v, should have been %v", expected, centr)
	}
	twoByTwo.Set(0, 0, 3.0)
	expected.Set(0, 0, 2.0)
	centr, err = ComputeCentroid(twoByTwo)
	if err != nil {
		t.Errorf("Could not compute centroid, err=%v", err)
	}
	if !matrix.Equals(centr, expected) {
		t.Errorf("Incorrect centroid: was %v, should have been %v", expected, centr)
	}
}


func TestKmeansp(t *testing.T) {
	dataPoints, err := Load("./testSetSmall.txt")
	if err != nil {
		t.Errorf("Load returned: %v", err)
		return
	}
	
	var ed matutil.EuclidDist
	var cc RandCentroids
	//centroidsdata := []float64{1.5,1.5,2,2,3,3,0.9,0,9}
	//centroids := matrix.MakeDenseMatrix(centroidsdata, 4,2)

	centroidMeans, centroidSqDist, err := Kmeansp(dataPoints, 4, cc, ed)
	if err != nil {
		t.Errorf("Kmeans returned: %v", err)
		return
	}

	if 	a, b := centroidMeans.GetSize(); a == 0 || b == 0 {
		t.Errorf("Kmeans centroidMeans is of size %d, %d.", a,b)
	}

	if c, d := centroidSqDist.GetSize(); c == 0 || d == 0 {
		t.Errorf("Kmeans centroidSqDist is of size %d, %d.", c,d)
	}
}
   
func TestAddPairPointToCentroidJob(t *testing.T) {
	r := 4
	c := 2
	jobs := make(chan PairPointCentroidJob, r)
	results := make(chan PairPointCentroidResult, minimum(1024, r))
	dataPoints := matrix.Zeros(r, c)
	centroidSqDist := matrix.Zeros(r, c)
	centroids := matrix.Zeros(r, c)

	var ed matutil.EuclidDist
	
	go addPairPointCentroidJobs(jobs, dataPoints, centroids, centroidSqDist,ed ,results)
	i := 0
	for ; i < r; i++ {
        <-jobs 
		//fmt.Printf("Drained %d\n", i)
    }

	if i  != r {
		t.Errorf("addPairPointToCentroidJobs number of jobs=%d.  Should be %d", i, r)
	}
}
	
func TestDoPairPointCentroidJobs(t *testing.T) {
	r := 4
	c := 2
	dataPoints := matrix.Zeros(r, c)
	centroidSqDist := matrix.Zeros(r, c)
	centroids := matrix.Zeros(r, c)

	done := make(chan int)
	jobs := make(chan PairPointCentroidJob, r)
	results := make(chan PairPointCentroidResult, minimum(1024, r))

	var md matutil.ManhattanDist

	go addPairPointCentroidJobs(jobs, dataPoints, centroids, centroidSqDist, md, results)

	for i := 0; i < r; i++ {
		go doPairPointCentroidJobs(done, jobs)
	}

	j := 0
	for ; j < r; j++ {
        <- done
    }

	if j  != r {
		t.Errorf("doPairPointToCentroidJobs jobs processed=%d.  Should be %d", j, r)
	}
}

func TestAssessClusters(t *testing.T) {
	r, c := DATAPOINTS.GetSize()
	clusterAssessment := matrix.Zeros(r, c)

	done := make(chan int)
	jobs := make(chan PairPointCentroidJob, r)
	results := make(chan PairPointCentroidResult, minimum(1024, r))

	var md matutil.ManhattanDist
	go addPairPointCentroidJobs(jobs, DATAPOINTS, CENTROIDS, clusterAssessment, md, results)

	for i := 0; i < r; i++ {
		go doPairPointCentroidJobs(done, jobs)
	}
	go awaitPairPointCentroidCompletion(done, results)

    clusterChanged := assessClusters(clusterAssessment, results)
	if clusterChanged != true {
		t.Errorf("TestAssessClusters: clusterChanged=%b and should be true.", clusterChanged)
	}

	if clusterAssessment.Get(9, 0) != 0 || clusterAssessment.Get(10, 0) != 1 {
		t.Errorf("TestAssessClusters: rows 9 and 10 should have 0 and 1 in column 0, but received %v", clusterAssessment)
	}
}

/* TODO rewrite for new version
func TestKmeansbi(t *testing.T) {
	var ed matutil.EuclidDist
	var cc RandCentroids

	matCentroidlist, clusterAssignment, err := Kmeansp(DATAPOINTS, 4, cc, ed)
	if err != nil {
		t.Errorf("Kmeans returned: %v", err)
		return
	}

	if 	a, b := matCentroidlist.GetSize(); a == 0 || b == 0 {
		t.Errorf("Kmeans centroidMeans is of size %d, %d.", a,b)
	}

	if c, d := clusterAssignment.GetSize(); c == 0 || d == 0 {
		t.Errorf("Kmeans clusterAssessment is of size %d, %d.", c,d)
	}
	// TODO deterministic test
}
*/
  
func TestPointProb(t *testing.T) {
	R := 10010.0
	Ri := 100.0
	M := 2.0
	V := 20.000000

	point := matrix.MakeDenseMatrix([]float64{5, 7},
		1,2)

	mean := matrix.MakeDenseMatrix([]float64{6, 8},
		1,2)

	var ed matutil.EuclidDist 

	//	pointProb(R, Ri, M int, V float64, point, mean *matrix.DenseMatrix, measurer matutil.VectorMeasurer) (float64, error) 
	pp := pointProb(R, Ri, M, V, point, mean, ed)

	E :=  0.011473
	epsilon := .000001
	na := math.Nextafter(E, E + 1) 
	diff := math.Abs(pp - na) 

	if diff > epsilon {
		t.Errorf("TestPointProb: expected %f but received %f.  The difference %f exceeds epsilon %f", E, pp, diff, epsilon)
	}
}

func TestFreeparams(t *testing.T) {
	K := 6
	M := 3

	r := freeparams(K, M)
	if r != 24 {
		t.Errorf("TestFreeparams: Expected 24 but received %f.", r)
	}
}

func TestVariance(t *testing.T) {
	var ed matutil.EuclidDist
	_, dim := DATAPOINTS_D.GetSize()
	// Model D
	c := cluster{DATAPOINTS_D, CENTROIDS_D, dim, 0}
	v := variance(c, ed)
	
	E := 24.000
	epsilon := .000001
	na := math.Nextafter(E, E + 1) 
	diff := math.Abs(v - na) 

	if diff > epsilon {
		t.Errorf("TestVariance: for model D excpected %f but received %f.  The difference %f exceeds epsilon %f.", E, v, diff, epsilon)
	}

	// Variance a cluster with a perfectly centered centroids
	_, dim0 := DATAPOINTS_D0.GetSize()
	c0 := cluster{DATAPOINTS_D0, CENTROID_D0, dim0, 0}
	v0 := variance(c0, ed)
	
	E = 2.00
	na = math.Nextafter(E, E + 1) 
	diff = math.Abs(v0 - na) 

	if diff > epsilon {
		t.Errorf("TestVariance: for model D excpected %f but received %f.  The difference %f exceeds epsilon %f.", E, v0, diff, epsilon)
	}
}

/*
func TestLogLikelih(t *testing.T) {
	// Model D
	R, M := DATAPOINTS_D.GetSize()
	K, _ := CENTROIDS_D.GetSize()

	clusterAssessment := makeClusterAssessment(DATAPOINTS_D, CENTROIDS_D)
	var ed matutil.EuclidDist
	vari := variance(DATAPOINTS_D, CENTROIDS_D, clusterAssessment, K, ed)

	cs := make([]cluster, 1)
	cs[0] = cluster{DATAPOINTS_D, CENTROIDS_D, clusterAssessment, M, vari}

	ll := loglikelih(R, cs)

	epsilon := .000001
	E := -35.042733
	na := math.Nextafter(E, E + 1) 
	diff := math.Abs(ll - na) 

	if diff > epsilon {
		t.Errorf("TestLoglikeli: For model D expected %f but received %f.  The difference %f exceeds epsilon %f", E, ll, diff, epsilon)
	}

	// Model Dn with two cluster with one centroid each
	K = 1
	datapoints_0, err := DATAPOINTS_D.FiltCol(0,4, 0)
	if err != nil {
		t.Errorf("TestLoglikelih: FiltCol for datapoints_0 err=%v", err)
	}

	centroids_0 := matrix.MakeDenseMatrix([]float64{2.0,3.0}, 1,2)
	clusterAssessment_0 := makeClusterAssessment(datapoints_0, centroids_0)
	var0 := variance(datapoints_0, centroids_0, clusterAssessment_0, K, ed)
	cluster_0 := cluster{datapoints_0, centroids_0, clusterAssessment_0, M, var0}


	datapoints_1, err := DATAPOINTS_D.FiltCol(5, 100, 0)
	if err != nil {
		t.Errorf("TestLoglikelih: FiltCol for datapoints_1 err=%v", err)
	}

	centroids_1 := matrix.MakeDenseMatrix([]float64{9.0,7.0}, 1,2)
	clusterAssessment_1 := makeClusterAssessment(datapoints_1, centroids_1)
	var1 := variance(datapoints_1, centroids_1, clusterAssessment_1, K, ed)
	cluster_1 := cluster{datapoints_1, centroids_1, clusterAssessment_1, M, var1}

	cs_n := []cluster{cluster_0, cluster_1}

	ll_n := loglikelih(R, cs_n)

	E = -20.970731
	na = math.Nextafter(E, E + 1) 
	diff = math.Abs(ll_n - na) 

	if diff > epsilon {
		t.Errorf("TestLoglikeli: For model Dn expected %f but received %f.  The difference %f exceeds epsilon %f", E, ll_n, diff, epsilon)
	}
	
}
*/

// Create two tight clusters and test the scores for a model with 1 centroid 
// that is equidistant between the two and a model with 2 centroids where 
// the centroids are in the center of each cluster.
// 
// The BIC of the second should always be better.
//
//Model 1
//                                     *
//                                  *     *
//                                     *
//                        +
//
//           *
//        *     *
//           *
//
//Model 2
//                                     *
//                                  *  +  *
//                                     *
//                        
//
//           *
//        *  +  *
//           *
//
/*func TestBic(t *testing.T) {
	R, M := DATAPOINTS.GetSize()
//	Rn := []int{R} // for testing a model without a parent
	K := 1

	clusterAssessment := makeClusterAssessment(DATAPOINTS_D, CENTROIDS_D)
	numparams := freeparams(K, M)
	var ed matutil.EuclidDist
	vari := variance(DATAPOINTS_D, CENTROIDS_D, clusterAssessment, 1, ed)
//	fmt.Printf("var1=%f\n", v1)

	c := []cluster{cluster{DATAPOINTS_D, CENTROIDS_D, clusterAssessment, M, vari}}
	loglikeh1 := loglikelih(R, c)
//	fmt.Printf("loglikelihood1 = %f\n", loglikeh1)

	bic1 := bic(loglikeh1, numparams, R)
//	fmt.Printf("bic1=%f\n", bic1)
	
	// Model 2
	K = 1
	numparamsnew := freeparams(K, M)

	datapoints_0 := matrix.MakeDenseMatrix([]float64{2,3, 3,2, 3,4, 4,3},  4,2 )
	datapoints_1 := matrix.MakeDenseMatrix( []float64{8,7, 9,6, 9,8, 10,7}, 4, 2)
	newcents_0 := matrix.MakeDenseMatrix([]float64{2,3}, 1,2)
	newcents_1 := matrix.MakeDenseMatrix([]float64{9,7}, 1,2)
	newca := makeClusterAssessment(DATAPOINTS_D, CENTROIDS_D)

	v0 := variance(datapoints_0, newcents_0, newca,1,  ed)
	v1 := variance(datapoints_1, newcents_1, newca, 1, ed)
	fmt.Println(v0, v1)
	cnew := make([]cluster, 2)
	cnew[0] = cluster{datapoints_0, newcents_0, newca, M, v0}
	cnew[1] = cluster{datapoints_1, newcents_1, newca, M, v1}

	loglikehnew := loglikelih(R, cnew)
//	fmt.Printf("loglikelihood2 = %f\n", loglikeh2)

	bic2 := bic(loglikehnew, numparamsnew, R)
//	fmt.Printf("bic2=%f\n", bic2)

	if bic1 >= bic2 {
		t.Errorf("TestBicComp: bic2 should be greater than bic1, but received bic1=%f and bic2=%f", bic1, bic2)
	}

}
*/