package core

// computeSections splits a file of the given size into n contiguous byte
// ranges. Integer division remainder is absorbed by the final section, so the
// ranges always cover [0, size-1] with no gaps or overlaps.
func computeSections(size int64, n int) [][2]int64 {
	secs := make([][2]int64, n)
	each := size / int64(n)
	for i := 0; i < n; i++ {
		if i == 0 {
			secs[i][0] = 0
		} else {
			secs[i][0] = secs[i-1][1] + 1
		}
		if i < n-1 {
			secs[i][1] = secs[i][0] + each
		} else {
			secs[i][1] = size - 1
		}
	}
	return secs
}

// remainingRange returns the byte range still needed for a section given how
// many bytes are already on disk, and whether anything remains.
func remainingRange(sec [2]int64, have int64) (start, end int64, ok bool) {
	total := sec[1] - sec[0] + 1
	if have >= total {
		return 0, 0, false
	}
	return sec[0] + have, sec[1], true
}

// ewma is an exponentially weighted moving average; alpha weights the new
// sample. The first sample (prev==0) is taken as-is.
func ewma(prev, sample, alpha float64) float64 {
	if prev == 0 {
		return sample
	}
	return alpha*sample + (1-alpha)*prev
}

// etaSeconds estimates seconds remaining; returns -1 when speed is unknown.
func etaSeconds(remaining int64, speed float64) int {
	if speed <= 0 {
		return -1
	}
	return int(float64(remaining) / speed)
}
