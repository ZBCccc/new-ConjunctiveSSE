package HDXT

func CalculatesIdListSize(sIDList []string) int {
	size := 0
	for _, d := range sIDList {
		size += len(d)
	}
	return size
}
