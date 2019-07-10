package controller

import (
	"cloud/db"
	"fmt"
	"net/http"

	valid "github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
[]*db.Product 数据结构(model 中的) Product 组成的一个slice，
Products: 一个名字，类似NAME，先判请求中是否含有名为“products”的key，
然后查看该key是否对应了一个array类型的数据格式(因为 []*db.Product 是array)
再检查model中 Product struct（db.Product） 的各个参数的datatype是否符合，如果全部符合才可以bind
*/
type cartRequest struct {
	Products []*db.Product `json:"products"`
}

//CreateCart func
func CreateCart(c echo.Context) error {

	if !Authorize(c) { //Authorize,用户认证，检测用户是否注册
		return c.JSON(http.StatusUnauthorized, "forbidden")
	}

	req := &cartRequest{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	//map[KeyDatatype]valueDatatype, 用make定义一个map
	productMap := make(map[primitive.ObjectID]int32)
	//init a slice of ObjectID called ids
	var ids []primitive.ObjectID
	//use for loop to iterate []*db.Product in cartRequest
	//p represents each element in the product array([]*db.Product)
	//_ represents index of each element, no need to be used in this time
	//for := range 固定写法
	for _, p := range req.Products {
		//add new element into map(productMap)
		productMap[p.ID] = p.Quantity
		//add new element(objectID of each product) to the end of the ids(objectID array)
		ids = append(ids, p.ID)
	}

	products, err := db.ReadProducts(ids)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	var totalPrice float64
	for _, p := range products {
		p.Desc = ""
		p.Quantity = productMap[p.ID]
		totalPrice += p.Price * float64(productMap[p.ID])
	}
	id := c.Request().Header.Get("id")
	cart := &db.Cart{}
	cart.CustomerID, _ = primitive.ObjectIDFromHex(id)
	cart.ID = primitive.NewObjectID()
	cart.Products = products
	fs := fmt.Sprintf("%.2f", totalPrice)
	cart.TotalPrice, _ = valid.ToFloat(fs)
	if err := db.AddCart(cart); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, cart.ID)
}
