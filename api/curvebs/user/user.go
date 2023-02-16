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

package user

import (
	"github.com/mcuadros/go-defaults"
	"github.com/opencurve/curve-manager/api/curvebs/agent"
	"github.com/opencurve/curve-manager/api/curvebs/core"
	"github.com/opencurve/curve-manager/internal/errno"
	"github.com/opencurve/pigeon"
)

func Login(r *pigeon.Request, ctx *Context) bool {
	data := ctx.Data.(*LoginRequest)
	userInfo, err := agent.Login(r, data.UserName, data.PassWord)
	if err != errno.OK {
		return core.Exit(r, err)
	}
	return core.ExitSuccessWithData(r, userInfo)
}

func CreateUser(r *pigeon.Request, ctx *Context) bool {
	data := ctx.Data.(*CreateUserRequest)
	defaults.SetDefaults(data)
	err := agent.CreateUser(r, data.UserName, data.PassWord, data.Email, data.Permission)
	return core.Exit(r, err)
}

func DeleteUser(r *pigeon.Request, ctx *Context) bool {
	data := ctx.Data.(*DeleteUserRequest)
	err := agent.DeleteUser(r, data.UserName)
	return core.Exit(r, err)
}

func ChangePassWord(r *pigeon.Request, ctx *Context) bool {
	data := ctx.Data.(*ChangePassWordRequest)
	err := agent.ChangePassWord(r, data.UserName, data.OldPassWord, data.NewPassWord)
	return core.Exit(r, err)
}

func ResetPassWord(r *pigeon.Request, ctx *Context) bool {
	data := ctx.Data.(*ResetPassWordRequest)
	err := agent.ResetPassWord(r, data.UserName)
	return core.Exit(r, err)
}

func UpdateUserInfo(r *pigeon.Request, ctx *Context) bool {
	data := ctx.Data.(*UpdateUserInfoRequest)
	err := agent.UpdateUserInfo(r, data.UserName, data.Email, data.Permission)
	return core.Exit(r, err)
}

func ListUser(r *pigeon.Request, ctx *Context) bool {
	users, err := agent.ListUser(r)
	if err != errno.OK {
		return core.Exit(r, err)
	}
	return core.ExitSuccessWithData(r, users)
}
