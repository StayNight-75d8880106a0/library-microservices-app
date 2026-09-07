package helper

import "math"

type ErlangCResult struct {
	Utilization    float64 // ρ (Rho)
	ProbWait       float64 // P(W > 0)
	AvgQueueLen    float64 // Lq
	AvgWaitMin     float64 // Wq (menit)
	OptimalServers int     // c* (Rekomendasi jumlah stok ideal)
	IsSaturated    bool    // Menandakan apakah sistem jenuh (ρ >= 1)
}

func Factorial(num int) float64 {
	if num <= 1 {
		return 1.0
	}

	result := 1.0

	for i := 2; i <= num; i++ {
		result *= float64(i)
	}

	return result

}

func CalculateErlangC(lambda float64, mu float64, c int) ErlangCResult {
	if c <= 0 {
		c = 1
	}

	a := lambda / mu      // Traffic Intensity (Erlangs)
	rho := a / float64(c) // System Utilization

	// Jika sistem jenuh (rho >= 1), batasi rho agar tidak panic/division by zero
	effectiveRho := rho
	if effectiveRho >= 1.0 {
		effectiveRho = 0.99
	}

	// 1. Hitung P(0) - Probabilitas sistem kosong
	var sum float64
	for k := 0; k < c; k++ {
		sum += math.Pow(a, float64(k)) / Factorial(k)
	}

	erlangTerm := (math.Pow(a, float64(c)) / Factorial(c)) * (1.0 / (1.0 - effectiveRho))
	p0 := 1.0 / (sum + erlangTerm)

	// 2. Hitung Probabilitas Menunggu P(Wait)
	probWait := erlangTerm * p0
	if probWait > 1.0 {
		probWait = 1.0
	}

	// 3. Rata-rata Panjang Antrean (Lq)
	lq := (probWait * effectiveRho) / (1.0 - effectiveRho)

	// 4. Rata-rata Waktu Menunggu dalam Menit (Wq)
	wqHours := lq / lambda
	wqMinutes := wqHours * 60.0

	// 5. Hitung Optimal Servers (c*) -> Cari jumlah server minimum agar Rho < 0.8 (beban sehat)
	optimalC := c
	for {
		if a/float64(optimalC) < 0.8 {
			break
		}
		optimalC++
	}

	if mu <= 0 || lambda < 0 {
		return ErlangCResult{OptimalServers: c, IsSaturated: true}
	}

	return ErlangCResult{
		Utilization:    rho,
		ProbWait:       probWait,
		AvgQueueLen:    lq,
		AvgWaitMin:     wqMinutes,
		OptimalServers: optimalC,
	}
}
