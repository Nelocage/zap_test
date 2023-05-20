package model

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"zap_test/store"
)

// Component 组件选配表
type Component struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name"`                //组件名字
	Price       float64                `bson:"price" json:"price"`              //组件基础价格
	ImageUrl    string                 `bson:"image_url" json:"image_url"`      //组件的图片
	CheckRules  map[string]interface{} `bson:"check_rules" json:"check_rules"`  //选择规则
	CreatedAt   int64                  `bson:"created_at" json:"-"`             //添加时间
	UpdatedAt   int64                  `bson:"updated_at" json:"-"`             //修改时间
	DeletedAt   int64                  `bson:"deleted_at" json:"-" `            //删除时间
	Description string                 `bson:"description" json:"description" ` //产品详情描述
	ParentID    *primitive.ObjectID    `bson:"parent_id,omitempty" json:"-"`    //父组件的id
}

type ComponentResponse struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name"`                   //组件名字
	Price       float64                `bson:"price" json:"price"`                 //组件基础价格
	Description string                 `bson:"description" json:"description"`     //产品详情描述
	ImageUrl    string                 `bson:"image_url" json:"image_url"`         //组件的图片
	CheckRules  map[string]interface{} `bson:"check_rules" json:"check_rules"`     //选择规则
	ParentID    *primitive.ObjectID    `bson:"parent_id,omitempty" json:"-"`       //父组件的id
	Children    []Component            `bson:"children,omitempty" json:"children"` //孩子
}

type FindByModuleName struct {
	Name string `bson:"name"`
}

func GetAllModuleList() []*Component {
	// 查询多个
	// 将选项传递给Find()
	findOptions := options.Find()
	//findOptions.SetLimit(2)

	client := store.GetMgoCli()
	var results []*Component
	collection := client.Database("topcloud").Collection("components")
	// 把bson.D{{}}作为一个filter来匹配所有文档
	cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		fmt.Println(err)
	}

	// 查找多个文档返回一个光标
	// 遍历游标允许我们一次解码一个文档
	for cur.Next(context.TODO()) {
		// 创建一个值，将单个文档解码为该值
		var elem Component
		err := cur.Decode(&elem)
		if err != nil {
			fmt.Println(err)
		}
		results = append(results, &elem)
		fmt.Println(elem.ID.Hex())
	}

	if err := cur.Err(); err != nil {
		fmt.Println(err)
	}
	// 完成后关闭游标
	cur.Close(context.TODO())

	var resultsResponse []ComponentResponse
	copier.Copy(&resultsResponse, results)

	fmt.Printf("Found multiple documents (array of pointers): %#v\n", results)
	return results

}

func GetModuleDetail(moduleName string) {
	filter := bson.D{{"name", moduleName}}
	var result Component

	client := store.GetMgoCli()
	collection := client.Database("topcloud").Collection("cart")
	err := collection.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Found a single document: %+v\n", result)
}

func reverse(s []string) {
	for i := 0; i < len(s)/2; i++ {
		j := len(s) - i - 1
		s[i], s[j] = s[j], s[i]
	}
}

func GetAllModuleListV2() []ComponentResponse {
	client := store.GetMgoCli()
	collection := client.Database("topcloud").Collection("components_new")
	// 构造聚合管道
	pipeline := bson.A{
		bson.M{"$match": bson.M{"parent_id": bson.M{"$exists": false}}}, // 查询根节点
		bson.M{
			"$graphLookup": bson.M{
				"from":             "components",
				"startWith":        "$_id",
				"connectFromField": "_id",
				"connectToField":   "parent_id",
				"as":               "children",
			},
		},
	}

	// 执行聚合操作，获取结果
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		panic(err)
	}
	defer cursor.Close(context.Background())
	var result []ComponentResponse
	if err = cursor.All(context.Background(), &result); err != nil {
		panic(err)
	}

	// 输出结果
	//for _, item := range result {
	//	fmt.Printf("ID: %d, Name: %s, Children: %v\n", item.ID, item.Name, item.Children)
	//}

	// 关闭连接
	err = client.Disconnect(context.Background())
	if err != nil {
		panic(err)
	}
	return result
}

// QueryIdParentPath 根据id 得到路径上所有节点的名字
func QueryIdParentPath(idList []string) [][]string {
	// 创建 MongoDB 客户端并连接到数据库
	var paths [][]string
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://172.22.114.78:27017"))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.Background())
	// 获取组件集合
	col := client.Database("topcloud").Collection("components")

	for _, item := range idList {
		var path []string
		// 构建查询条件
		id, err := primitive.ObjectIDFromHex(item)
		if err != nil {
			panic(err)
		}
		// 构建 $graphLookup 查询
		pipeline := bson.A{
			bson.M{"$match": bson.M{"_id": id}}, // 假设要查询的 ID 为 componentID5
			bson.M{
				"$graphLookup": bson.M{
					"from":             "components",
					"startWith":        "$_id",
					"connectFromField": "parent_id",
					"connectToField":   "_id",
					"as":               "path",
				},
			},
			bson.M{"$project": bson.M{"path.name": 1, "_id": 0}}, // 只保留路径上的名称信息
		}
		// 执行查询并输出结果
		cursor, err := col.Aggregate(context.Background(), pipeline)
		if err != nil {
			panic(err)
		}
		var result []bson.M
		if err = cursor.All(context.Background(), &result); err != nil {
			panic(err)
		}

		// 输出嵌套关系名称
		for _, item := range result[0]["path"].(primitive.A) {
			//fmt.Printf("%s ", item.(primitive.M)["name"])
			path = append(path, item.(primitive.M)["name"].(string))
		}
		reverse(path)
		paths = append(paths, path)
		cursor.Close(context.Background())
	}
	return paths
}

func InitData() {
	// 创建 MongoDB 客户端并连接到数据库
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://172.22.114.78:27017"))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.Background())

	// 获取组件集合
	col := client.Database("topcloud").Collection("components_new")
	// 插入测试数据
	rule1 := make(map[string]interface{})
	rule1["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component1 := Component{
		Name:        "车蓬主体结构（双车位）",
		Price:       72000,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		CheckRules:  rule1,
		Description: "含钢结构与铝型材",
		ParentID:    nil,
	}
	res, err := col.InsertOne(context.Background(), component1)
	if err != nil {
		panic(err)
	}
	//componentID1 := res.InsertedID.(primitive.ObjectID)

	rule2 := make(map[string]interface{})
	rule2["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component2 := Component{
		Name:        "车蓬主体结构（单车位）",
		Price:       72000,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		CheckRules:  rule2,
		Description: "含钢结构与铝型材",
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component2)
	if err != nil {
		panic(err)
	}
	//componentID2 := res.InsertedID.(primitive.ObjectID)

	rule3 := make(map[string]interface{})
	rule3["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component3 := Component{
		Name:        "交流充电桩（单枪）",
		Price:       0,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		CheckRules:  rule3,
		Description: "",
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component3)
	if err != nil {
		panic(err)
	}
	componentID3 := res.InsertedID.(primitive.ObjectID)

	component4 := Component{
		Name:        "7KW",
		Price:       1350,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		Description: "",
		ParentID:    &componentID3,
	}
	res, err = col.InsertOne(context.Background(), component4)
	if err != nil {
		panic(err)
	}

	rule5 := make(map[string]interface{})
	rule5["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component5 := Component{
		Name:        "交流充电桩（双枪）",
		Price:       0,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		CheckRules:  rule5,
		Description: "",
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component5)
	componentID5 := res.InsertedID.(primitive.ObjectID)
	if err != nil {
		panic(err)
	}

	component6 := Component{
		Name:        "7KW",
		Price:       2350,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		Description: "",
		ParentID:    &componentID5,
	}
	res, err = col.InsertOne(context.Background(), component6)

	if err != nil {
		panic(err)
	}

	rule7 := make(map[string]interface{})
	rule7["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component7 := Component{
		Name:        "直流充电桩（单枪）",
		Price:       0,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		CheckRules:  rule7,
		Description: "",
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component7)
	componentID7 := res.InsertedID.(primitive.ObjectID)
	if err != nil {
		panic(err)
	}

	component8 := Component{
		Name:        "60kW",
		Price:       23360,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		Description: "",
		ImageUrl:    "",
		ParentID:    &componentID7,
	}
	res, err = col.InsertOne(context.Background(), component8)

	if err != nil {
		panic(err)
	}

	rule9 := make(map[string]interface{})
	rule9["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component9 := Component{
		Name:        "直流充电桩（双枪）",
		Price:       0,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		Description: "",
		ImageUrl:    "",
		CheckRules:  rule9,
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component9)
	componentID9 := res.InsertedID.(primitive.ObjectID)
	if err != nil {
		panic(err)
	}

	component10 := Component{
		Name:        "120kW",
		Price:       36800,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		Description: "",
		ImageUrl:    "",

		ParentID: &componentID9,
	}
	res, err = col.InsertOne(context.Background(), component10)

	if err != nil {
		panic(err)
	}

	rule11 := make(map[string]interface{})
	rule11["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component11 := Component{
		Name:        "光伏背板",
		Price:       20000,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		Description: "",
		ImageUrl:    "",
		CheckRules:  rule11,
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component11)

	if err != nil {
		panic(err)
	}

	rule12 := make(map[string]interface{})
	rule12["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component12 := Component{
		Name:        "储能电柜（小）",
		Price:       0,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		Description: "",
		ImageUrl:    "",
		CheckRules:  rule12,
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component12)
	componentID12 := res.InsertedID.(primitive.ObjectID)
	if err != nil {
		panic(err)
	}

	component13 := Component{
		Name:        "8KW 20度电",
		Price:       86000,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		Description: "",
		ImageUrl:    "",
		ParentID:    &componentID12,
	}
	res, err = col.InsertOne(context.Background(), component13)

	if err != nil {
		panic(err)
	}

	rule14 := make(map[string]interface{})
	rule14["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component14 := Component{
		Name:        "气象站",
		Price:       13600,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		Description: "",
		ImageUrl:    "",
		CheckRules:  rule14,
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component14)

	if err != nil {
		panic(err)
	}

	rule15 := make(map[string]interface{})
	rule15["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component15 := Component{
		Name:        "摄像头",
		Price:       0,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		CheckRules:  rule15,
		Description: "",
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component15)
	componentID15 := res.InsertedID.(primitive.ObjectID)
	if err != nil {
		panic(err)
	}

	component16 := Component{
		Name:        "2K分辨率",
		Price:       800,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		Description: "",
		ParentID:    &componentID15,
	}
	res, err = col.InsertOne(context.Background(), component16)
	if err != nil {
		panic(err)
	}

	rule17 := make(map[string]interface{})
	rule17["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component17 := Component{
		Name:        "LED显示屏（小）",
		Price:       2350,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		CheckRules:  rule17,
		Description: "",
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component17)
	componentID17 := res.InsertedID.(primitive.ObjectID)
	if err != nil {
		panic(err)
	}

	component18 := Component{
		Name:        "P10",
		Price:       1000,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		Description: "",
		ParentID:    &componentID17,
	}
	res, err = col.InsertOne(context.Background(), component18)
	if err != nil {
		panic(err)
	}

	rule19 := make(map[string]interface{})
	rule19["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component19 := Component{
		Name:        "雨水回收系统",
		Price:       0,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		CheckRules:  rule19,
		Description: "",
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component19)
	componentID19 := res.InsertedID.(primitive.ObjectID)
	if err != nil {
		panic(err)
	}

	component20 := Component{
		Name:      "水箱、台盆、水位计",
		Price:     5000,
		CreatedAt: 1684399050000,
		UpdatedAt: 1684399050000,
		DeletedAt: 0,
		ImageUrl:  "",

		Description: "",
		ParentID:    &componentID19,
	}
	res, err = col.InsertOne(context.Background(), component20)
	if err != nil {
		panic(err)
	}

	rule21 := make(map[string]interface{})
	rule21["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component21 := Component{
		Name:        "安全屋",
		Price:       43000,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		CheckRules:  rule21,
		Description: "",
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component21)
	if err != nil {
		panic(err)
	}

	rule22 := make(map[string]interface{})
	rule22["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component22 := Component{
		Name:        "热站",
		Price:       0,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		CheckRules:  rule22,
		Description: "",
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component22)
	componentID22 := res.InsertedID.(primitive.ObjectID)
	if err != nil {
		panic(err)
	}

	component23 := Component{
		Name:        "双源热泵",
		Price:       64000,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		Description: "",
		ImageUrl:    "",

		ParentID: &componentID22,
	}
	res, err = col.InsertOne(context.Background(), component23)
	if err != nil {
		panic(err)
	}

	rule24 := make(map[string]interface{})
	rule24["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component24 := Component{
		Name:        "智慧灯杆",
		Price:       150000,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		ImageUrl:    "",
		CheckRules:  rule24,
		Description: "",
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component24)
	if err != nil {
		panic(err)
	}

	rule25 := make(map[string]interface{})
	rule25["radio_or_checkbox"] = "single" //该字段只能赋值为single multi
	component25 := Component{
		Name:        "母从插座",
		Price:       0,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		Description: "",
		ImageUrl:    "",
		CheckRules:  rule25,
		ParentID:    nil,
	}
	res, err = col.InsertOne(context.Background(), component25)
	componentID25 := res.InsertedID.(primitive.ObjectID)
	if err != nil {
		panic(err)
	}

	component26 := Component{
		Name:        "防水插座",
		Price:       300,
		CreatedAt:   1684399050000,
		UpdatedAt:   1684399050000,
		DeletedAt:   0,
		Description: "",
		ImageUrl:    "",
		ParentID:    &componentID25,
	}
	res, err = col.InsertOne(context.Background(), component26)
	if err != nil {
		panic(err)
	}
}
