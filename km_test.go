package goxmeans

import (
	"bufio"
	"code.google.com/p/gomatrix/matrix"
	"fmt"
	"os"
	"testing"
)

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
		t.Errorf("Could not create test file %s err=%v", err)
	}
	defer os.Remove(fname)

	if _, err := Load(fname); err != nil {
		t.Errorf("Load(%s) err=%v", fname, err)
	}
}

func TestRandCentroids(t *testing.T) {
	rows := 3
	cols := 3
	k := 4
	data := []float64{1, 2.0, 3, -4.945, 5, -6.1, 7, 8, 9}
	mat := matrix.MakeDenseMatrix(data, rows, cols)
	centroids := RandCentroids(mat, k)

	r, c := centroids.GetSize()
	if r != k || c != cols {
		t.Errorf("Returned centroid was %dx%d instead of %dx%d", r, c, rows, cols)
	}
}

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

func TestAssignPointToCentroid(t *testing.T) {
	centroids := matrix.MakeDenseMatrix([]float64{1.0,1.0,100.0,100.0}, 2, 2)
	datapoint := matrix.MakeDenseMatrix([]float64{2.0, 2.0}, 1, 2)
	minIndex, _, err := AssignPointToCentroid(datapoint,centroids)
	if err != nil {
		t.Errorf("AssignCentroid returned: %v", err)
	}
	if minIndex != 0 {
		t.Errorf("AssignCentroid returned minIndex=%f instead of MinIndex=0.", minIndex)
	}
}

func TestAssignPointToCentroidErr(t *testing.T) {
	centroids := matrix.MakeDenseMatrix([]float64{1.0,1.0,100.0,100.0}, 2, 2)
	datapoint := matrix.Zeros(4,4)
	_, _, err := AssignPointToCentroid(datapoint,centroids)
	if err == nil {
		t.Errorf("AssginCentroid should returned error.  Passed datapoint matrix with %d rows.", 4)
	}
}