package database

import (
	"fmt"
	"github.com/mxk/go-sqlite/sqlite3"
	"io/ioutil"
	"math"
	"path"
	"strconv"
	"strings"
	"time"
)

//TODO: CONST
const (

	// Default database filename
	SQLITE_PATH = "/home/pi/data"
	SQLITE_FILE = "PiScanStudentDB.sqlite"

	// Default sql definitions file
	TABLE_SQL_DEFINITIONS = "tables.sql"


	BAD_PK = -1
	// 学生相关操作
	ADD_STUDENT = "INSERT INTO Students (Name, StuId) values ($n, $i)"
	UPDATE_STUDENT="UPDATE Students SET Name = $n, StuId = $i, Submitted: $s WHERE id = $d"
	DELETE_STUDENT="DELETE FROM Students WHERE id = $i"
	SEARCH_EXISTING_STUDENT="SELECT id FROM Students WHERE StuId = $i"
	SEARCH_ALL_STUDENTS="SELECT * FROM Students"

	ADD_RECORD="INSERT INTO SubRecords (StuId, SubTime) values ($i,$t)"
	UPDATE_RECORD="UPDATE SubRecords SET StuId = $i, SubTime = $t WHERE id = $d"
	DELETE_RECORD="DELETE FROM SubRecords WHERE id = $i"
	SEARCH_EXISTING_RECORD="SELECT id FROM SubRecords WHERE StuId = $i"
	SEARCH_ALL_RECORDS

	GET_RECORDS = "SELECT StuId, SubTime FROM SubRecords WHERE SubId = $i"
	GET_STUDENT_RECORDS="SELECT SubId, SubTime FROM SubRecords WHERE StuId = $s"
	GET_ALL_SUBMISSIONS="SELECT * FROM Submissions"

	ADD_SUBMISSION="INSERT INTO Submissions (Description, SubId, SubCreationTime, SubDeadline) values" +
		" ($d, $s, $c, $l)"
	UPDATE_SUBMISSION="UPDATE Submissions SET Description = $d, SubDeadline = $l WHERE SubId = $i"
	DELETE_SUBMISSION="DELETE FROM Submissions WHERE SubId = $i"
	SEARCH_EXISTING_SUBMISSION="SELECT SubId, Description, SubDeadline, SubCreationTime FROM Submissions WHERE SubId = $i"

)

//TODO: VAR
var (
	INTERVALS   = []string{"year", "month", "day", "hour", "minute"}
	SECONDS_PER = map[string]int64{"minute": 60, "hour": 3600, "day": 86400, "month": 2592000, "year": 31536000}
)

// 数据库坐标
type ConnCoordinates struct {
	DBPath string
	DBFile string
	DBTablesPath string
}

////作业上交集群条目
//type Submissions struct {
//	Description string
//	SubId int64
//	SubCreationTime int64
//	SubDeadline int64
//}

// 作业上交条目
type SubRecords struct {
	StuId int64
	SubTime int64
}

// 学生
type Students struct {
	id int64
	Name string
	StuId int64
	Submitted int64
}

func calculateTimeSince(posted string) string {
	result := "just now" // default reply

	// try to convert the posted string into unix time
	i, err := strconv.ParseInt(posted, 10, 64)
	if err == nil {
		tm := time.Unix(i, 0)

		// calculate the time since posted
		// and return a human readable
		// '[interval] ago' string
		duration := time.Since(tm)
		if duration.Seconds() < 60.0 {
			if duration.Seconds() == 1.0 {
				result = fmt.Sprintf("%2.0f second ago", duration.Seconds())
			} else {
				result = fmt.Sprintf("%2.0f seconds ago", duration.Seconds())
			}
		} else {
			for _, interval := range INTERVALS {
				v := math.Trunc(duration.Seconds() / float64(SECONDS_PER[interval]))
				if v > 0.0 {
					if v == 1.0 {
						result = fmt.Sprintf("%2.0f %s ago", v, interval)
					} else {
						// plularize the interval label
						result = fmt.Sprintf("%2.0f %ss ago", v, interval)
					}
					break
				}
			}
		}
	}

	return result
}


func getPK(db *sqlite3.Conn, table string) int64 {
	// find and return the most recently-inserted
	// primary key, based on the table name
	sql := fmt.Sprintf("select seq from sqlite_sequence where name='%s'", table)

	var rowid int64
	for s, err := db.Query(sql); err == nil; err = s.Next() {
		s.Scan(&rowid)
	}
	return rowid
}

// 查询数据库中学生
// 输入值: 数据库; 学生StuId
// 返回值: 未找到为 -1，找到则返回该学生的rowid
func SearchExistingStudent(db *sqlite3.Conn, StuId int64) int64 {
	args := sqlite3.NamedArgs{"$i": StuId} // 用于Sqlite查询语句
	var rowid int64
	rowid = BAD_PK
	for s, err := db.Query(SEARCH_EXISTING_STUDENT, args); err == nil; err = s.Next() {
		s.Scan(&rowid)
	}
	return rowid
}

func SearchAllStudents(db *sqlite3.Conn) ([]*Students, error) {
	results := make([]*Students, 0)
	args := sqlite3.NamedArgs{}
	row := make(sqlite3.RowMap)
	for s, err := db.Query(SEARCH_ALL_STUDENTS, args); err == nil; err = s.Next() {
		var rowid int64
		s.Scan(&rowid,row)
		Name, NameFound := row["Name"]
		StuId, StuIdFound := row["StuId"]
		Submitted, SubmittedFound := row["Submitted"]
		if StuIdFound {
			result := new(Students)
			result.id = rowid
			result.StuId = StuId.(int64)
			if NameFound {
				result.Name = Name.(string)
			}
			if SubmittedFound {
				result.Submitted = Submitted.(int64)
			}
			results = append(results, result)
		}
	}
	return results, nil
}

// 向数据库中添加学生，指向学生
func (s *Students) AddStudent(db *sqlite3.Conn) (int64, error) {
	rowid := SearchExistingStudent(db, s.StuId)
	if rowid == BAD_PK {
		args := sqlite3.NamedArgs{"$n": s.Name, "$i": s.StuId}
		result := db.Exec(ADD_STUDENT, args)
		if result == nil {
			pk := getPK(db, "Students")
			return pk, result
		} else {
			return BAD_PK, result
		}
	} else {
		return SearchExistingStudent(db, s.StuId), nil
	}
}

// 修改学生信息
func (s *Students) UpdateStudent(db *sqlite3.Conn) error {
	args := sqlite3.NamedArgs{"$n": s.Name, "$i": s.StuId, "$d":s.id, "$s":s.Submitted}
	return db.Exec(UPDATE_STUDENT, args)
}

// 删除学生信息记录
func (s *Students) DeleteStudent(db *sqlite3.Conn) error {
	// delete the student
	rowid := SearchExistingStudent(db, s.StuId)
	args := sqlite3.NamedArgs{"$i":rowid}
	return db.Exec(DELETE_STUDENT, args)
}

func SearchExistingRecord(db *sqlite3.Conn, StuId int64) int64 {
	args := sqlite3.NamedArgs{"$i":StuId}
	var rowid int64
	rowid = BAD_PK
	for s, err := db.Query(SEARCH_EXISTING_RECORD, args); err == nil; err = s.Next() {
		s.Scan(&rowid)
	}
	return rowid
}

func (r *SubRecords) AddRecord(db *sqlite3.Conn) (int64, error) {
	if SearchExistingRecord(db, r.StuId) == BAD_PK {
		args := sqlite3.NamedArgs{"$i":r.StuId, "$t":r.SubTime}
		result := db.Exec(ADD_RECORD, args)
		if result == nil {
			pk := getPK(db, "SubRecords")
			return pk, result
		} else {
			return BAD_PK, result
		}
	} else {
		return SearchExistingRecord(db, r.StuId), nil
	}
}

func (r *SubRecords) DeleteRecord(db *sqlite3.Conn) error {
	rowid := SearchExistingRecord(db, r.StuId)
	args := sqlite3.NamedArgs{"$i":rowid}
	return db.Exec(DELETE_RECORD, args)
}

func (r *SubRecords) UpdateRecord(db *sqlite3.Conn) error {
	rowid := SearchExistingRecord(db, r.StuId)
	args := sqlite3.NamedArgs{"$i":r.StuId,"$t":r.SubTime,"$d":rowid}
	return db.Exec(UPDATE_RECORD, args)
}

//func fetchSubmissionRecords(db *sqlite3.Conn, i *Submissions, sql string) ([]*Students, error) {
//	// find all the items for this account
//	results := make([]*Students, 0)
//
//	args := sqlite3.NamedArgs{"$i": i.SubId}
//	row := make(sqlite3.RowMap)
//	for s, err := db.Query(sql, args); err == nil; err = s.Next() {
//		s.Scan(row)
//		StuId, StuIdFound := row["StuId"]
//		if StuIdFound {
//			result := new(Students)
//			result.StuId = StuId.(int64)
//			results = append(results, result)
//		}
//	}
//	return results, nil
//}

//func GetRecordStudents(db *sqlite3.Conn, i *Submissions) ([]*Students, error) {
//	return fetchSubmissionRecords(db, i, GET_RECORDS)
//}

func GetStudentRecord(db *sqlite3.Conn, StuId int64) ([]*SubRecords, error) {

	results := make([]*SubRecords, 0)
	args := sqlite3.NamedArgs{"$i": StuId}
	row := make(sqlite3.RowMap)
	for s, err := db.Query("SELECT * FROM SubRecords WHERE StuId = $i", args); err == nil; err = s.Next() {
		s.Scan(row)
		StuId, StuIdFound := row["StuId"]
		SubTime, SubTimeFound := row["SubTime"]
		if StuIdFound {
			result := new(SubRecords)
			result.StuId = StuId.(int64)
			if SubTimeFound {
				result.SubTime = SubTime.(int64)
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func GetSingleStudent(db *sqlite3.Conn, StuId int64) (*Students, error) {
	student := new(Students)
	student.StuId = BAD_PK
	students, err := SearchAllStudents(db)
	for _, i := range students {
		if i.StuId == StuId {
			return i, err
		}
	}
	return student, err
}

//func (s *Submissions) AddSubmission(db *sqlite3.Conn) error {
//	args := sqlite3.NamedArgs{"$d": s.Description, "$s": s.SubId, "$c": s.SubCreationTime, "$l": s.SubDeadline}
//	return db.Exec(ADD_SUBMISSION, args)
//}
//
//// 修改学生信息
//func (s *Submissions) UpdateSubmissions(db *sqlite3.Conn, Description string, SubDeadline int64) error {
//	// update the student with with user contribution (description)
//	args := sqlite3.NamedArgs{"$i":s.SubId, "$d": Description, "$l": SubDeadline}
//	return db.Exec(UPDATE_SUBMISSION, args)
//}
//
//// 删除学生信息记录
//func (s *Submissions) DeleteSubmissions(db *sqlite3.Conn) error {
//	// delete the student
//	args := sqlite3.NamedArgs{"$i": s.SubId}
//	return db.Exec(DELETE_SUBMISSION, args)
//}
//
//func GetSubmission (db *sqlite3.Conn, SubId int64) (*Submissions, error) { //SubId, Description, SubDeadline, SubCreationTime
//	args := sqlite3.NamedArgs{"$i":SubId}
//	result := new(Submissions)
//	var rowid int64
//	rowid = BAD_PK
//	row := make(sqlite3.RowMap)
//	for s, err := db.Query(SEARCH_EXISTING_SUBMISSION, args); err == nil; err = s.Next() {
//		s.Scan(&rowid, row)
//	}
//	if rowid != BAD_PK {
//		result.SubId = rowid
//		Description, DescriptionFound := row["Description"]
//		SubDeadline, SubDeadlineFound := row["SubDeadline"]
//		SubCreationTime, SubCreationTimeFound := row["SubCreationTime"]
//		if DescriptionFound {
//			result.Description = Description.(string)
//		}
//		if SubDeadlineFound {
//			result.SubDeadline = SubDeadline.(int64)
//		}
//		if SubCreationTimeFound {
//			result.SubCreationTime = SubCreationTime.(int64)
//		}
//	}
//	return result, nil
//}
//
//
//func GetAllSubmissions(db *sqlite3.Conn) ([]*SubRecords, error) {
//	results := make([]*Submissions, 0)
//	args := sqlite3.NamedArgs{}
//	row := make(sqlite3.RowMap)
//	for s, err := db.Query(GET_STUDENT_RECORDS, args); err == nil; err = s.Next() {
//		s.Scan(row)
//		SubId, SubIdFound := row["SubId"]
//		Description, DescriptionFound := row["Description"]
//		SubDeadline, SubDeadlineFound := row["SubDeadline"]
//		SubCreationTime, SubCreationTimeFound := row["SubCreationTime"]
//		if SubIdFound {
//			result := new(Submissions)
//			result.SubId = SubId.(int64)
//			if DescriptionFound {
//				result.Description = Description.(string)
//			}
//			if SubDeadlineFound {
//				result.SubDeadline = SubDeadline.(int64)
//			}
//			if SubCreationTimeFound {
//				result.SubCreationTime = SubCreationTime.(int64)
//			}
//			results = append(results, result)
//		}
//	}
//	return results, nil
//}
//
func InitializeDB(coords ConnCoordinates) (*sqlite3.Conn, error) {
	// attempt to open the sqlite db file
	db, dbErr := sqlite3.Open(path.Join(coords.DBPath, coords.DBFile))
	if dbErr != nil {
		return db, dbErr
	}

	// load the table definitions file, if coords.DBTablesPath is defined
	if len(coords.DBTablesPath) > 0 {
		content, err := ioutil.ReadFile(path.Join(coords.DBTablesPath, TABLE_SQL_DEFINITIONS))
		if err != nil {
			return db, err
		}

		// attempt to create (if not exists) each table
		tables := strings.Split(string(content), ";")
		for _, table := range tables {
			err = db.Exec(table)
			if err != nil {
				return db, err
			}
		}
	}

	return db, nil
}
