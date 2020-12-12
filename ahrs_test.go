package ahrs_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tracktum/go-ahrs"
	"gonum.org/v1/gonum/num/quat"
)

const dataFile = "./data.csv"

type dataLine struct {
	time float64
	acce [3]float64
	gyro [3]float64
	magn [3]float64
}

func readCSV(fileName string) ([]dataLine, error) {
	result := make([]dataLine, 0)

	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	data := string(bytes)
	for _, line := range strings.Split(data, "\n")[1:] {
		if line == "" {
			continue
		}
		var value dataLine
		fmt.Sscanf(
			line, "%f;%f;%f;%f;%f;%f;%f;%f;%f;%f",
			&value.time,
			&value.acce[0], &value.acce[1], &value.acce[2],
			&value.gyro[0], &value.gyro[1], &value.gyro[2],
			&value.magn[0], &value.magn[1], &value.magn[2],
		)
		result = append(result, value)
	}
	return result, nil
}

func TestReadCSV(t *testing.T) {
	data, err := readCSV(dataFile)
	require.NoError(t, err)
	require.Equal(t, 8993, len(data))

	first := data[0]
	require.Equal(t, float64(0.08), first.time)
	require.Equal(t, float64(-0.071594), first.acce[0])
	require.Equal(t, float64(0.21157), first.acce[1])
	require.Equal(t, float64(9.7958), first.acce[2])
	require.Equal(t, float64(0.002314), first.gyro[0])
	require.Equal(t, float64(-0.00634), first.gyro[1])
	require.Equal(t, float64(0.001322), first.gyro[2])
	require.Equal(t, float64(-0.33533), first.magn[0])
	require.Equal(t, float64(0.19856), first.magn[1])
	require.Equal(t, float64(-0.88708), first.magn[2])

	last := data[len(data)-1]
	require.Equal(t, float64(90), last.time)
	require.Equal(t, float64(-0.07637), last.acce[0])
	require.Equal(t, float64(2.61), last.acce[1])
	require.Equal(t, float64(-8.8695), last.acce[2])
	require.Equal(t, float64(0.37059), last.gyro[0])
	require.Equal(t, float64(-0.026818), last.gyro[1])
	require.Equal(t, float64(-0.017078), last.gyro[2])
	require.Equal(t, float64(-0.35742), last.magn[0])
	require.Equal(t, float64(-0.4499), last.magn[1])
	require.Equal(t, float64(0.84045), last.magn[2])
}

// TestAHRS tests if diff between madwick and mahony is less than 0.2
func TestAHRS(t *testing.T) {
	data, err := readCSV(dataFile)
	require.NoError(t, err)

	madgwick := ahrs.NewMadgwick(0.1, 100)
	mahony := ahrs.NewDefaultMahony(100)
	var e float64
	for _, d := range data {
		q := madgwick.Update9D(
			d.gyro[0], d.gyro[1], d.gyro[1],
			d.acce[0], d.acce[1], d.acce[1],
			d.magn[0], d.magn[1], d.magn[1],
		)
		q1 := quat.Number{q[0], q[1], q[2], q[3]}
		q = mahony.Update9D(
			d.gyro[0], d.gyro[1], d.gyro[1],
			d.acce[0], d.acce[1], d.acce[1],
			d.magn[0], d.magn[1], d.magn[1],
		)
		q2 := quat.Number{q[0], q[1], q[2], q[3]}

		diff := quat.Abs(quat.Sub(q1, q2))
		e += diff
	}
	e /= float64(len(data))
	require.LessOrEqual(t, e, 0.2)
}

func BenchmarkMadgwick(b *testing.B) {
	data, err := readCSV(dataFile)
	l := len(data)
	require.NoError(b, err)
	madgwick := ahrs.NewMadgwick(0.1, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := data[i%l]
		madgwick.Update9D(
			d.gyro[0], d.gyro[1], d.gyro[1],
			d.acce[0], d.acce[1], d.acce[1],
			d.magn[0], d.magn[1], d.magn[1],
		)
	}
}

func BenchmarkMahony(b *testing.B) {
	data, err := readCSV(dataFile)
	l := len(data)
	require.NoError(b, err)
	mahony := ahrs.NewDefaultMahony(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := data[i%l]
		mahony.Update9D(
			d.gyro[0], d.gyro[1], d.gyro[1],
			d.acce[0], d.acce[1], d.acce[1],
			d.magn[0], d.magn[1], d.magn[1],
		)
	}
}
