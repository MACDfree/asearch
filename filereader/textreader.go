package filereader

import "io/ioutil"

func textRead(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func init() {
	Regist("txt", FileReaderFunc(textRead))
	Regist("go", FileReaderFunc(textRead))
	Regist("md", FileReaderFunc(textRead))
	Regist("java", FileReaderFunc(textRead))
}
