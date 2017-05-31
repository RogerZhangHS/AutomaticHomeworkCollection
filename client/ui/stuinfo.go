// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package ui provides http request handlers for the Pi client WebApp

package ui

import (
	"PiScanStudent/client/database"
//	"PiScanStudent/server/digest"
	"net/http"
//	"net/url"
	"strconv"
	"strings"
)

// InputUnknownStudent handles the form for user contributions of unknown
// barcode scans: a GET presents the form, and a POST responds to the
// user-contributed input
func InputUnknownStudent(w http.ResponseWriter, r *http.Request, dbCoords database.ConnCoordinates, opts ...interface{}) {
	// attempt to connect to the db
	db, err := database.InitializeDB(dbCoords)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

//	// get the Account for this request
//	acc, accErr := database.GetDesignatedAccount(db)
//	if accErr != nil {
//		http.Error(w, accErr.Error(), http.StatusInternalServerError)
//		return
//	}

	// prepare the html page response
	form := &StudentForm{Title: "Set Student Information",
		CancelUrl:    HOME_URL }

	//lookup the item from the request id
	// and show the input form (if a GET)
	// or process it (if a POST)
	if "GET" == r.Method {
		// derive the item id from the url path
		urlPaths := strings.Split(r.URL.Path[1:], "/")
		if len(urlPaths) >= 2 {
			StuId, StuIdErr := strconv.ParseInt(urlPaths[1], 10, 64)
			if StuIdErr == nil {
				student, studentErr := database.GetSingleStudent(db, StuId)
				if studentErr == nil {
					if student.StuId != database.BAD_PK {
						// requested student has been found and is valid
						form.Student = student
					}
				}
			}
		}

		if form.Student == nil {
			// no matching item was found
			http.Error(w, BAD_REQUEST, http.StatusInternalServerError)
			return
		}

	} else if "POST" == r.Method {
		// get the item id from the posted data
		r.ParseForm()
		idVal, idExists := r.PostForm["id"]
		stuIdVal, stuIdExists := r.PostForm["StuId"]
		stuNameVal, stuNameExists := r.PostForm["stuName"]
		if idExists && stuIdExists && stuNameExists {
			id, idErr := strconv.ParseInt(idVal[0], 10, 64)
			if idErr != nil {
				form.FormError = idErr.Error()
			} else {
				student, studentErr := database.GetSingleStudent(db, id)
				if studentErr != nil {
					form.FormError = studentErr.Error()
				} else {
					// the hidden barcode value must match the retrieved student
					stuId, _ := strconv.ParseInt(stuIdVal[0],10,64)
					if student.StuId == stuId {
						// update the student in the local client db
						student.Name = stuNameVal[0]
						student.UpdateStudent(db)
						// return success
						http.Redirect(w, r, HOME_URL, http.StatusFound)
						return
					} else {
						// bad form post: the hidden barcode value does not match the retrieved student
						form.FormError = BAD_POST
					}
				}
			}
		} else {
			// required form parameters are missing
			form.FormError = BAD_POST
		}
	}

	renderSubmissionEditTemplate(w, form)
}
