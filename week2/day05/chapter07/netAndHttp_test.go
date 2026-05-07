package chapter07

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// 在这个文件，主要是学习这个go的网络编程:net/http

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// 向客户端返回数据
	w.Write([]byte("Hello Go HTTP!"))
}

func TestHttp(t *testing.T) {
	http.HandleFunc("/", helloHandler)
	// 注意，服务启动不能在测试函数进行，因为测试函数一直运行会超过固定时间会报错
	// 2. 启动服务：监听 8080 端口，nil 表示使用默认路由
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

// 路由注册的两种方式
// 方式1:直接注册函数

func helloLogin(w http.ResponseWriter, r *http.Request) {
	// 向客户端返回数据
	w.Write([]byte("这是一个处理登录的接口"))
}

func TestRouter(t *testing.T) {

	http.HandleFunc("/login", helloLogin)
	t.Log("服务启动成功")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

// 方式2:注册处理器对象
// 自定义结构体，实现 http.Handler 接口
type UserHandler struct {
	// 假设需要携带一些状态
	conf string
}

func (u *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("取到对象的状态", u.conf)
	w.Write([]byte("用户页面"))
}

// 注册路由
func TestRouterStruct(t *testing.T) {

	http.Handle("/user", &UserHandler{
		conf: "test",
	})
	t.Log("服务启动成功")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

// 处理http请求 get/post
func userHandler(w http.ResponseWriter, r *http.Request) {
	// 方法1：获取单个参数（推荐）
	name := r.URL.Query().Get("name")
	age := r.URL.Query().Get("age")

	// 方法2：获取所有参数
	// params := r.URL.Query()

	// 返回结果
	w.Write([]byte(fmt.Sprintf("姓名：%s，年龄：%s", name, age)))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// 必须先解析表单
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "参数解析失败", http.StatusBadRequest)
		return
	}

	// 获取表单参数
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")
	fmt.Println("获取到参数", password)

	w.Write([]byte(fmt.Sprintf("登录成功：用户名=%s", username)))
}

// 如何去处理这个json结构体
// 定义结构体接收 JSON
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	// 解码请求体中的 JSON 到结构体
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "JSON 格式错误", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	w.Write([]byte(fmt.Sprintf("JSON 参数：%+v", user)))
}

// 如何去构造这个http响应

func jsonRespHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头：告诉客户端返回 JSON 格式
	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	// 构造数据
	user := User{Username: "李四", Password: "123456"}

	// 编码为 JSON 并写入响应
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, "响应失败", http.StatusInternalServerError)
	}
}

func TestGetandPost(t *testing.T) {

	http.HandleFunc("/getuser", userHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/testJson", jsonHandler)
	t.Log("服务启动成功")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

// 如何请求第三方接口，调用API

func TestSendGet(t *testing.T) {
	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close() // 必须关闭响应体，防止内存泄漏

	// 读取响应数据
	body, _ := io.ReadAll(resp.Body)
	println(string(body))
	// 自定义 HTTP 客户端（设置 5 秒超时）
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 发送请求
	resp1, _ := client.Get("http://localhost:8080")
	defer resp1.Body.Close()

}

func TestSendPost(t *testing.T) {
	// 构造 JSON 数据
	user := User{Username: "测试", Password: "666"}
	jsonData, _ := json.Marshal(user)

	// 发送 POST 请求
	resp, _ := http.Post(
		"http://localhost:8080/json",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	defer resp.Body.Close()
}
