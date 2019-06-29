package main
import (
	   "net/http"
	   "github.com/gin-gonic/gin"
	   "fmt"
	   "os"
	   "database/sql"
	   _ "github.com/lib/pq"
	   "log"
	   "strconv"  
)
func main(){

	createTable()
	r := setupRouter()
	r.Run(":2019")
}

type Customer struct {
	ID   int     `json:"id"`
	Name string  `json:"name"`	
	Email string  `json:"email"`
	Status string `json:"status"`
}
func createTable() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("fatal", err.Error())
	}
	defer db.Close()

	createTb := `
	CREATE TABLE IF NOT EXISTS customers(
		id SERIAL PRIMARY KEY ,
		name TEXT,
		email TEXT,
		status TEXT
	);
	`
	_, err = db.Exec(createTb)
	if err != nil {
		log.Fatal("can't create", err.Error())
	}

}

func getCustomer(c *gin.Context){
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer db.Close()
	stmt, err := db.Prepare("SELECT * FROM customers")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rows, err := stmt.Query()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// data set row
	cusno := []Customer{}
	for rows.Next() {
		t := Customer{}
		err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		cusno = append(cusno, t)
		
	}
	  c.JSON(http.StatusOK, cusno)
	
}

func postCustomer(c *gin.Context){
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()

	t := Customer{}
	if err := c.ShouldBindJSON(&t) ; err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	
	query := `INSERT INTO customers (name, email, status) values ($1, $2, $3) RETURNING id`
	var id int
	row := db.QueryRow(query, &t.Name, &t.Email, &t.Status)
	err = row.Scan(&id)
	if err != nil {
		log.Fatal("can't scan id ", id)
	}
	t.ID = id
	fmt.Println(id)
	c.JSON(http.StatusCreated, t)
}
func getCustomerByID(c *gin.Context){
	
	id := c.Param("id")
	fmt.Println(id)

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()

	stmt, err:= db.Prepare("select id, name, email, status from customers where id= $1")	
	if err != nil {
		log.Fatal("error : can't prepare stmt ")
	}
	row := stmt.QueryRow(id)

	t := Customer{}
	err = row.Scan(&t.ID, &t.Name, &t.Email, &t.Status)
	if err != nil {
		log.Fatal("error", err.Error())
	}
	c.JSON(http.StatusOK, t)
}
	
func putCustomer(c *gin.Context){
	id := c.Param("id")
	ids, _ := strconv.Atoi(id)
	
	fmt.Println(id)
	t := Customer{}
	if err := c.ShouldBindJSON(&t) ; err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	stmt, err := db.Prepare("UPDATE customers SET name =$1 , email =$2, status =$3 WHERE id =$4;")
	if err != nil {
		log.Fatal("prepare error" , err.Error())
	}
	
	if _, err := stmt.Exec(&t.Name, &t.Email, &t.Status, id); err != nil{
		log.Fatal("exec error", err.Error())
	}
	t.ID = ids
	fmt.Println("update success")
	c.JSON(http.StatusOK, t)
}

func delCustomer(c *gin.Context){
	id := c.Param("id")
	ids, _ := strconv.Atoi(id)
	fmt.Println(id)

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM customers WHERE id =$1;")
	if err != nil {
		log.Fatal("delete error" , err.Error())
	}

	if _, err := stmt.Exec(ids); err != nil{
		log.Fatal("delete error", err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
}

func authMiddleware(c *gin.Context){

		token := c.GetHeader("Authorization")
		fmt.Println("Token", token)
  
		if token != "token2019" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
			c.Abort()
			return
		}
		c.Next()
	
}

func setupRouter() *gin.Engine{
	r := gin.Default()
	r.Use(authMiddleware)

	r.GET("/customers", getCustomer) 
	r.GET("/customers/:id", getCustomerByID) 
	r.POST("/customers", postCustomer)
	r.PUT("/customers/:id", putCustomer)
	r.DELETE("/customers/:id", delCustomer)
	
	return r
}

