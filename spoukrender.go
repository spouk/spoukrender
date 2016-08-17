package spoukrender

import (
	"html/template"
	"io"
	"fmt"
	"strings"
	"github.com/labstack/echo"
	"bytes"
	"github.com/labstack/gommon/log"
)

const (
	ErrorRenderContent = "\n[spoukrender][content] `%s`\n"
	ErrorCatcherPanic =  "\n[spoukrender][catcherPanic] `%v`\n"
)
//определение дефолтных значений фильтров и функций
var (
	Local_filter = map[string]interface{}{
		"count" : strings.Count,
		"split" : strings.Split,
		"title" : strings.Title,
		"lower" : strings.ToLower,
		"totitle" : strings.ToTitle,
		"makemap":MakeMap,
		"in":MapIn,
		"andlist":AndList,
		"upper" : strings.ToUpper,
		"html2" : func(value string) template.HTML {
			return template.HTML(fmt.Sprint(value))
		},
	}
)

type (
//рендеринг инстанс
	SpoukRender struct {
		Temp        *template.Template
		FIlters     template.FuncMap
		Debug       bool
		Path        string
	}
)

func NewSpoukRender(path string, debug bool) *SpoukRender {
	//создаю стек для дефолтных функций
	sf := new(SpoukRender)
	sf.FIlters = template.FuncMap{}
	sf.AddFilters(Local_filter)
	sf.Path = path
	sf.Debug = debug
	defer sf.catcherPanic()
	sf.Temp = template.Must(template.New("spoukindex").Funcs(sf.FIlters).ParseGlob(sf.Path))
	return sf
}
func (s *SpoukRender) catcherPanic() {
	msgPanic := recover()
	if msgPanic != nil {
		log.Warn(fmt.Sprintf(ErrorCatcherPanic, msgPanic))
	}
}
func (s *SpoukRender) reloadTemplate() {
	defer s.catcherPanic()
	if s.Debug {
		s.Temp = template.Must(template.New("spoukindex").Funcs(s.FIlters).ParseGlob(s.Path))
	}
}
func (s *SpoukRender) HTMLTrims(body []byte) []byte {
	result := []string{}
	for _, line := range strings.Split(string(body), "\n") {
		if len(line) != 0 && len(strings.TrimSpace(line)) != 0 {
			result = append(result, line)
		}
	}
	return []byte(strings.Join(result, "\n"))
}
func (s *SpoukRender) SpoukRenderContext(code int, name string, data interface{}, c *echo.Context) (err error) {
	buf := new(bytes.Buffer)
	if err = s.Temp.ExecuteTemplate(buf, name, data); err != nil {
		log.Warn(fmt.Sprintf(ErrorRenderContent, err.Error()))
		return
	}
	resp := c.Response()
	resp.Header().Add("Content-Type", "text/html;charset=utf-8")
	resp.WriteHeader(code)
	resp.Write(s.HTMLTrims(buf.Bytes()))
	return
}
//фуункция рендеринг, такой формат функции требует `Echo`.rendere для замещения дефолтного
//want Render(io.Writer, string, interface {}, echo.Context) error {v.2}
//have Render(io.Writer, string, interface {}) error {{v.1}}
func (s *SpoukRender) Render(w io.Writer, name string, data interface{}) error {
	defer s.catcherPanic()
	s.reloadTemplate()
	if err := s.Temp.ExecuteTemplate(w, name, data); err != nil {
		log.Warn(err)
		return err
	}
	return nil
}
func (s *SpoukRender) AddUserFilter(name string, f interface{}) {
	s.FIlters[name] = f
}
func (s *SpoukRender) AddFilters(stack map[string]interface{}) {
	for k, v := range stack {
		s.FIlters[k] = v
	}
}
func (s *SpoukRender) ShowStack() {
	fmt.Println(s.FIlters)
}
//---------------------------------------------------------------------------
//  рандомные полезные функции для шаблонов
//---------------------------------------------------------------------------
func MapIn(value interface{}, stock interface{}) bool {
	switch value.(type) {
	case int64:
		for _, x := range stock.([]int64) {
			if x == value.(int64) {
				return true
			}
		}
	case int:
		for _, x := range stock.([]int) {
			if x == value.(int) {
				return true
			}
		}
	case string:
		for _, x := range stock.([]string) {
			if x == value.(string) {
				return true
			}
		}

	}
	return false
}
func MakeMap(value ...string) ([]string) {
	return value
}
func AndList(listValues ...interface{}) (bool) {
	for _, v := range listValues {
		if v == nil {
			return false
		}
	}
	return true
}
