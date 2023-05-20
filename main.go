package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"time"
	"zap_test/model"
)

var logger *zap.Logger
var sugar *zap.SugaredLogger

func main() {
	//InitLogger()
	//for i := 0; i < 9; i++ {
	//	sugar.Debugf("查询用户信息开始 id:%d", 1)
	//	sugar.Infof("查询用户信息成功 name:%s age:%d", "zhangSan", 20)
	//	sugar.Errorf("查询用户信息失败 error:%v", "未该查询到该用户信息")
	//}
	////------------------------------------------------
	//var (
	//	client  = store.GetMgoCli()
	//	iResult *mongo.InsertManyResult
	//	err     error
	//	//id      primitive.ObjectID
	//)
	////2.选择数据库 my_db
	//collection := client.Database("topcloud").Collection("cart")
	////插入某一条数据
	//if iResult, err = collection.InsertMany(context.TODO(), model.ModuleList); err != nil {
	//	fmt.Print(err)
	//	return
	//}
	//iResult = iResult
	////------------------------------------------------
	//_id:默认生成一个全局唯一ID
	//for _, v := range iResult.InsertedIDs {
	//	id = v.(primitive.ObjectID)
	//	fmt.Println("自增ID", id.Hex())
	//}

	//result := model.GetAllModuleList()
	//model.GetModuleDetail("车蓬主体结构（双车位）")
	//model.PassTest()
	//model.InitData()
	//model.QueryIdParentPath("6465ede666f25181aa26d665")
	//fmt.Println("---------------------------------")
	result := model.GetAllModuleListV2()
	//var idList = []string{
	//	"6465ede666f25181aa26d654",
	//	"6465ede666f25181aa26d658",
	//	"6465ede666f25181aa26d65b",
	//	"6465ede666f25181aa26d65d",
	//}
	//fmt.Println(model.QueryIdParentPath(idList))
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "failed to read request body")
			return
		}
		defer r.Body.Close()

		var jsonData map[string]interface{}
		err = json.Unmarshal(body, &jsonData)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "invalid json data: %s", err.Error())
			return
		}

		encoded, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "failed to encode response json data")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes.NewBuffer(encoded).Bytes())
	})

	if err := http.ListenAndServe(":5555", nil); err != nil {
		panic(err)
	}
	//fmt.Println("---------------------------------")
}
func InitLogger() {
	writeSyncer := getLogWriter()
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	logger = zap.New(core, zap.AddCaller())
	defer logger.Sync()
	sugar = logger.Sugar()

}

// getEncoder
//
//	@Description:
//	@return zapcore.Encoder
func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = customTimeEncoder
	//encoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	//encoderConfig.EncodeCaller = zapcore.FullCallerEncoder  全路径 用处不大

	//这些字段都可以指定
	//TimeKey:        "time",
	//LevelKey:       "level",
	//NameKey:        "logger",
	//CallerKey:      "caller",
	//MessageKey:     "msg",
	//StacktraceKey:  "stacktrace",
	//LineEnding:     zapcore.DefaultLineEnding

	return zapcore.NewConsoleEncoder(encoderConfig)

	//return zapcore.NewConsoleEncoder(encoderConfig)
	//2023-05-15 14:53:00	[35mDEBUG[0m	zap_test/main.go:16	查询用户信息开始 id:1

	//return zapcore.NewJSONEncoder(encoderConfig)
	//{"level":"\u001b[31mERROR\u001b[0m","ts":"2023-05-15 14:48:15","caller":"zap_test/main.go:19","msg":"查询用户信息失败 error:未该查询到该用户信息"}
}
func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./test.log", // 日志文件位置
		MaxSize:    1,            // 进行切割之前，日志文件最大值(单位：MB)，默认100MB
		MaxBackups: 5,            // 保留旧文件的最大个数
		MaxAge:     1,            // 保留旧文件的最大天数
		Compress:   false,        // 是否压缩/归档旧文件
	}
	return zapcore.AddSync(lumberJackLogger)
}
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func Add(a int, b int) int {
	return a + b
}

func Mul(a int, b int) int {
	return a * b
}
