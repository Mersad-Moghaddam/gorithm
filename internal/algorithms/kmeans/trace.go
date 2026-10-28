package kmeans

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"math"
)

func validateKMeansInput(points []Point, k int) error {
	if len(points) < 1 || len(points) > 64 || k < 1 || k > 64 {
		return fmt.Errorf("require 1–64 points and k in [1,64]")
	}
	// Bound arithmetic, not just coordinates: all pair distances and sums must remain finite.
	sx, sy := 0.0, 0.0
	for _, p := range points {
		if math.IsNaN(p.X) || math.IsInf(p.X, 0) || math.IsNaN(p.Y) || math.IsInf(p.Y, 0) {
			return fmt.Errorf("points must be finite")
		}
		if math.Abs(p.X) > math.MaxFloat64/4 || math.Abs(p.Y) > math.MaxFloat64/4 {
			return fmt.Errorf("point domain padding would overflow")
		}
		if math.IsInf(p.X*p.X, 0) || math.IsInf(p.Y*p.Y, 0) {
			return fmt.Errorf("coordinates overflow squared distance")
		}
		sx += math.Abs(p.X)
		sy += math.Abs(p.Y)
		for _, q := range points {
			if math.IsInf(distSq(p, q), 0) || math.IsNaN(distSq(p, q)) {
				return fmt.Errorf("point distances overflow")
			}
		}
	}
	if math.IsInf(sx, 0) || math.IsInf(sy, 0) {
		return fmt.Errorf("point sums overflow")
	}
	return nil
}

// Trace records actual assignment and mean operations, preserving stride seeds and first-index ties.
func Trace(points []Point, k, maxIters int) []trace.Step {
	b := trace.NewBuilder()
	s := renderers.KMeansLessonState{RequestedK: k, ActivePoint: -1, Closest: -1, PreviousAssignment: -1, XMin: -1, XMax: 1, YMin: -1, YMax: 1}
	if err := validateKMeansInput(points, k); err != nil {
		s.Phase = "invalid-input"
		s.StopReason = "invalid-input"
		b.AddStep("L1", trace.ActionInit, err.Error(), s, nil)
		b.AddStep("L16", trace.ActionDone, "Clustering stopped: invalid input.", s, map[string]any{"converged": false, "stopReason": s.StopReason})
		return b.Steps()
	}
	if k > len(points) {
		k = len(points)
	}
	if maxIters <= 0 {
		maxIters = 100
	}
	s.Points = make([][2]float64, len(points))
	s.Assignments = make([]int, len(points))
	for i, p := range points {
		s.Points[i] = [2]float64{p.X, p.Y}
		s.Assignments[i] = -1
	}
	emit := func(line string, action trace.ActionType, phase, reason string) {
		s.Phase = phase
		c := s
		c.Points = append([][2]float64(nil), s.Points...)
		c.Centroids = append([][2]float64(nil), s.Centroids...)
		c.PreviousCentroids = append([][2]float64(nil), s.PreviousCentroids...)
		c.Sums = append([][2]float64(nil), s.Sums...)
		c.Counts = append([]int(nil), s.Counts...)
		c.Assignments = append([]int(nil), s.Assignments...)
		c.DistanceSquared = append([]float64(nil), s.DistanceSquared...)
		b.AddStep(line, action, reason, c, map[string]any{"iteration": s.Iteration, "requestedK": s.RequestedK, "effectiveK": k, "point": s.ActivePoint, "closest": s.Closest, "previousAssignment": s.PreviousAssignment, "changed": s.Changed, "moved": s.Moved, "converged": s.Converged, "stopReason": s.StopReason})
	}
	emit("L1", trace.ActionInit, "data", fmt.Sprintf("%d points; requested k=%d, effective k=%d.", len(points), s.RequestedK, k))
	for j := 0; j < k; j++ {
		s.Centroids = append(s.Centroids, s.Points[j*(len(points)/k)])
	}
	b.IncOperations()
	emit("L2", trace.ActionSet, "seeds", "Seed centroids by uniform input-index stride.")
	emit("L3", trace.ActionInit, "unassigned", "Assignments start at -1: no point belongs to a cluster yet.")
	for iter := 1; iter <= maxIters; iter++ {
		s.Iteration = iter
		s.ActivePoint = -1
		s.Closest = -1
		s.DistanceSquared = nil
		s.PreviousCentroids = nil
		s.Sums = nil
		s.Counts = nil
		emit("L4", trace.ActionHighlight, "iteration", fmt.Sprintf("Iteration %d: assign each point using current centroids.", iter))
		s.Changed = false
		s.Moved = false
		emit("L5", trace.ActionSet, "reset", "Reset membership-change flag.")
		for i, p := range points {
			s.ActivePoint = i
			s.Closest = -1
			s.PreviousAssignment = s.Assignments[i]
			s.DistanceSquared = nil
			emit("L6", trace.ActionVisit, "focus", fmt.Sprintf("P%d = (%g, %g): compare squared distances.", i, p.X, p.Y))
			best := 0
			for j, c := range s.Centroids {
				d := distSq(p, Point{c[0], c[1]})
				s.DistanceSquared = append(s.DistanceSquared, d)
				b.IncComparisons()
				if d < s.DistanceSquared[best] {
					best = j
				}
			}
			s.Closest = best
			emit("L7", trace.ActionCompare, "distances", fmt.Sprintf("P%d chooses C%d: minimum d²=%g; ties keep first centroid.", i, best, s.DistanceSquared[best]))
			emit("L8", trace.ActionCompare, "membership-check", fmt.Sprintf("P%d membership: %d versus %d.", i, s.PreviousAssignment, best))
			if s.Assignments[i] != best {
				s.Assignments[i] = best
				b.IncSwaps()
				emit("L9", trace.ActionSet, "assign", fmt.Sprintf("P%d now belongs to C%d; point position stays fixed.", i, best))
				s.Changed = true
				emit("L10", trace.ActionSet, "changed", "Membership changed; another iteration may be needed.")
			}
		}
		s.ActivePoint = -1
		s.Closest = -1
		s.DistanceSquared = nil
		s.Counts = make([]int, k)
		s.Sums = make([][2]float64, k)
		s.PreviousCentroids = append([][2]float64(nil), s.Centroids...)
		for i, p := range s.Points {
			j := s.Assignments[i]
			s.Counts[j]++
			s.Sums[j][0] += p[0]
			s.Sums[j][1] += p[1]
		}
		for j := range s.Centroids {
			if s.Counts[j] > 0 {
				next := [2]float64{s.Sums[j][0] / float64(s.Counts[j]), s.Sums[j][1] / float64(s.Counts[j])}
				if distSq(Point{s.Centroids[j][0], s.Centroids[j][1]}, Point{next[0], next[1]}) > 1e-9 {
					s.Moved = true
				}
				s.Centroids[j] = next
			}
		}
		b.IncOperations()
		emit("L13", trace.ActionSet, "means", "Move each nonempty centroid to sum / count; retain empty centroids.")
		s.Converged = !s.Changed && !s.Moved
		emit("L14", trace.ActionCompare, "convergence", fmt.Sprintf("changed=%t, moved=%t; stable only when both are false.", s.Changed, s.Moved))
		if s.Converged {
			break
		}
	}
	s.StopReason = "iteration-limit"
	if s.Converged {
		s.StopReason = "stable"
	}
	emit("L16", trace.ActionDone, "done", fmt.Sprintf("Stopped after %d iterations: %s.", s.Iteration, s.StopReason))
	steps := b.Steps()
	xmin, xmax, ymin, ymax := points[0].X, points[0].X, points[0].Y, points[0].Y
	for _, p := range points {
		xmin = math.Min(xmin, p.X)
		xmax = math.Max(xmax, p.X)
		ymin = math.Min(ymin, p.Y)
		ymax = math.Max(ymax, p.Y)
	}
	pad := func(lo, hi float64) (float64, float64) {
		p := (hi - lo) * .1
		if hi == lo {
			p = math.Abs(lo) * .1
		}
		if p < 1 {
			p = 1
		}
		return lo - p, hi + p
	}
	xmin, xmax = pad(xmin, xmax)
	ymin, ymax = pad(ymin, ymax)
	for i := range steps {
		c := steps[i].State.(renderers.KMeansLessonState)
		c.XMin = xmin
		c.XMax = xmax
		c.YMin = ymin
		c.YMax = ymax
		steps[i].State = c
		if steps[i].LineID == "L8" && i+1 < len(steps) && steps[i+1].LineID == "L9" {
			steps[i].HoldMS = 600
			steps[i+1].HoldMS = 1200
		}
	}
	return steps
}
