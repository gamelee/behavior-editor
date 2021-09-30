package main

import (
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"
)

const (
	PathAssets  = `./assets/`
	PathPublic  = PathAssets + "public/"
	PathProject = PathAssets + "project/"
	PathConf    = PathAssets + "conf/"

	FileIrisConf = PathConf + "iris.tml"
)

func main() {

	app := iris.New()
	app.HandleDir("/", PathPublic, iris.DirOptions{
		IndexName: "index.html",
	})
	projectApi := app.Party("/v1/project")
	{
		projectApi.Get("/{user:string}", jsonRespWrap(listProject))
		projectApi.Get("/{user:string}/{id:string}", jsonRespWrap(getProject))
	}

	config := iris.WithConfiguration(iris.TOML(FileIrisConf))
	app.Run(iris.Addr(":8080"), config)
}

func jsonRespWrap(hd func(ctx context.Context) (interface{}, error)) context.Handler {
	return func(ctx context.Context) {
		rsp := Response{
			Code:    iris.StatusOK,
			Message: "success",
			Data:    nil,
		}
		err := Recovery(func() error {
			data, err := hd(ctx)
			if err != nil {
				return err
			}
			rsp.Data = data
			return nil
		})
		if err != nil {
			rsp.Message = err.Error()
		}
		ctx.Write(rsp.Bytes())
	}
}

func getProject(ctx context.Context) (interface{}, error) {
	user := strings.Trim(ctx.Params().GetString("user"), " ")
	proj := strings.Trim(ctx.Params().GetString("id"), " ")
	if user == "" || proj == ""{
		return nil, errors.New("未指定项目")
	}
	path := PathProject + user + "/" + proj

	stat, err := os.Stat(path)
	if err != nil {
		return nil, errors.New("项目不存在")
	}
	if stat.IsDir() {
		return nil, errors.New("项目状态异常")
	}
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("获取项目信息失败")
	}
	return buf, nil
}

func listProject(ctx context.Context) (interface{}, error) {
	user := strings.Trim(ctx.Params().GetString("user"), " ")
	if user == "" {
		return nil, errors.New("未指定用户")
	}
	stat, err := os.Stat(PathProject + user)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	if !stat.IsDir() {
		return nil, errors.New("用户状态异常")
	}
	fs, err := filepath.Glob(PathProject + user + "/*.json")
	if err != nil {
		return nil, errors.New("获取用户项目失败")
	}
	return fs, err
}

type Response struct {
	Code    int
	Message string
	Data    interface{}
}

const HttpJsonErrorFormatter = `{"code":500,"msg":"%s"}`

func (resp *Response) Bytes() []byte {
	buf, err := jsoniter.Marshal(resp)
	if err != nil {
		buf = String2bytes(fmt.Sprintf(HttpJsonErrorFormatter, err.Error()))
	}
	return buf
}

func Recovery(f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			buf := make([]byte, 64<<10)
			buf = buf[:runtime.Stack(buf, false)]
			err = fmt.Errorf("error: %w;\ntrace: %v", err, Bytes2String(buf))
		}
	}()
	err = f()
	return
}

func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func String2bytes(s string) []byte {
	tmp := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{tmp[0], tmp[1], tmp[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}
