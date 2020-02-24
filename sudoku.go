package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const STR_END = ""
const STR_LINE_SPACE = "\n  -----+------------------+------------------+------------------+-"

type Record struct {
	palace   int
	x        int
	y        int
	value    int
	isIssues bool

	up    *Record
	down  *Record
	left  *Record
	right *Record
}

func (r *Record) HasDown() bool {
	return r.down != nil
}

func (r *Record) HasUp() bool {
	return r.up != nil
}

func (r *Record) HasRight() bool {
	return r.right != nil
}

func (r *Record) HasLeft() bool {
	return r.left != nil
}

func (r *Record) GetDown() *Record {
	return r.down
}

func (r *Record) GetUp() *Record {
	return r.up
}

func (r *Record) GetRight() *Record {
	return r.right
}

func (r *Record) GetLeft() *Record {
	return r.left
}

type Sudoku struct {
	data [][][]Record
	//answer [][][]Record

	rowLinkTable    []*Record
	colLinkTable    []*Record
	palaceLinkTable [][]*Record

	finished bool
	writer   *io.Writer
}

func (s *Sudoku) Init() {

	s.data = make([][][]Record, 9)
	for i := 0; i < len(s.data); i++ {
		s.data[i] = make([][]Record, 3)
		for j := 0; j < len(s.data[i]); j++ {
			s.data[i][j] = make([]Record, 3)
		}
	}
	s.finished = false

	for x := 1; x <= 9; x++ {
		for y := 1; y <= 9; y++ {
			s.SetValue(x, y, 0, false)
		}
	}

	s.rowLinkTable = make([]*Record, 9)
	for i := 1; i <= 9; i++ {
		s.rowLinkTable[i-1] = s.GetValue(i, 1)
	}

	s.colLinkTable = make([]*Record, 9)
	for i := 1; i <= 9; i++ {
		s.colLinkTable[i-1] = s.GetValue(1, i)
	}

	s.palaceLinkTable = make([][]*Record, 9)
	for i := 0; i < 9; i++ {

		s.palaceLinkTable[i] = make([]*Record, 9)

		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				s.palaceLinkTable[i][(j*3)+k] = &s.data[i][j][k]
			}
		}
	}
}

func (s *Sudoku) GetColHeader(x int) *Record {

	return s.colLinkTable[x]
}

func (s *Sudoku) GetRowHeader(y int) *Record {

	return s.rowLinkTable[y]
}

func (s *Sudoku) GetPalace(i int) []*Record {
	return s.palaceLinkTable[i]
}

func (s *Sudoku) GetValue(x, y int) *Record {

	r := &s.data[s.palace(x, y)-1][(x-1)%3][(y-1)%3]
	return r
}

func (s *Sudoku) SetValue(x, y, value int, isIssues bool) {

	r := s.data[s.palace(x, y)-1][(x-1)%3][(y-1)%3]
	r.palace = s.palace(x, y)
	r.x = x
	r.y = y
	r.value = value
	r.isIssues = isIssues

	if x > 1 && r.up == nil {
		r.up = &s.data[s.palace(x-1, y)-1][(x-1-1)%3][(y-1)%3]
	}
	if x < 9 && r.down == nil {
		r.down = &s.data[s.palace(x+1, y)-1][(x-1+1)%3][(y-1)%3]
	}
	if y > 1 && r.left == nil {
		r.left = &s.data[s.palace(x, y-1)-1][(x-1)%3][(y-1-1)%3]
	}
	if y < 9 && r.right == nil {
		r.right = &s.data[s.palace(x, y+1)-1][(x-1)%3][(y-1+1)%3]
	}
	s.data[s.palace(x, y)-1][(x-1)%3][(y-1)%3] = r
}

func (s *Sudoku) Fill(rs []Record) {

	for _, v := range rs {
		s.SetValue(v.x, v.y, v.value, v.isIssues)
	}
}

func (s *Sudoku) palace(x, y int) int {

	row := (x - 1) / 3
	col := (y - 1) / 3

	switch {
	case row == 0 && col == 0:
		return 1
	case row == 1 && col == 0:
		return 4
	case row == 2 && col == 0:
		return 7
	case row == 0 && col == 1:
		return 2
	case row == 1 && col == 1:
		return 5
	case row == 2 && col == 1:
		return 8
	case row == 0 && col == 2:
		return 3
	case row == 1 && col == 2:
		return 6
	case row == 2 && col == 2:
		return 9
	default:
		errors.New("划分宫错误.")
	}
	return 0
}

func (s *Sudoku) PrintAll() {

	fmt.Printf("\033[2J\n       |")
	for i := 1; i <= 9; i++ {
		fmt.Printf("  %2d  ", i)
		if i%3 == 0 {
			fmt.Print("|")
		}
	}
	fmt.Print(STR_LINE_SPACE)

	for i := 1; i <= 9; i++ {

		fmt.Printf("\n  %2d   |", i)
		s.WalkRow(i, func(r Record) {
			if r.isIssues {
				fmt.Printf("  %c[31m%2d%c[0m  ", 0x1B, r.value, 0x1B)
			} else if r.value == 0 {
				fmt.Printf("  %2d  ", r.value)
			} else {
				fmt.Printf("  %c[33m%2d%c[0m  ", 0x1B, r.value, 0x1B)
			}

			if r.y%3 == 0 {

				fmt.Printf("|")
			}
		})
		if i%3 == 0 {
			fmt.Print(STR_LINE_SPACE)
		}
	}
	fmt.Println()
}

func (s *Sudoku) PrintRow(rowNum int) {

	fmt.Printf("Line %d: { ", rowNum)
	s.WalkRow(rowNum, func(r Record) {
		fmt.Printf("%d, ", r.value)
	})
	fmt.Printf(" }\n")
}

func (s *Sudoku) PrintCol(colNum int) {

	fmt.Printf("Column %d: { ", colNum)
	s.WalkCol(colNum, func(r Record) {
		fmt.Printf("%d, ", r.value)
	})
	fmt.Printf(" }\n")
}

func (s *Sudoku) PrintPlace(i int) {

	fmt.Printf("Place %d: { ", i)
	s.WalkPalace(i, func(r *Record) {
		//fmt.Println(r)
		fmt.Printf("[%d x %d]: %d, ", r.x, r.y, r.value)
	})
	fmt.Printf(" }\n")

}

func (s *Sudoku) WalkPalace(palaceNum int, walkFunc func(r *Record)) {

	for i := 0; i < len(s.palaceLinkTable[palaceNum-1]); i++ {
		walkFunc(s.palaceLinkTable[palaceNum-1][i])
	}

}

func (s *Sudoku) WalkCol(colNum int, printFunc func(r Record)) {

	colLinkTable := make([]Record, 9)
	for i := 1; i <= 9; i++ {
		colLinkTable[i-1] = *s.GetValue(1, i)
	}

	r := colLinkTable[colNum-1]
	for {
		printFunc(r)
		if r.down != nil {
			r = *r.down
		} else {
			break
		}
	}
}

func (s *Sudoku) WalkRow(rowNum int, printFunc func(r Record)) {

	rowLinkTable := make([]Record, 9)
	for i := 1; i <= 9; i++ {
		rowLinkTable[i-1] = *s.GetValue(i, 1)
	}

	r := rowLinkTable[rowNum-1]
	for {
		printFunc(r)
		if r.right != nil {
			r = *r.right
		} else {
			break
		}
	}
}

func (s *Sudoku) Check(r *Record, v int) bool {

	if s.CheckRow(r.x, v) && s.CheckCol(r.y, v) && s.CheckPalace(r.palace, v) {
		return true
	} else {
		return false
	}
}

func (s *Sudoku) CheckPalace(i int, v int) bool {

	temp := make([]int, 10)

	s.WalkPalace(i, func(r *Record) {

		if r.value == 0 {
			return
		} else {
			temp[r.value] = 1
		}
	})
	return temp[v] == 0
}

func (s *Sudoku) CheckCol(x int, v int) bool {

	temp := make([]int, 10)

	for i := 1; i <= 9; i++ {

		r := s.GetValue(i, x)

		if r.value == 0 {
			continue
		} else {
			temp[r.value] = 1
		}
	}

	return temp[v] == 0
}

func (s *Sudoku) CheckRow(x int, v int) bool {

	temp := make([]int, 10)

	for i := 1; i <= 9; i++ {

		r := s.GetValue(x, i)

		if r.value == 0 {
			continue
		} else {
			temp[r.value] = 1
		}
	}

	return temp[v] == 0
}

func (s *Sudoku) Solving() {

	s.backTrace(1, 1)
}

func (s *Sudoku) backTrace(x int, y int) {

	if x == 10 && y == 1 {

		s.httpDisplay(func(sb *strings.Builder) {

			for i := 1; i <= 9; i++ {
				sb.WriteString(" <tr>")
				for j := 1; j <= 9; j++ {

					strClass := `xx`
					strClass2 := `fix`
					r := s.GetValue(i, j)

					if r.isIssues {

						strClass2 = `big`
					}

					if i%3 == 0 {
						strClass = `bb`
						if j%3 == 0 {
							strClass = `br`
						}
					} else {

						if j%3 == 0 {
							strClass = `rr`
						}
					}

					sb.WriteString(fmt.Sprintf("<td class=\"%s\"><input id=\"%dx%d\" name=\"%dx%d\" class=\"%s\" maxlength=\"1\" value=\"%d\"></td>\n", strClass, i, j, i, j, strClass2, r.value))
				}
				sb.WriteString(" </tr>")
			}

		})
		//s.PrintAll()
		s.finished = true
		return
	}

	r := s.GetValue(x, y)

	if r.value == 0 {

		for i := 1; i <= 9; i++ {

			if s.Check(r, i) {

				s.SetValue(x, y, i, false)
				s.backTrace(s.NextAddr(x, y))

				s.SetValue(x, y, 0, false)

				if s.finished {
					break
				}
			}

		}
	} else {

		s.backTrace(s.NextAddr(x, y))
	}
}

func (s Sudoku) parseLine(line []byte) (a int, b int, v int) {

	re := regexp.MustCompile(`([0-9])\s+([0-9])\s+([0-9])`)
	matches := re.FindAllSubmatch(line, -1)

	for _, m := range matches {
		a, _ = strconv.Atoi(string(m[1]))
		b, _ = strconv.Atoi(string(m[2]))
		v, _ = strconv.Atoi(string(m[3]))
	}

	return
}

//func (s *Sudoku) Start(input io.Reader) {
//
//	s.Init()
//
//	for {
//
//		s.PrintAll()
//		fmt.Printf("\n\r输入位置及数字：")
//		reader := bufio.NewReader(input)
//
//		line, _, err := reader.ReadLine()
//		if err != nil {
//			log.Panic(err)
//		}
//
//		if bytes.Equal(line, []byte(STR_END)) {
//			fmt.Printf("\n\n输入完毕。")
//			break
//		}
//
//		a, b, v := s.parseLine(line)
//
//		r := Record{x: a, y: b, value: v, isIssues: true}
//		records = append(records, r)
//
//		s.Fill(records)
//	}
//
//	s.PrintAll()
//	s.Solving()
//}
func (s *Sudoku) NextAddr(x int, y int) (int, int) {

	y++

	if y > 9 {
		x++
		y = 1
	}

	return x, y
}

func main() {

	sudoku := Sudoku{}
	//sudoku.Start(os.Stdin)

	//funcMaps := template.FuncMap{"fdate": fdate}
	//t := template.New("./templates/layout.html").Funcs(funcMaps)
	//t, err := t.Parse(tmpl)
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println("-----------------------")
	//fmt.Println(t.Name())
	//fmt.Println("-----------------------")
	//t.Execute(w, "hello word")

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

		//fmt.Println()
		if req.Method == http.MethodGet {

			w.Header().Set("Content-Type", "text/html")

			sudoku.Init()

			writer := w.(io.Writer)
			sudoku.setWriter(&writer)
			sudoku.httpDisplay(func(sb *strings.Builder) {
				for i := 1; i <= 9; i++ {
					sb.WriteString(" <tr>")
					for j := 1; j <= 9; j++ {

						strClass := `xx`

						if i%3 == 0 {
							strClass = `bb`
							if j%3 == 0 {
								strClass = `br`
							}
						} else {

							if j%3 == 0 {
								strClass = `rr`
							}
						}
						sb.WriteString(fmt.Sprintf("<td class=\"%s\"><input id=\"%dx%d\" name=\"%dx%d\" class=\"fix\" maxlength=\"1\"></td>\n", strClass, i, j, i, j))
					}
					sb.WriteString(" </tr>")
				}
			})
			//io.WriteString(w, str_html)
		} else {

			var records []Record

			writer := w.(io.Writer)
			sudoku.setWriter(&writer)

			req.ParseForm()

			for x := 1; x <= 9; x++ {
				for y := 1; y <= 9; y++ {
					v := strings.Trim(req.Form.Get(fmt.Sprintf("%dx%d", x, y)), "")
					if v != "" {

						value, err := strconv.Atoi(v)
						if err != nil {
							io.WriteString(w, fmt.Sprintf("%v", err))
						}

						r := Record{x: x, y: y, value: value, isIssues: true}
						records = append(records, r)

						sudoku.Fill(records)
					}
				}
			}

			//sudoku.PrintAll()
			sudoku.Solving()

		}
	})

	fmt.Printf("数独解题程序已启动\n请自行使用浏览器访问页面, 地址: http://localhost:8088\n")
	log.Fatal(http.ListenAndServe(":8088", nil))

}

func (s *Sudoku) httpDisplay(displayFunc func(sb *strings.Builder)) {

	str_html := strings.Builder{}
	str_html.WriteString(HTML_HEAD_CODE)
	displayFunc(&str_html)
	str_html.WriteString(HTML_FOOT_CODE)

	io.WriteString(*s.writer, str_html.String())
	//return str_html.String()
}
func (s *Sudoku) setWriter(w *io.Writer) {
	s.writer = w
}

var HTML_HEAD_CODE = `

<!DOCTYPE html>
<head>
    <meta charset="UTF-8">
    <title>数独</title>
</head>
<body>

<style>

    .bk {table-layout: fixed; background-color:#fff; width: 420px; vertical-align: middle;  border-collapse: collapse; text-align: center;}

    .sd {table-layout: fixed;border: #443 3px solid;width: 355px; height:355px; background-color:#fff;
        vertical-align: middle;  border-collapse: collapse; text-align: center; margin-top:5%}

    td.xx { border-right: #999 1px solid; border-top: #999 1px solid; width: 30px; height: 30px; text-align: center; LINE-height:30px; }
    td.rr { border-right: #443 2px solid; border-top: #999 1px solid; width: 30px; height: 30px; text-align: center; LINE-height:30px;}
    td.bb { border-right: #999 1px solid; border-top: #999 1px solid; border-bottom: #443 2px solid; width: 30px; height: 30px; text-align: center;  LINE-height:30px; }
    td.br { border-right: #443 2px solid; border-top: #999 1px solid; border-bottom: #443 2px solid; width: 30px; height: 30px; text-align: center;  LINE-height:30px; }

    .fix {FONT-SIZE: 25px; border:none;background-color: transparent; WIDTH: 30px;HEIGHT: 30px; LINE-HEIGHT: 28px;TEXT-ALIGN: center; margin:0px; COLOR: #000000; FONT-FAMILY:  Verdana;}
    .fix2 {FONT-SIZE: 25px; border:2px solid red;background-color: transparent; WIDTH: 26px;HEIGHT: 26px; LINE-HEIGHT: 25px;TEXT-ALIGN: center; margin:0px; COLOR: #000000; FONT-FAMILY:  Verdana;}
    .big {FONT-SIZE: 25px; border:none;background-color: transparent; WIDTH: 30px;HEIGHT: 30px; LINE-HEIGHT: 28px;TEXT-ALIGN: center; margin:0px; COLOR: red; FONT-FAMILY:  Verdana;}
    .big2 {FONT-SIZE: 25px; border:2px solid red;background-color: transparent; WIDTH: 26px;HEIGHT: 26px; LINE-HEIGHT: 25px;TEXT-ALIGN: center; margin:0px; COLOR: #0000FF; FONT-FAMILY:  Verdana;}

    .err {FONT-SIZE: 25px; border:none;background-color: #ff0000; WIDTH: 30px;HEIGHT: 30px; LINE-HEIGHT: 28px;TEXT-ALIGN: center; margin:0px; COLOR: #0000FF; FONT-FAMILY:  Verdana;}
    .cad {FONT-SIZE: 10px; border:none;background-color: transparent; WIDTH: 33px;HEIGHT: 28px; LINE-HEIGHT: 28px;TEXT-ALIGN: center; margin:0px; COLOR: #0000FF; FONT-FAMILY:  Verdana;}

    .fieldset{background-color:#fff;font:14px bold;color:#0000FF; cursor:default; border:2px blue solid;padding-bottom:5px;}
    .radio{background-color:#fff;font:14px bold;color:#000000; cursor:default}
    /*.button{font-family: Verdana; font-size: 12px;  height:25px;}*/

    td.sel {
        width: 25px;
        height: 25px;
        text-align: center;
        LINE-height:25px;
        border:0px solid blue;
    }
    .numsel{background-color:#fff;
        font:20px bold;color:blue;width:30px;height:25px;
        border:2px solid blue;
        text-align: center; cursor:pointer;padding: 0px;
    }
    .numsel2{background-color:#aaa;font:20px bold;color:blue;width:30px;height:25px;
        border:2px solid #blue;
        text-align: center; cursor:pointer;padding: 0px;}

    ul#nav {
        clear:both;
        float:left;
        margin: 0px 0px 0px 5px;
        padding: 0px 0px 0px 0px;
        list-style: none;
        width:100%;
        color:#FFFFFF;
        background-color:#fff;
        font-size: 12px;
        font-family: "Times New Roman", Times, serif;
    }

    ul#nav li {
        margin: 0px 0px 0px 0px;
        padding: 0px 3px 0px 0px;
        float: left;
    }

    ul#nav li a {
        display: block;
        margin: 0px 2px 0px 0px;
        padding: 0px 5px 0px 5px;
        color: #FFFFFF;
        background-color:#0065CE;
        font-weight: normal;
        text-decoration: none;
    }

    ul#nav li a:hover {
        background-color:#FF0;
        color: #000;
    }

    ul#nav a.selected {
        background-color:#FF0;
        color: #000;
    }

    ul#nav2 {
        float:left;
        list-style: none;
    }

    ul#nav2 li {
        margin: 0px 0px 0px 0px;
        padding: 0px 1px 0px 0px;
        float: right;
    }

    input.button {
        float: right;
        margin: 20px 0 10px 5px;
        padding: 5px 20px;
        font: 600 13px "Source Sans Pro", sans-serif;
        text-transform: uppercase;
        letter-spacing: 1px;
        background: #ffbb33;
        margin-top: 20px;
        margin-right:35%

    }

</style>

<form method="post" action="/">
	<h1 align="center" style="margin-top:5%">好好学习，少玩电脑!</h1>
    <table class="sd" border="0" align="center" cellspacing="1" cellpadding="1">
        <tbody>
`

var HTML_FOOT_CODE = `
        </tbody></table>
    <input type="submit" value=" 计 算 " class="button" >
</form>
</body>
</html>
`
