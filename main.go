package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sort"

	"strconv"
	"time"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
)

const uri = "mongodb+srv://root:0000@cluster0.kncismv.mongodb.net/?retryWrites=true&w=majority"

type User struct {
	UserName     string `json:"username" binding:"required"`
	Name         string `json:"name"`
	Surname      string `json:"surname"`
	Phone        string `json:"phone"`
	Country      string `json:"country"`
	Password     string `json:"password"`
	RegisterDate string `json:"register_date"`
	Blocked      bool   `json:"blocked"`
	UserId       string `json:"user_id"`
	UserStatus   string `json:"user_status"`
	UserRole     string `json:"user_role"`
}

type ProductCategory struct {
	CategoryName string `json:"category_name"`
	CategoryId   string `json:"category_id"`
	CategoryIcon string `json:"category_icon"`
}

type Product struct {
	ProductId      string `json:"product_id"`
	ProductName    string `json:"product_name"`
	ProductDesc    string `json:"product_desc"`
	ProductCatId   string `json:"product_cat_id"`
	ProductPrice   int64  `json:"product_price"`
	ProductBenefit int64  `json:"product_benefit"`
	ProductStock   string `json:"product_stock"`
	ProductStatus  string `json:"product_status"`
	ProductDate    string `json:"product_date"`
	ProductSeller  string `json:"product_seller"`
	ProductNumber  int64  `json:"product_number"`
}

type Transaction struct {
	TransactionId          string `json:"transaction_id"`
	TransactionDate        string `json:"transaction_date"`
	TransactionSeller      string `json:"transaction_seller"`
	TransactionProductName string `json:"transaction_product_name"`
	TransactionProduct     string `json:"transaction_product"`
	TransactionNumber      int64  `json:"transaction_number"`
	TransactionPrice       int64  `json:"transaction_price"`
	TransactionStatus      string `json:"transaction_status"`
	TransactionBenefit     int64  `json:"transaction_benefit"`
}

func passwordHash(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		fmt.Println(err)
	}
	return string(hash)
}

func generateUserId() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	length := 32
	b := make([]rune, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	router := gin.Default()
	router.POST("/register", register)
	router.POST("/login", login)
	router.GET("/getAllUser", getAllUser)
	router.GET("/getUser", getUser)
	router.PUT("/updatePassword", updatePassword)
	router.PUT("/updateBlocked", updateBlocked)
	router.PUT("/updateUserRole", updateUserRole)
	router.PUT("/updateUser", updateUser)
	router.POST("/addCategory", addCategory)
	router.GET("/getAllCategory", getAllCategory)
	router.POST("/addProduct", addProduct)
	router.GET("/getAllProduct", getAllProduct)
	router.GET("/getProductsByCategory", getProductsByCategory)
	router.GET("/getProduct", getProduct)
	router.PUT("/updateProduct", updateProduct)
	router.DELETE("/deleteProduct", deleteProduct)
	router.DELETE("/deleteCategory", deleteCategory)
	router.POST("/productSell", productSell)
	router.POST("/addProductSell", addProductSell)
	router.GET("/getUserProductSell", getUserProductSell)
	router.GET("/getProductSell", getProductSell)
	router.GET("/getAllSell", getAllSell)
	router.GET("/getSellTransaction", getSellTransaction)
	router.DELETE("/deleteUser", deleteUser)

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	router.Use(cors.New(config))
	router.Run()

	
	
}

func register(c *gin.Context) {
	var user User
	c.BindJSON(&user)
	user.Password = passwordHash(user.Password)
	user.RegisterDate = time.Now().Format("2006-01-02 15:04:05")
	user.Blocked = false
	user.UserId = generateUserId()
	user.UserStatus = "active"
	user.UserRole = "user"
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	var result User
	collection := client.Database("DataBase").Collection("users")
	err = collection.FindOne(ctx, bson.M{"username": user.UserName}).Decode(&result)
	if err != nil {
		fmt.Println(err)
	}
	if result.UserName == user.UserName {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "User already exists"})
		return
	}
	if user.UserName == "" || user.Password == "" || user.Name == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Please fill all fields"})
		return
	}
	if len(user.Password) < 6 {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Password must be at least 6 characters"})
		return
	}

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		fmt.Println(err)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User created"})
}

func login(c *gin.Context) {
	var user User
	c.BindJSON(&user)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	var result User
	collection := client.Database("DataBase").Collection("users")
	err = collection.FindOne(ctx, bson.M{"username": user.UserName}).Decode(&result)
	if err != nil {
		fmt.Println(err)
	}
	if result.UserName != user.UserName {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "User not found"})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Wrong password"})
		return
	}
	if result.Blocked {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "User blocked"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"username": result.UserName, "name": result.Name, "surname": result.Surname, "role": result.UserRole, "phone": result.Phone, "blocked": result.Blocked, "userid": result.UserId, "userstatus": result.UserStatus, "registerdate": result.RegisterDate})
}

func getAllUser(c *gin.Context) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	var result []User
	collection := client.Database("DataBase").Collection("users")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	for cursor.Next(ctx) {
		var user User
		cursor.Decode(&user)
		result = append(result, user)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": result})
}

func getUser(c *gin.Context) {
	var user User
	c.BindJSON(&user)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	var result User
	collection := client.Database("DataBase").Collection("users")
	err = collection.FindOne(ctx, bson.M{"username": user.UserName}).Decode(&result)
	if err != nil {
		fmt.Println(err)
	}
	if result.UserName != user.UserName {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": result})
}

func updatePassword(c *gin.Context) {
	var user User
	c.BindJSON(&user)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	var result User
	collection := client.Database("DataBase").Collection("users")
	err = collection.FindOne(ctx, bson.M{"username": user.UserName}).Decode(&result)
	if err != nil {
		fmt.Println(err)
	}
	if result.UserName != user.UserName {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "User not found"})
		return
	}
	hash := passwordHash(user.Password)
	_, err = collection.UpdateOne(ctx, bson.M{"username": user.UserName}, bson.M{"$set": bson.M{"password": hash}})
	if err != nil {
		fmt.Println(err)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Password updated"})
}

func updateBlocked(c *gin.Context) {
	var user User
	c.BindJSON(&user)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	var result User
	collection := client.Database("DataBase").Collection("users")
	err = collection.FindOne(ctx, bson.M{"username": user.UserName}).Decode(&result)
	if err != nil {
		fmt.Println(err)
	}
	if result.UserName != user.UserName {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "User not found"})
		return
	}
	_, err = collection.UpdateOne(ctx, bson.M{"username": user.UserName}, bson.M{"$set": bson.M{"blocked": user.Blocked}})
	if err != nil {
		fmt.Println(err)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Blocked updated"})
}

func updateUserRole(c *gin.Context) {
	var user User
	c.BindJSON(&user)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	var result User
	collection := client.Database("DataBase").Collection("users")
	err = collection.FindOne(ctx, bson.M{"username": user.UserName}).Decode(&result)
	if err != nil {
		fmt.Println(err)
	}
	if result.UserName != user.UserName {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "User not found"})
		return
	}
	if result.UserRole == "creator" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "You can't change role of creator"})
		return
	}
	_, err = collection.UpdateOne(ctx, bson.M{"username": user.UserName}, bson.M{"$set": bson.M{"userrole": user.UserRole}})
	if err != nil {
		fmt.Println(err)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Role updated"})
}

func addCategory(c *gin.Context) {
	var category ProductCategory
	c.BindJSON(&category)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("categories")
	category.CategoryId = generateUserId()
	if category.CategoryName == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Category name can't be empty"})
		return
	}
	if category.CategoryName == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Category name can't be empty"})
		return
	}
	_, err = collection.InsertOne(ctx, category)
	if err != nil {
		fmt.Println(err)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Category added"})
}

func getAllCategory(c *gin.Context) {
	var categories []ProductCategory
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("categories")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	for cursor.Next(ctx) {
		var category ProductCategory
		cursor.Decode(&category)
		categories = append(categories, category)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": categories})
}

func getAllProduct(c *gin.Context) {
	var products []Product
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("products")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	for cursor.Next(ctx) {
		var product Product
		cursor.Decode(&product)
		products = append(products, product)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": products})
}

func getProductsByCategory(c *gin.Context) {
	var products []Product
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("products")
	cursor, err := collection.Find(ctx, bson.M{"productcatid": c.Query("categoryId")})
	if err != nil {
		fmt.Println(err)
	}
	for cursor.Next(ctx) {
		var product Product
		cursor.Decode(&product)
		products = append(products, product)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": products})
}

func getProduct(c *gin.Context) {
	var product Product
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("products")
	err = collection.FindOne(ctx, bson.M{"productid": c.Query("productId")}).Decode(&product)
	if err != nil {
		fmt.Println(err)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": product})
}

func updateProduct(c *gin.Context) {
	var product Product
	c.BindJSON(&product)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("products")

	if product.ProductName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Product name is required"})
		return
	}
	if product.ProductPrice < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Product price is required"})
		return
	}
	if product.ProductCatId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Product category is required"})
		return
	}
	if product.ProductBenefit < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Product benefit is required"})
		return
	}
	if product.ProductDesc == "" {
		product.ProductDesc = "Izohlar yo'q"
	}

	_, err = collection.UpdateOne(ctx,
		bson.M{"productid": c.Query("productId")},
		bson.M{"$set": bson.M{
			"productname":    product.ProductName,
			"productprice":   product.ProductPrice,
			"productcatid":   product.ProductCatId,
			"productbenefit": product.ProductBenefit,
			"productdesc":    product.ProductDesc,
			"ProductBenefit": product.ProductBenefit}})
	if err != nil {
		fmt.Println(err)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Product updated"})
}

func deleteProduct(c *gin.Context) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("products")
	_, err = collection.DeleteOne(ctx, bson.M{"productid": c.Query("productId")})
	if err != nil {
		fmt.Println(err)
	}
	if c.Query("productId") == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Product not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Product deleted"})
}

func deleteCategory(c *gin.Context) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("categories")
	_, err = collection.DeleteOne(ctx, bson.M{"categoryid": c.Query("categoryId")})
	if err != nil {
		fmt.Println(err)
	}
	if c.Query("categoryId") == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Category not found"})
		return
	}
	// delete products
	collection = client.Database("DataBase").Collection("products")
	var products []Product
	cursor, err := collection.Find(ctx, bson.M{"productcatid": c.Query("categoryId")})
	if err != nil {
		fmt.Println(err)
	}
	for cursor.Next(ctx) {
		var product Product
		cursor.Decode(&product)
		products = append(products, product)
	}
	for _, product := range products {
		_, err = collection.DeleteOne(ctx, bson.M{"productid": product.ProductId})
		if err != nil {
			fmt.Println(err)
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Category deleted"})
}

func productSell(c *gin.Context) {
	var product Product
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("products")

	filter := bson.M{"productid": c.Query("productId")}
	number, err := strconv.Atoi(c.Query("number"))
	if err != nil {
		fmt.Println(err)
	}
	err = collection.FindOne(ctx, filter).Decode(&product)
	if err != nil {
		fmt.Println(err)
	}

	if int(product.ProductNumber) < number {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Not enough products"})
		return
	}
	if number < 1 {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Number must be greater than 0"})
		return
	}
	if c.Query("userId") == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "User not found"})
		return
	}
	if c.Query("productId") == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Product not found"})
		return
	}

	addition := int64(0)
	addition, err = strconv.ParseInt(c.PostForm("addition_price"), 10, 64)
	if err != nil {
		addition = 0
	}
	if product.ProductNumber < -1 {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Not enough products"})
		return
	}

	product.ProductNumber = product.ProductNumber - int64(number)
	_, err = collection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"productnumber": product.ProductNumber}})
	if err != nil {
		fmt.Println(err)
	}
	var transaction Transaction
	transaction.TransactionId = generateUserId()
	transaction.TransactionProductName = product.ProductName
	transaction.TransactionProduct = product.ProductId
	transaction.TransactionNumber = int64(number)
	transaction.TransactionPrice = product.ProductPrice
	transaction.TransactionBenefit = addition + product.ProductBenefit
	transaction.TransactionDate = time.Now().Format("2006-01-02 15:04:05")
	transaction.TransactionSeller = c.Query("userId")
	transaction.TransactionStatus = "sold"
	collection = client.Database("DataBase").Collection("transactions")
	_, err = collection.InsertOne(ctx, transaction)
	if err != nil {
		fmt.Println(err)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Product sold"})
}

func addProductSell(c *gin.Context) {
	var product Product
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("products")

	filter := bson.M{"productid": c.Query("productId")}
	number, err := strconv.Atoi(c.Query("number"))
	if err != nil {
		fmt.Println(err)
	}
	if number < 1 {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Number must be greater than 0"})
		return
	}

	if c.Query("productId") == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Product not found"})
		return
	}

	if c.Query("userId") == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "User not found"})
		return
	}

	var transaction Transaction
	transaction.TransactionId = generateUserId()
	transaction.TransactionProductName = product.ProductName
	transaction.TransactionNumber = int64(number)
	transaction.TransactionStatus = "added"
	transaction.TransactionProduct = c.Query("productId")
	transaction.TransactionPrice, err = strconv.ParseInt(c.PostForm("transaction_price"), 10, 64)
	if err != nil {
		transaction.TransactionPrice = 0
	}
	transaction.TransactionBenefit, err = strconv.ParseInt(c.PostForm("transaction_benefit"), 10, 64)
	if err != nil {
		transaction.TransactionBenefit = 0
	}
	if transaction.TransactionBenefit < 0 {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Benefit must be greater than 0"})
		return
	}
	transaction.TransactionProductName = c.PostForm("transaction_product_name")

	transaction.TransactionDate = time.Now().Format("2006-01-02 15:04:05")
	transaction.TransactionSeller = c.Query("userId")
	err = collection.FindOne(ctx, filter).Decode(&product)
	if err != nil {
		fmt.Println(err)
	}

	if transaction.TransactionPrice < 1 {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Price must be greater than 0"})
		return
	}

	product.ProductNumber = product.ProductNumber + int64(number)
	_, err = collection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"productnumber": product.ProductNumber}})
	if err != nil {
		fmt.Println(err)
	}
	collection = client.Database("DataBase").Collection("transactions")
	_, err = collection.InsertOne(ctx, transaction)
	if err != nil {
		fmt.Println(err)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Product added"})
}

func addProduct(c *gin.Context) {
	var product Product
	c.BindJSON(&product)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("products")
	product.ProductId = generateUserId()

	if product.ProductName == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Product name can't be empty"})
		return
	}
	if product.ProductPrice < 0 {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Product price can't be empty"})
		return
	}
	if product.ProductCatId == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Product category id empty"})
		return
	}
	if product.ProductSeller == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Product seller can't be empty"})
		return
	}
	if product.ProductNumber < 0 {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Product number can't be empty"})
		return
	}
	if product.ProductBenefit < 0 {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Product benefit can't be empty"})
		return
	}
	if product.ProductDesc == "" {
		product.ProductDesc = "Izohlar yo'q"
	}

	product.ProductId = generateUserId()
	product.ProductDate = time.Now().Format("2006-01-02 15:04:05")

	_, err = collection.InsertOne(ctx, product)
	if err != nil {
		fmt.Println(err)
	}
	collection = client.Database("DataBase").Collection("transactions")
	var transaction Transaction
	transaction.TransactionId = generateUserId()
	transaction.TransactionProduct = product.ProductId
	transaction.TransactionNumber = product.ProductNumber
	transaction.TransactionProductName = product.ProductName
	transaction.TransactionStatus = "added"
	transaction.TransactionProduct = product.ProductId
	transaction.TransactionPrice = product.ProductPrice
	transaction.TransactionBenefit = 0
	transaction.TransactionDate = time.Now().Format("2006-01-02 15:04:05")
	transaction.TransactionSeller = product.ProductSeller
	_, err = collection.InsertOne(ctx, transaction)
	if err != nil {
		fmt.Println(err)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Product added"})
}

func getAllSell(c *gin.Context) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("transactions")
	var transactions []Transaction
	cursor, err := collection.Find(ctx, bson.M{"transactionstatus": c.Query("status")})
	if err != nil {
		fmt.Println(err)
	}
	months := c.Query("months")
	monthsInt, err := strconv.Atoi(months)
	if err != nil {
		fmt.Println(err)
	}
	if monthsInt == 3 {
		monthsInt = 2184
	}
	if monthsInt == 2 {
		monthsInt = 1464
	}
	if monthsInt == 1 {
		monthsInt = 744
	}
	if monthsInt == 0 {
		c.JSON(http.StatusOK, gin.H{"status": "success", "error": "months must be greater than 0"})
		return
	}
	if monthsInt == 7 {
		monthsInt = 192
	}
	price := 0
	benefit := 0

	for cursor.Next(ctx) {
		var transaction Transaction
		cursor.Decode(&transaction)
		transactionDate, err := time.Parse("2006-01-02 15:04:05", transaction.TransactionDate)
		if err != nil {
			fmt.Println(err)
		}
		if time.Now().Sub(transactionDate).Hours() < float64(monthsInt) {
			transactions = append(transactions, transaction)
			price = price + int(transaction.TransactionPrice)
			benefit = benefit + int(transaction.TransactionBenefit)
		} else {
			if time.Now().Sub(transactionDate).Hours() > 2184 {
				_, err = collection.DeleteOne(ctx, bson.M{"transactionid": transaction.TransactionId})
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].TransactionDate > transactions[j].TransactionDate
	})
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": transactions, "price": price, "benefit": benefit})
	price = 0
	benefit = 0
}

func getProductByName(productId string) string {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("products")
	//filter product by id and return product name
	var product Product
	err = collection.FindOne(ctx, bson.M{"productid": productId}).Decode(&product)
	if err != nil {
		fmt.Println(err)
	}
	return product.ProductName
}

func getSellTransaction(c *gin.Context) {
	//get all transactions from database sorted by date and time and return it sold and added
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("transactions")
	var transactions []Transaction
	//get all transactions from database
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	months := c.Query("months")
	sells := c.Query("sells")
	monthsInt, err := strconv.Atoi(months)
	if err != nil {
		fmt.Println(err)
	}
	if monthsInt == 3 {
		monthsInt = 2184
	}
	if monthsInt == 2 {
		monthsInt = 1464
	}
	if monthsInt == 1 {
		monthsInt = 744
	}
	if monthsInt == 0 {
		monthsInt = 24
	}
	if monthsInt < -1 {
		c.JSON(http.StatusOK, gin.H{"status": "success", "error": "months must be greater than 0"})
		return
	}
	price := 0
	benefit := 0

	for cursor.Next(ctx) {
		var transaction Transaction
		cursor.Decode(&transaction)
		transactionDate, err := time.Parse("2006-01-02 15:04:05", transaction.TransactionDate)
		if err != nil {
			fmt.Println(err)
		}
		if time.Now().Sub(transactionDate).Hours() < float64(monthsInt) {
			if sells == "sold"{
				if transaction.TransactionStatus == "sold" {
				transactions = append(transactions, transaction)
				price = price + int(transaction.TransactionPrice)
				benefit = benefit + int(transaction.TransactionBenefit)
				}
			}
			if sells == "added" {
				if transaction.TransactionStatus == "added" {
					transactions = append(transactions, transaction)
					price = price + int(transaction.TransactionPrice)
					benefit = benefit + int(transaction.TransactionBenefit)
				}
			}
			if sells == "all"||sells==""{
				transactions = append(transactions, transaction)
				price = price + int(transaction.TransactionPrice)
				benefit = benefit + int(transaction.TransactionBenefit)
			}
		} else {
			if time.Now().Sub(transactionDate).Hours() > 2184 {
				_, err = collection.DeleteOne(ctx, bson.M{"transactionid": transaction.TransactionId})
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].TransactionDate > transactions[j].TransactionDate
	})
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": transactions, "price": price, "benefit": benefit})
	price = 0
	benefit = 0
}

func getProductSell(c *gin.Context) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("transactions")
	var transactions []Transaction
	cursor, err := collection.Find(ctx, bson.M{
		"transactionproduct": c.Query("productId"),})
	if err != nil {
		fmt.Println(err)
	}
	months := c.Query("months")
	sells := c.Query("sells")
	monthsInt, err := strconv.Atoi(months)
	if err != nil {
		fmt.Println(err)
	}
	if monthsInt == 3 {
		monthsInt = 2184
	}
	if monthsInt == 2 {
		monthsInt = 1464
	}
	if monthsInt == 1 {
		monthsInt = 744
	}
	if monthsInt == 0 {
		monthsInt = 24
	}
	if monthsInt < -1 {
		c.JSON(http.StatusOK, gin.H{"status": "success", "error": "months must be greater than 0"})
		return
	}

	price := 0
	benefit := 0

	for cursor.Next(ctx) {
		var transaction Transaction
		cursor.Decode(&transaction)
		transactionDate, err := time.Parse("2006-01-02 15:04:05", transaction.TransactionDate)
		if err != nil {
			fmt.Println(err)
		}
		if time.Now().Sub(transactionDate).Hours() < float64(monthsInt) {
			if sells == "sold"{
				if transaction.TransactionStatus == "sold" {
					transactions = append(transactions, transaction)
					price = price + int(transaction.TransactionPrice)
					benefit = benefit + int(transaction.TransactionBenefit)
				}
			}
			if sells == "added" {
				if transaction.TransactionStatus == "added" {
					transactions = append(transactions, transaction)
					price = price + int(transaction.TransactionPrice)
					benefit = benefit + int(transaction.TransactionBenefit)
				}
			}
			if sells == "all"||sells==""{
				transactions = append(transactions, transaction)
				price = price + int(transaction.TransactionPrice)
				benefit = benefit + int(transaction.TransactionBenefit)
			}
		} else {
			if time.Now().Sub(transactionDate).Hours() > 2184 {
				_, err = collection.DeleteOne(ctx, bson.M{"transactionid": transaction.TransactionId})
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].TransactionDate > transactions[j].TransactionDate
	})
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": transactions, "price": price, "benefit": benefit})
	price = 0
	benefit = 0
}

func getUserProductSell(c *gin.Context) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("transactions")
	var transactions []Transaction
	cursor, err := collection.Find(ctx, bson.M{"transactionseller": c.Query("userId")})
	if err != nil {
		fmt.Println(err)
	}
	months := c.Query("months")
	sells := c.Query("sells")
	monthsInt, err := strconv.Atoi(months)
	if err != nil {
		fmt.Println(err)
	}
	if monthsInt == 3 {
		monthsInt = 2184
	}
	if monthsInt == 2 {
		monthsInt = 1464
	}
	if monthsInt == 1 {
		monthsInt = 744
	}
	if monthsInt == 0 {
		monthsInt = 24
	}
	if monthsInt < -1 {
		c.JSON(http.StatusOK, gin.H{"status": "success", "error": "months must be greater than 0"})
		return
	}
	price := 0
	benefit := 0

	for cursor.Next(ctx) {
		var transaction Transaction
		cursor.Decode(&transaction)
		transactionDate, err := time.Parse("2006-01-02 15:04:05", transaction.TransactionDate)
		if err != nil {
			fmt.Println(err)
		}
		if time.Now().Sub(transactionDate).Hours() < float64(monthsInt) {
			if sells == "sold"{
				if transaction.TransactionStatus == "sold" {
					transactions = append(transactions, transaction)
					price = price + int(transaction.TransactionPrice)
					benefit = benefit + int(transaction.TransactionBenefit)
				}
			}
			if sells == "added" {
				if transaction.TransactionStatus == "added" {
					transactions = append(transactions, transaction)
					price = price + int(transaction.TransactionPrice)
					benefit = benefit + int(transaction.TransactionBenefit)
				}
			}
			if sells == "all"||sells==""{
				transactions = append(transactions, transaction)
				price = price + int(transaction.TransactionPrice)
				benefit = benefit + int(transaction.TransactionBenefit)
			}
		} else {
			if time.Now().Sub(transactionDate).Hours() > 2184 {
				_, err = collection.DeleteOne(ctx, bson.M{"transactionid": transaction.TransactionId})
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].TransactionDate > transactions[j].TransactionDate
	})
	c.JSON(http.StatusOK, gin.H{"status": "success", "transactions": transactions, "price": price, "benefit": benefit})
	price = 0
	benefit = 0
}

func deleteUser(c *gin.Context) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("users")
	_, err = collection.DeleteOne(ctx, bson.M{"userid": c.Query("userid")})
	if err != nil {
		fmt.Println(err)
	}
	if c.Query("userid") == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User deleted"})
}

func updateUser(c *gin.Context) {
	var user User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		fmt.Println(err)
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("DataBase").Collection("users")
	var result User
	err = collection.FindOne(ctx, bson.M{"userid": c.Query("userId")}).Decode(&result)
	if err != nil {
		fmt.Println(err)
	}
	if c.Query("userId") == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "User not found"})
		return
	}
	if user.UserName == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Username cannot be empty"})
		return
	}
	if user.Name == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Name cannot be empty"})
		return
	}
	if user.Surname == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Surname cannot be empty"})
		return
	}
	if user.Phone == "" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Phone number cannot be empty"})
		return
	}
	_, err = collection.UpdateOne(ctx, bson.M{"userid": c.Query("userId")}, bson.M{"$set": bson.M{
		"username": user.UserName,
		"name":     user.Name,
		"surname":  user.Surname,
		"phone":    user.Phone,
	}})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User updated"})

}
