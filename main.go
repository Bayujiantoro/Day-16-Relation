package main

import (
	"context"
	"fmt"
	
	"io"
	"log"

	// "math"
	"net/http"
	"project-pertama/connection"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"project-pertama/middleware"

	"github.com/labstack/echo-contrib/session"
)

// "github.com/gorilla/sessions"


type Template struct {
	templates *template.Template
}

type FormatProject struct{
	ID int
	TitleProject string
	Duration string
	StartDate time.Time
	EndDate time.Time
	Description string
	Technology []string
	Formatstart string
	FStartDate string
	FEndDate string
	Author string
	Image string
}

type Users struct{
	ID int
	Name string
	Email string
	Password string
}

// var DataProject = []FormatProject {
// 	{
// 		TitleProject: "Web Store",
// 		Duration: "3 bulan",
// 		Description: "Lorem ipsum dolor sit, amet consectetur adipisicing elit. Incidunt molestiae ipsam atque est impedit consectetur enim molestias officia sunt necessitatibus dignissimos mollitia quidem saepe cupiditate labore pariatur, obcaecati quo aperiam.",
		
// 	},
// 	{
// 		TitleProject: "Web Store",
// 		Duration: "3 bulan",
// 		Description: "Lorem ipsum dolor sit, amet consectetur adipisicing elit. Incidunt molestiae ipsam atque est impedit consectetur enim molestias officia sunt necessitatibus dignissimos mollitia quidem saepe cupiditate labore pariatur, obcaecati quo aperiam.",
		
// 	},
// }

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {

	connection.DatabaseConnect()
	e := echo.New()

	// root statis untuk mengakses folder public
	e.Static("/public", "public") //public
	e.Static("/upload", "upload") //upload

	// untuk menambahkan midleware untuk penghubung
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("session"))))

	t := &Template{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}

	// renderer
	e.Renderer = t

	// routing
	e.GET("/", home)
	e.GET("/contact", contactMe)
	e.GET("/form-project", formProject)
	e.GET("/project-detail/:id", projectDetail)
	e.POST("/add-project", middleware.UploadFile(addProject))
	e.GET("/testimoni" , testimoni)
	e.GET("/delete-project/:id", deleteProject)
	e.GET("/form-register", formRegister)
	e.GET("/form-login", formLogin)
	e.POST("/register", register)
	e.POST("/login", login)
	e.GET("/logout", logout)

	fmt.Println("localhost: 5004 sucssesfully")
	e.Logger.Fatal(e.Start("localhost: 5004"))
}

// <span class="icon d-flex flex-row">
// 	{{range $index, $data := $data.Technologies}}
// 	<i class="fab fa-{{$data}} me-3"></i>
// 	{{end}}
// </span>



func formLogin(c echo.Context) error {
	sess, _ := session.Get("session", c)

	delete(sess.Values, "message")
	delete(sess.Values, "status")

	tmpl, err := template.ParseFiles("views/form-login.html")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message" : err.Error()})
	}

	return tmpl.Execute(c.Response(),nil)
}


// func formLogin(c echo.Context) error {
// 	return c.Render(http.StatusOK, "form-login.html", nil)
// }

func formRegister(c echo.Context) error {
	return c.Render(http.StatusOK, "form-register.html", nil)
}

func testimoni(c echo.Context) error {
	return c.Render(http.StatusOK, "testimoni.html", nil)
}

func deleteProject(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_project WHERE id=$1", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"Message ": err.Error()})
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func home(c echo.Context) error {
	sess,_ := session.Get("session", c)
	

	

	data,_ :=connection.Conn.Query(context.Background(), "SELECT tb_project.id,title,description,technology,start_date,end_date, images, users.name as author FROM tb_project LEFT JOIN users ON tb_project.author = users.id ORDER BY id DESC")

	var result []FormatProject
	for data.Next() {
		var each = FormatProject{}

		err := data.Scan(&each.ID, &each.TitleProject, &each.Description, &each.Technology, &each.StartDate, &each.EndDate, &each.Image, &each.Author)

		if err != nil {
			fmt.Println(err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"Message ": err.Error()})
		}

		// duration := each.EndDate.Sub(each.StartDate)
		// resultTime := math.Floor(duration.Hours())
		// println(duration)
		result = append(result, each)
	}
	
	fmt.Println(result[5].Technology)
	
	blogs := map[string]interface{}{
		"DataProjects": result,
		"FlashStatus" : sess.Values["status"],
		"FlashMessage" : sess.Values["message"],
		"FlashName" : sess.Values["name"],
	}

	delete(sess.Values, "message")
	// delete(sess.Values, "status")
	sess.Save(c.Request(),c.Response())
	
	return c.Render(http.StatusOK, "index.html", blogs)
}


func contactMe(c echo.Context) error {
	return c.Render(http.StatusOK, "contact-form.html", nil)
}

func formProject(c echo.Context) error {
	return c.Render(http.StatusOK, "form-project.html", nil)
}

func projectDetail(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var result = FormatProject{}


	err := connection.Conn.QueryRow(context.Background(), "SELECT tb_project.id,title,description,technology,start_date,end_date, images, users.name as author FROM tb_project LEFT JOIN users ON tb_project.author = users.id WHERE tb_project.id=$1", id).Scan(&result.ID , &result.TitleProject, &result.Description ,&result.Technology,  &result.StartDate , &result.EndDate, &result.Image, &result.Author)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"Message ": err.Error()})
	}

	// data := map[string]interface{}{
	// 	"Blog": BlogDetail,
	// }

	// return c.Render(http.StatusOK, "blog-detail.html", data)

	result.FStartDate = result.StartDate.Format("07 February 2006")
	result.FEndDate = result.EndDate.Format("07 February 2006")

	data := map[string]interface{}{
		"Projects":      result,
		
	}
	return c.Render(http.StatusOK, "blog-project.html", data)
}

func addProject(c echo.Context) error {
	name := c.FormValue("nameProject")
	startDate := c.FormValue("startDate")
	endDate := c.FormValue("endDate")
	description := c.FormValue("description")
	nodeJs := c.FormValue("nodeJs")
	nextJs := c.FormValue("nextJs")
	reactJs := c.FormValue("reactJs")
	typeScript := c.FormValue("typeScript")
	image := c.Get("dataFile").(string)

	sess,_ := session.Get("session", c)
	authorId := sess.Values["id"]

	var tech [] string

	if nodeJs == "on" {
		tech = append(tech, "NodeJS")
	}
	if nextJs == "on" {
		tech = append(tech, "NextJS")
	}
	if reactJs == "on" {
		tech = append(tech, "ReactJS")
	}
	if typeScript == "on" {
		tech = append(tech, "Typescript")
	}


	_, err := connection.Conn.Exec(context.Background(), "INSERT INTO public.tb_project(title, description, start_date, end_date, technology, author, images) VALUES($1, $2, $3, $4, $5, $6, $7)",name , description , startDate, endDate, tech, authorId, image)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"Message ": err.Error()})
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}


func register( c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")

	fmt.Println("ini nama : "+name)
	fmt.Println("ini email : " + email)
	fmt.Println("ini password : " +password)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password),10)

	_, errr := connection.Conn.Exec(context.Background(), "INSERT INTO public.users(name, email, password) VALUES($1, $2, $3)" ,name , email , passwordHash)


	if errr != nil {
		 redirectWithMessages(c, "Register failed , please try again", false , "/form-register")

		
	}

	return redirectWithMessages(c, "Register Success", true , "/form-login")

}

func login(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	email := c.FormValue("email")
	password := c.FormValue("password")

	user := Users{}

	errs := connection.Conn.QueryRow(context.Background(),"SELECT * FROM users WHERE email=$1", email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	
	if errs != nil {
		return redirectWithMessages(c, "Email Salah Bro", false , "/form-login")
	}

	err = bcrypt.CompareHashAndPassword( []byte(user.Password), []byte(password))

	fmt.Println(user.Password)

	if err != nil {
		fmt.Println("password Salah")
		return redirectWithMessages(c, "Password Anda Salah !!", false , "/form-login")
	}
	

	fmt.Println("after")
	fmt.Println(user.Email)

	sess,_:= session.Get("session" , c)
	sess.Options.MaxAge = 10800
	sess.Values["message"] = "Login Success"
	sess.Values["status"] = true
	sess.Values["name"] = user.Name
	sess.Values["id"] = user.ID
	sess.Values["isLogin"] = true
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func logout(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func redirectWithMessages(c echo.Context, message string , status bool, path string) error {
	sess, _ := session.Get("session" , c)
	sess.Values["message"] = message
	sess.Values["status"] = status
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusMovedPermanently, path)
}