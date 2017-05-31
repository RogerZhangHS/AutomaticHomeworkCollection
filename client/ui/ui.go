package ui

import (
	"bytes"
//	"encoding/json"
	"fmt"
	"PiScanStudent/client/database"
	"github.com/mxk/go-sqlite/sqlite3"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
//	"PiScanStudent/client/ui"
	"encoding/json"
)

const (
	// Errors
	BAD_REQUEST = "Sorry, that is an invalid request"
	BAD_POST    = "Sorry, we cannot respond to that request. Please try again."

	// Info messages
	EMAIL_SENT = "The selected items have been sent to your email address"

	// urls
	HOME_URL    = "/submissions/"
	STUDENTS_URL = "/students/"
)

var (
	TEMPLATE_LIST = func(templatesFolder string, templateFiles []string) []string {
		t := make([]string, 0)
		for _, f := range templateFiles {
			t = append(t, path.Join(templatesFolder, f))
		}
		return t
	}

	UNSUPPORTED_TEMPLATE_FILE = "browser_not_supported.html"

	SUBMISSION_LIST_TEMPLATES *template.Template
	SUBMISSION_EDIT_TEMPLATES *template.Template

	TEMPLATES_INITIALIZED = false

	SUBMISSION_LIST_TEMPLATE_FILES = []string{"items.html", "head.html", "navigation_tabs.html", "actions.html", "modal.html", "scripts.html"}
	SUBMISSION_EDIT_TEMPLATE_FILES = []string{"define_item.html", "head.html", "scripts.html"}
)

// Use this to redirect one request to another target (string)
//COMPLETED
func Redirect(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusFound)
	}
}

// Respond to requests using HTML templates and the standard Content-Type (i.e., "text/html")
//COMPLETED
func MakeHTMLHandler(fn func(http.ResponseWriter, *http.Request, database.ConnCoordinates, ...interface{}), db database.ConnCoordinates, opts ...interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, db, opts...)
	}
}

// Show the static template for unsupported browsers
//COMPLETED
func UnsupportedBrowserHandler(templatesFolder string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadFile(path.Join(templatesFolder, UNSUPPORTED_TEMPLATE_FILE))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, string(body))
	}
}

// Respond to requests that are not "text/html" Content-Types (e.g., for ajax calls)
//COMPLETED
func MakeHandler(fn func(*http.Request, database.ConnCoordinates, ...interface{}) string, db database.ConnCoordinates, mediaType string, opts ...interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", fmt.Sprintf("%s; charset=utf-8", mediaType))
		data := fn(r, db, opts...)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		fmt.Fprintf(w, data)
	}
}

/* JSON response struct */
//COMPLETED
type AjaxAck struct {
	Message string `json:"msg"`
	Error   string `json:"err,omitempty"`
}

/* HTML template structs */
//COMPLETED
type ActiveTab struct {
	Submission bool
	ShowTabs bool
}

//COMPLETED
type Action struct {
	Icon   string
	Link   string
	Action string
}

type SubmissionPage struct {
	Title string
	ActiveTab *ActiveTab
	Actions []*Action
	Students []*database.Students
	PageMessage string
}

type StudentForm struct {
	Title string
	Student *database.Students
	Submitted bool
	CancelUrl string
	FormError string
	FormMessage string
}

// 为页面获取Student提交状态的列表
func getStudents(w http.ResponseWriter, r *http.Request, dbCoords database.ConnCoordinates) {
	db, err := database.InitializeDB(dbCoords)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	fetch := func(db *sqlite3.Conn) ([]*database.Students, error) {
		return database.SearchAllStudents(db)
	}

	students := make([]*database.Students, 0)
	studentList, studentsErr := fetch(db)
	if studentsErr != nil {
		http.Error(w, studentsErr.Error(), http.StatusInternalServerError)
		return
	}
	for _, student := range studentList {
		students = append(students, student)
	}

	// 对于这些Submission的动作
	// TODO: 补全针对Submissions的Actions
	actions := make([]*Action, 0)
	actions = append(actions, &Action{Link: "/delete/", Icon: "fa fa-trash", Action: "Delete"})

	// 定义网页标题
	var titleBuffer bytes.Buffer
	titleBuffer.WriteString("Submission")

	p := &SubmissionPage{Title: titleBuffer.String(),
		ActiveTab:          &ActiveTab{Submission:true,ShowTabs:true},
		Actions:            actions,
		Students:           students}

	r.ParseForm()
	if msg, exists := r.Form["ack"]; exists {
		ackType := strings.Join(msg,"")
		if ackType == "email" {
			p.PageMessage = EMAIL_SENT
		}
	}

	renderSubmissionListTemplate(w, p)
}

// 处理Student
func processStudents(w http.ResponseWriter, r *http.Request, dbCoords database.ConnCoordinates, fn func(students *database.Students, db *sqlite3.Conn), successTarget string) {
	// attempt to connect to the db
	db, err := database.InitializeDB(dbCoords)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// get all the Items for this Account
	// and store them in a map by their Idscanned/
	students, studentsErr := database.SearchAllStudents(db)
	if studentsErr != nil {
		http.Error(w, studentsErr.Error(), http.StatusInternalServerError)
		return
	}
	thisSubmissions := make(map[int64]*database.Students)
	for _, submission := range students {
		thisSubmissions[submission.StuId] = submission
	}

	// get the list of item ids from the POST values
	// and apply the processing function
	if "POST" == r.Method {
		r.ParseForm()
		if idVals, exists := r.PostForm["student"]; exists {
			for _, idString := range idVals {
				id, idErr := strconv.ParseInt(idString, 10, 64)
				if idErr == nil {
					if accountItem, ok := thisSubmissions[id]; ok {
						fn(accountItem, db)
					}
				}
			}
		}
	}

	// finally, return home, to the scanned items list
	http.Redirect(w, r, successTarget, http.StatusFound)
}

// 删除学生的提交状态
func deleteStudentRecord(db *sqlite3.Conn, StuId int64) (bool, error) {
	records, recordsErr := database.GetStudentRecord(db, StuId)
	result := false
	if recordsErr == nil {
		for _, r := range records {
			r.DeleteRecord(db)
			result = true
		}
	}
	return result, recordsErr
}

func renderSubmissionListTemplate(w http.ResponseWriter, p *SubmissionPage) {
	if TEMPLATES_INITIALIZED {
		SUBMISSION_LIST_TEMPLATES.Execute(w, p)
	}
}

func renderSubmissionEditTemplate(w http.ResponseWriter, f *StudentForm) {
	if TEMPLATES_INITIALIZED {
		SUBMISSION_EDIT_TEMPLATES.Execute(w, f)
	}
}

func InitializeTemplates(folder string) {
	SUBMISSION_LIST_TEMPLATES = template.Must(template.ParseFiles(TEMPLATE_LIST(folder, SUBMISSION_LIST_TEMPLATE_FILES)...))
	SUBMISSION_EDIT_TEMPLATES = template.Must(template.ParseFiles(TEMPLATE_LIST(folder, SUBMISSION_EDIT_TEMPLATE_FILES)...))
	TEMPLATES_INITIALIZED = true
}

func CreatedStudents(w http.ResponseWriter, r *http.Request, db database.ConnCoordinates, opts ...interface{}) {
	getStudents(w, r, db)
}

// 删除一个学生
func DeleteStudent(w http.ResponseWriter, r *http.Request, dbCoords database.ConnCoordinates, opts ...interface{}) {
	del := func(i *database.Students, db *sqlite3.Conn) {
		i.DeleteStudent(db)
	}
	processStudents(w, r, dbCoords, del, "/")
}

// 删除一个学生的提交状态
func RemoveSingleSubmission(r *http.Request, dbCoords database.ConnCoordinates, opts ...interface{}) string {
	// prepare the ajax reply object
	ack := AjaxAck{Message: "", Error: ""}

	// attempt to connect to the db
	db, err := database.InitializeDB(dbCoords)
	if err != nil {
		ack.Error = err.Error()
	}
	defer db.Close()

	if err == nil {
		// find the specific Submission to remove
		// get the item id from the POST values
		if "POST" == r.Method {
			r.ParseForm()
			if idVal, exists := r.PostForm["id"]; exists {
				if len(idVal) > 0 {
					id, idErr := strconv.ParseInt(idVal[0], 10, 64)
					if idErr != nil {
						ack.Error = idErr.Error()
					} else {
						deleteSuccess, deleteErr := deleteStudentRecord(db, id)
						if deleteSuccess {
							ack.Message = "Ok"
						} else {
							if deleteErr != nil {
								ack.Error = deleteErr.Error()
							} else {
								ack.Error = "No such item"
							}
						}
					}
				} else {
					ack.Error = "Missing item id"
				}
			} else {
				ack.Error = BAD_POST
			}
		} else {
			ack.Error = BAD_REQUEST
		}
	}

	// convert the ajax reply object to json
	ackObj, ackObjErr := json.Marshal(ack)
	if ackObjErr != nil {
		return ackObjErr.Error()
	}
	return string(ackObj)
}
