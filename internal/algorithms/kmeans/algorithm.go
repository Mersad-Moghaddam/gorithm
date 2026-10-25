package kmeans

import "math"

// Point represents a 2D coordinate for clustering.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Cluster executes K-Means clustering in 2D space.
// ponytail: uniform index stride centroid initialization used for deterministic simplicity; upgrade to K-Means++ with probabilistic seeding if handling adversarial cluster distributions.
func Cluster(points []Point, k, maxIters int) ([]Point, []int) {
	if len(points) == 0 || k <= 0 {
		return nil, nil
	}
	if k > len(points) {
		k = len(points)
	}
	if maxIters <= 0 {
		maxIters = 100
	}

	centroids := make([]Point, k)
	step := len(points) / k
	for i := 0; i < k; i++ {
		centroids[i] = points[i*step]
	}

	assignments := make([]int, len(points))
	for i := range assignments {
		assignments[i] = -1
	}

	for iter := 0; iter < maxIters; iter++ {
		changed := false
		for i, p := range points {
			bestCluster := nearestCentroid(p, centroids)
			if assignments[i] != bestCluster {
				assignments[i] = bestCluster
				changed = true
			}
		}

		counts := make([]int, k)
		sums := make([]Point, k)
		for i, p := range points {
			c := assignments[i]
			counts[c]++
			sums[c].X += p.X
			sums[c].Y += p.Y
		}

		moved := false
		for j := 0; j < k; j++ {
			if counts[j] > 0 {
				newCentroid := Point{
					X: sums[j].X / float64(counts[j]),
					Y: sums[j].Y / float64(counts[j]),
				}
				if distSq(centroids[j], newCentroid) > 1e-9 {
					moved = true
				}
				centroids[j] = newCentroid
			}
		}

		if iter > 0 && !changed && !moved {
			break
		}
	}

	return centroids, assignments
}

func distSq(p1, p2 Point) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return dx*dx + dy*dy
}

func nearestCentroid(p Point, centroids []Point) int {
	bestIdx := 0
	bestDist := math.Inf(1)
	for i, c := range centroids {
		d := distSq(p, c)
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	return bestIdx
}
