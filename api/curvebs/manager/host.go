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
* Created Date: 2023-02-15
* Author: wanghai (SeanHai)
 */

package manager

import (
	comm "github.com/opencurve/curve-manager/api/common"
	"github.com/opencurve/curve-manager/api/curvebs/core"
	"github.com/opencurve/curve-manager/internal/agent"
	"github.com/opencurve/curve-manager/internal/errno"
	"github.com/opencurve/pigeon"
)

func ListHost(r *pigeon.Request, ctx *Context) bool {
	data := ctx.Data.(*ListHostRequest)
	hosts, err := agent.ListHost(data.Size, data.Page)
	if err != nil {
		r.Logger().Error("ListHost failed",
			pigeon.Field("error", err),
			pigeon.Field("requestId", r.HeadersIn[comm.HEADER_REQUEST_ID]))
		return core.Exit(r, errno.LIST_HOST_FAILED)
	}
	return core.ExitSuccessWithData(r, hosts)
}

func GetHostPerformance(r *pigeon.Request, ctx *Context) bool {
	data := ctx.Data.(*GetHostPerformanceRequest)
	performance, err := agent.GetHostPerformance(data.HostName)
	if err != nil {
		r.Logger().Error("GetHostPerformance failed",
			pigeon.Field("error", err),
			pigeon.Field("requestId", r.HeadersIn[comm.HEADER_REQUEST_ID]))
		return core.Exit(r, errno.GET_HOST_PERFORMANCE)
	}
	return core.ExitSuccessWithData(r, performance)
}
