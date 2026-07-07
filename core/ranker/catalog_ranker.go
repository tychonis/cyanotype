package ranker

// import (
// 	"sort"

// 	"github.com/tychonis/cyanotype/core/catalog"
// 	"github.com/tychonis/cyanotype/core/process"
// )

// type CatalogRanker struct {
// 	catalog *catalog.Catalog
// }

// func (r *CatalogRanker) RankCoProcess(cps []*process.CoProcess) ([]*process.CoProcess, error) {
// 	sort.SliceStable(cps, func(i, j int) bool {
// 		rankI, errI := r.GetCoProcessRank(cps[i])
// 		rankJ, errJ := r.GetCoProcessRank(cps[j])

// 		if errI != nil || errJ != nil {
// 			return false
// 		}

// 		return catalog.CmpRank(*rankI, *rankJ) < 0
// 	})
// 	return cps, nil
// }

// func (r *CatalogRanker) TopCoProcess(cps []*process.CoProcess) (*process.CoProcess, error) {
// 	if len(cps) > 0 {
// 		cps, err := r.RankCoProcess(cps)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return cps[0], nil
// 	}
// 	return nil, ErrNotFound
// }

// func (r *CatalogRanker) RankProcess(ps []*process.Process) ([]*process.Process, error) {
// 	sort.SliceStable(ps, func(i, j int) bool {
// 		rankI, errI := r.GetProcessRank(ps[i])
// 		rankJ, errJ := r.GetProcessRank(ps[j])

// 		if errI != nil || errJ != nil {
// 			return false
// 		}

// 		return catalog.CmpRank(*rankI, *rankJ) < 0
// 	})
// 	return ps, nil
// }

// func (r *CatalogRanker) TopProcess(ps []*process.Process) (*process.Process, error) {
// 	if len(ps) > 0 {
// 		ps, err := r.RankProcess(ps)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return ps[0], nil
// 	}
// 	return nil, ErrNotFound
// }
