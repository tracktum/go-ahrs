package ahrs

import "math"

const (
	MahonyDefaultKp = 0.2
	MahonyDefaultKi = 0.1
)

// Mahony instance
type Mahony struct {
	twoKp, twoKi                          float64
	integralFBx, integralFBy, integralFBz float64

	SampleFreq  float64
	Quaternions [4]float64
}

// NewMahony initiates a Mahony struct
func NewMahony(kp, ki, sampleFreq float64) Mahony {
	return Mahony{
		twoKp: 2 * kp,
		twoKi: 2 * ki,

		SampleFreq:  sampleFreq,
		Quaternions: [4]float64{1, 0, 0, 0},
	}
}

// NewDefaultMahony initiates a Mahony struct with default ki and kp
func NewDefaultMahony(sampleFreq float64) Mahony {
	return Mahony{
		twoKp: 2 * MahonyDefaultKp,
		twoKi: 2 * MahonyDefaultKi,

		SampleFreq:  sampleFreq,
		Quaternions: [4]float64{1, 0, 0, 0},
	}
}

// Update9D updates position using 9D, returning quaternions
func (m *Mahony) Update9D(gx, gy, gz, ax, ay, az, mx, my, mz float64) [4]float64 {
	var recipNorm float64
	var q0q0, q0q1, q0q2, q0q3, q1q1, q1q2, q1q3, q2q2, q2q3, q3q3 float64
	var hx, hy, bx, bz float64
	var halfvx, halfvy, halfvz, halfwx, halfwy, halfwz float64
	var halfex, halfey, halfez float64
	var qa, qb, qc float64

	q0 := m.Quaternions[0]
	q1 := m.Quaternions[1]
	q2 := m.Quaternions[2]
	q3 := m.Quaternions[3]
	integralFBx := m.integralFBx
	integralFBy := m.integralFBy
	integralFBz := m.integralFBz
	twoKi := m.twoKi
	twoKp := m.twoKp
	sampleFreq := m.SampleFreq

	// Compute feedback only if accelerometer measurement valid (avoids NaN in accelerometer normalisation)
	if !(ax == 0.0 && ay == 0.0 && az == 0.0) {
		// Normalise accelerometer measurement
		recipNorm = invSqrt(ax*ax + ay*ay + az*az)
		ax *= recipNorm
		ay *= recipNorm
		az *= recipNorm

		// Normalise magnetometer measurement
		recipNorm = invSqrt(mx*mx + my*my + mz*mz)
		mx *= recipNorm
		my *= recipNorm
		mz *= recipNorm

		// Auxiliary variables to avoid repeated arithmetic
		q0q0 = q0 * q0
		q0q1 = q0 * q1
		q0q2 = q0 * q2
		q0q3 = q0 * q3
		q1q1 = q1 * q1
		q1q2 = q1 * q2
		q1q3 = q1 * q3
		q2q2 = q2 * q2
		q2q3 = q2 * q3
		q3q3 = q3 * q3

		// Reference direction of Earth's magnetic field
		hx = 2.0 * (mx*(0.5-q2q2-q3q3) + my*(q1q2-q0q3) + mz*(q1q3+q0q2))
		hy = 2.0 * (mx*(q1q2+q0q3) + my*(0.5-q1q1-q3q3) + mz*(q2q3-q0q1))
		bx = math.Sqrt(hx*hx + hy*hy)
		bz = 2.0 * (mx*(q1q3-q0q2) + my*(q2q3+q0q1) + mz*(0.5-q1q1-q2q2))

		// Estimated direction of gravity and magnetic field
		halfvx = q1q3 - q0q2
		halfvy = q0q1 + q2q3
		halfvz = q0q0 - 0.5 + q3q3
		halfwx = bx*(0.5-q2q2-q3q3) + bz*(q1q3-q0q2)
		halfwy = bx*(q1q2-q0q3) + bz*(q0q1+q2q3)
		halfwz = bx*(q0q2+q1q3) + bz*(0.5-q1q1-q2q2)

		// Error is sum of cross product between estimated direction and measured direction of field vectors
		halfex = (ay*halfvz - az*halfvy) + (my*halfwz - mz*halfwy)
		halfey = (az*halfvx - ax*halfvz) + (mz*halfwx - mx*halfwz)
		halfez = (ax*halfvy - ay*halfvx) + (mx*halfwy - my*halfwx)

		// Compute and apply integral feedback if enabled
		if twoKi > 0.0 {
			integralFBx += twoKi * halfex * (1.0 / sampleFreq) // integral error scaled by Ki
			integralFBy += twoKi * halfey * (1.0 / sampleFreq)
			integralFBz += twoKi * halfez * (1.0 / sampleFreq)
			gx += integralFBx // apply integral feedback
			gy += integralFBy
			gz += integralFBz
		} else {
			integralFBx = 0.0 // prevent integral windup
			integralFBy = 0.0
			integralFBz = 0.0
		}

		// Apply proportional feedback
		gx += twoKp * halfex
		gy += twoKp * halfey
		gz += twoKp * halfez
	}

	// Integrate rate of change of quaternion
	gx *= (0.5 * (1.0 / sampleFreq)) // pre-multiply common factors
	gy *= (0.5 * (1.0 / sampleFreq))
	gz *= (0.5 * (1.0 / sampleFreq))
	qa = q0
	qb = q1
	qc = q2
	q0 += (-qb*gx - qc*gy - q3*gz)
	q1 += (qa*gx + qc*gz - q3*gy)
	q2 += (qa*gy - qb*gz + q3*gx)
	q3 += (qa*gz + qb*gy - qc*gx)

	// Normalise quaternion
	recipNorm = invSqrt(q0*q0 + q1*q1 + q2*q2 + q3*q3)
	m.Quaternions[0] = q0 * recipNorm
	m.Quaternions[1] = q1 * recipNorm
	m.Quaternions[2] = q2 * recipNorm
	m.Quaternions[3] = q3 * recipNorm

	return m.Quaternions
}

// Update6D updates position using 6D, returning quaternions
func (m *Mahony) Update6D(gx, gy, gz, ax, ay, az float64) [4]float64 {
	var recipNorm float64
	var halfvx, halfvy, halfvz float64
	var halfex, halfey, halfez float64
	var qa, qb, qc float64

	q0 := m.Quaternions[0]
	q1 := m.Quaternions[1]
	q2 := m.Quaternions[2]
	q3 := m.Quaternions[3]
	integralFBx := m.integralFBx
	integralFBy := m.integralFBy
	integralFBz := m.integralFBz
	twoKi := m.twoKi
	twoKp := m.twoKp
	sampleFreq := m.SampleFreq

	// Compute feedback only if accelerometer measurement valid (avoids NaN in accelerometer normalisation)
	if !(ax == 0.0 && ay == 0.0 && az == 0.0) {

		// Normalise accelerometer measurement
		recipNorm = invSqrt(ax*ax + ay*ay + az*az)
		ax *= recipNorm
		ay *= recipNorm
		az *= recipNorm

		// Estimated direction of gravity and vector perpendicular to magnetic flux
		halfvx = q1*q3 - q0*q2
		halfvy = q0*q1 + q2*q3
		halfvz = q0*q0 - 0.5 + q3*q3

		// Error is sum of cross product between estimated and measured direction of gravity
		halfex = (ay*halfvz - az*halfvy)
		halfey = (az*halfvx - ax*halfvz)
		halfez = (ax*halfvy - ay*halfvx)

		// Compute and apply integral feedback if enabled
		if twoKi > 0.0 {
			integralFBx += twoKi * halfex * (1.0 / sampleFreq) // integral error scaled by Ki
			integralFBy += twoKi * halfey * (1.0 / sampleFreq)
			integralFBz += twoKi * halfez * (1.0 / sampleFreq)
			gx += integralFBx // apply integral feedback
			gy += integralFBy
			gz += integralFBz
		} else {
			integralFBx = 0.0 // prevent integral windup
			integralFBy = 0.0
			integralFBz = 0.0
		}

		// Apply proportional feedback
		gx += twoKp * halfex
		gy += twoKp * halfey
		gz += twoKp * halfez
	}

	// Integrate rate of change of quaternion
	gx *= (0.5 * (1.0 / sampleFreq)) // pre-multiply common factors
	gy *= (0.5 * (1.0 / sampleFreq))
	gz *= (0.5 * (1.0 / sampleFreq))
	qa = q0
	qb = q1
	qc = q2
	q0 += (-qb*gx - qc*gy - q3*gz)
	q1 += (qa*gx + qc*gz - q3*gy)
	q2 += (qa*gy - qb*gz + q3*gx)
	q3 += (qa*gz + qb*gy - qc*gx)

	// Normalise quaternion
	recipNorm = invSqrt(q0*q0 + q1*q1 + q2*q2 + q3*q3)
	m.Quaternions[0] = q0 * recipNorm
	m.Quaternions[1] = q1 * recipNorm
	m.Quaternions[2] = q2 * recipNorm
	m.Quaternions[3] = q3 * recipNorm

	return m.Quaternions
}
