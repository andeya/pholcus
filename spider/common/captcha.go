package common

// import (
// 	"fmt"
// 	"github.com/henrylee2cn/captcha"
// 	"github.com/henrylee2cn/dbgutil"
// 	"image"
// 	"log"
// 	"strconv"
// )

// /*
// 1. 下载 100~200 左右验证码,主要是为了囊括 0~9 A~Z a~z
// 2. 训练->得到训练模型
// 3. 训练模型自动或者手动生成标准模型
// 4. 使用标准模型进行识别
// */

// // 通过
// const (
// 	Threshhole = 34000
// 	N          = 6
// )

// var trainModuleFile = `config/Train.dat`
// var stdModuleFile = `config/Std.dat`

// func trainloadsave() {
// 	//err:=captcha.ImageColorInfo(in,out string) ==> threshhole = 34000
// 	//err=captcha.DownCaptcha(dir string, n int) ==> get captcha..

// 	c := captcha.NewCaptcha(Threshhole, N)

// 	// 第一次训练
// 	capesOne := make([]image.Image, 0, 97)
// 	// 写入 样本....
// 	for i := 100; i < 197; i++ {
// 		img, err := captcha.ReadImg(`验证码\` + strconv.Itoa(i) + ".jpg")
// 		if err != nil {
// 			fmt.Println(err)
// 			continue
// 			//log.Fatal(err)
// 		}
// 		capesOne = append(capesOne, img)
// 	}

// 	trainModule, err := c.Train(capesOne, nil) // nil - newtrainModule
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	for char, binimgs := range trainModule {
// 		fmt.Println(string(char))
// 		for _, binimg := range binimgs {
// 			fmt.Println(binimg)
// 		}
// 	}
// 	dbgutil.Break()

// 	err = c.SaveTrainModule(trainModule, trainModuleFile) // 写入文件
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	/*
// 		// 第二次训练
// 		capesTwo := make([]image.Image, 100, 100)
// 		// 写入 样本....
// 		trainModule, err = c.Train(capesTwo, trainModuleFile) // nil - newtrainModule
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		err = c.SaveTrainModule(trainModule, trainModuleFile) // 写入文件
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	*/

// 	stdModule, err := c.AutoGenStdModuleFromMemory(trainModule)
// 	//or
// 	//stdModule, err := c.AutoGenStdModuleFromFile(trainModuleFile)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	for char, binimg := range stdModule {
// 		fmt.Println(string(char))
// 		fmt.Println(binimg)
// 	}
// 	dbgutil.Break()

// 	err = c.SaveStdModule(stdModule, stdModuleFile)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	c.ImportStdModule(stdModule)
// 	//or
// 	//stdModule, err := c.LoadStdModule(stdModuleFile)
// }

// func recognize() {
// 	c := captcha.NewCaptcha(Threshhole, N)

// 	//c.ImportStdModule(stdModule)
// 	//or
// 	//_, err := c.LoadStdModule(stdModuleFile)
// 	//stdModule, err := c.LoadStdModule(`d:\CCHelper\Golang\bin\Cleaned.dat`)
// 	stdModule, err := c.LoadStdModule(`Std.dat`)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	for _, char := range c.StdModuleCheck(stdModule, false) {
// 		fmt.Println(string(char))
// 	}

// 	for char, binimg := range stdModule {
// 		fmt.Println(string(char))
// 		fmt.Println(binimg)
// 	}
// 	dbgutil.Break()

// 	for i := 0; i < 100; i++ {
// 		img, err := captcha.ReadImg(`验证码\` + strconv.Itoa(i) + ".jpg")
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		fmt.Println(c.Recognize(img))
// 	}
// }

// func mains() {
// 	//trainloadsave()

// 	//recognize()

// 	c := captcha.NewCaptcha(Threshhole, N)
// 	train, err := c.LoadTrainModule(`Alphabet.dat`)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	std, err := c.AutoGenStdModuleFromMemory(train)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println("lack : ")
// 	for _, char := range c.StdModuleCheck(std, false) {
// 		fmt.Println(string(char))
// 	}

// 	c.ImportStdModule(std)

// 	for i := 0; i < 100; i++ {
// 		img, err := captcha.ReadImg(`验证码\` + strconv.Itoa(i) + ".jpg")
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		fmt.Println(c.Recognize(img))
// 	}
// }
