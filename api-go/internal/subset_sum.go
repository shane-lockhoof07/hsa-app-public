package internal

type state struct {
    sum     float64
    indices []int
}

func SubsetSum(amounts []float64, target float64) []int {
    dp := []state{{0, []int{}}}
    best := state{}
    
    for i, a := range amounts {
        newdp := make([]state, len(dp))
        copy(newdp, dp)
        
        for _, s := range dp {
            ns := s.sum + a
            if ns <= target && (target-ns) < (target-best.sum) {
                newIndices := make([]int, len(s.indices))
                copy(newIndices, s.indices)
                newIndices = append(newIndices, i)
                best = state{ns, newIndices}
            }
            
            newIndices := make([]int, len(s.indices))
            copy(newIndices, s.indices)
            newIndices = append(newIndices, i)
            newdp = append(newdp, state{ns, newIndices})
        }
        dp = newdp
    }
    
    return best.indices
}