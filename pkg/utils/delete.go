package utils

type deletePair struct {
	id       string
	keywords []string
}

func GenDeletePairs(filePath string, num int) []deletePair {
	// 1.读取.txt文件，对id对应的counter进行累加

	// 2.边累加边记录id，知道累加和达到num

	// 3.得到id，从mongodb数据库中读取id和对应的keywords，保存在deletePair中

	// 4.返回deletePair

	return nil
}
