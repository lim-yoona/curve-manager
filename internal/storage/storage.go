/*
*  Copyright (c) 2023 NetEase Inc.
*
*  Licensed under the Apache License, Version 2.0 (the "License");
*  you may not use this file except in compliance with the License.
*  You may obtain a copy of the License at
*
*      http://www.apache.org/licenses/LICENSE-2.0
*
*  Unless required by applicable law or agreed to in writing, software
*  distributed under the License is distributed on an "AS IS" BASIS,
*  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*  See the License for the specific language governing permissions and
*  limitations under the License.
 */

/*
* Project: Curve-Manager
* Created Date: 2023-02-11
* Author: wanghai (SeanHai)
 */

package storage

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/opencurve/curve-manager/internal/common"
	"github.com/opencurve/pigeon"
)

var (
	gStorage *Storage
)

const (
	SQLITE_DB_FILE      = "db.sqlite.filepath"
	NEW_PASSWORD_LENGTH = 8
)

type UserInfo struct {
	UserName   string `json:"userName" binding:"required"`
	PassWord   string `json:"-"`
	Email      string `json:"email"`
	Permission int    `json:"permission" binding:"required"`
	Token      string `json:"token" binding:"required"`
}

type Storage struct {
	db           *sql.DB
	mutex        *sync.Mutex
	session      map[string]int64
	sessionMutex *sync.Mutex
}

func Init(cfg *pigeon.Configure) error {
	dbfile := cfg.GetConfig().GetString(SQLITE_DB_FILE)
	if len(dbfile) == 0 {
		return fmt.Errorf("no sqlite db file found")
	}

	// new db
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return err
	}
	gStorage = &Storage{db: db, mutex: &sync.Mutex{}, session: make(map[string]int64), sessionMutex: &sync.Mutex{}}

	// init user table
	if err = gStorage.execSQL(CREATE_USER_TABLE); err != nil {
		return err
	}

	// create admin user
	if err = createAdminUser(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) execSQL(query string, args ...interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(args...)
	return err
}

func createAdminUser() error {
	passwd := common.GetMd5Sum32Little(USER_ADMIN_PASSWORD)
	return gStorage.execSQL(CREATE_ADMIN, USER_ADMIN_NAME, passwd, "", ADMIN_PERM)
}

func AddSession(userInfo *UserInfo) {
	now := time.Now().Unix()
	sigStr := fmt.Sprintf("username=%s&password=%s&timestamp=%d", userInfo.UserName, userInfo.PassWord, now)
	sig := common.GetMd5Sum32Little(sigStr)
	gStorage.sessionMutex.Lock()
	defer gStorage.sessionMutex.Unlock()
	gStorage.session[sig] = now
	userInfo.Token = sig
}

func CheckSession(s string, expireSec int) bool {
	now := time.Now().Unix()
	gStorage.sessionMutex.Lock()
	defer gStorage.sessionMutex.Unlock()
	if time, ok := gStorage.session[s]; ok {
		if time+int64(expireSec) < now {
			delete(gStorage.session, s)
			return false
		}
		gStorage.session[s] = now
		return true
	}
	return false
}

func GetUser(name string) (UserInfo, error) {
	var user UserInfo
	rows, err := gStorage.db.Query(GET_USER, name)
	if err != nil {
		return user, err
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&user.UserName, &user.PassWord, &user.Email, &user.Permission)
		if err != nil {
			return user, err
		}
	} else {
		return user, fmt.Errorf("user not exist")
	}
	return user, nil
}

func SetUser(name, passwd, email string, permission int) error {
	return gStorage.execSQL(CREATE_USER, name, passwd, email, permission)
}

func DeleteUser(name string) error {
	return gStorage.execSQL(DELETE_USER, name)
}

func UpdateUserPassWord(name, passwd string) error {
	return gStorage.execSQL(UPDATE_PASSWORD, passwd, name)
}

func UpdateUserInfo(name, email string, permission int) error {
	return gStorage.execSQL(UPDATE_USER_INFO, email, permission, name)
}

func ListUser() (interface{}, error) {
	rows, err := gStorage.db.Query(LIST_USER, USER_ADMIN_NAME)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []UserInfo
	for rows.Next() {
		var user UserInfo
		err = rows.Scan(&user.UserName, &user.Email, &user.Permission)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return &users, nil
}

func GetUserEmail(name string) (string, error) {
	rows, err := gStorage.db.Query(GET_USER_EMAIL, name)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var email string
	if rows.Next() {
		err = rows.Scan(&email)
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("user not exist")
	}
	return email, nil
}

func GetUserPassword(name string) (string, error) {
	rows, err := gStorage.db.Query(GET_USER_PASSWORD, name)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var passwd string
	if rows.Next() {
		err = rows.Scan(&passwd)
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("user not exist")
	}
	return passwd, nil
}

func GetNewPassWord() string {
	return common.GetRandString(NEW_PASSWORD_LENGTH)
}
