package agent

import (
	"fmt"

	"github.com/opencurve/curve-manager/internal/common"
	"github.com/opencurve/curve-manager/internal/metrics/bsmetric"
	bsmetricomm "github.com/opencurve/curve-manager/internal/metrics/common"
	bsrpc "github.com/opencurve/curve-manager/internal/rpc/curvebs"
)

// get chunkservers of server concurrently
func listChunkServer(pools *[]Pool, size int) error {
	ret := make(chan common.QueryResult, size)
	count := 0
	for pIndex, pool := range *pools {
		for zIndex, zone := range pool.Zones {
			for sIndex, server := range zone.Servers {
				go func(id uint32, addr *Server) {
					chunkservers, err := bsrpc.GMdsClient.ListChunkServer(id)
					ret <- common.QueryResult{
						Key:    addr,
						Result: chunkservers,
						Err:    err,
					}
				}(server.Id, &(*pools)[pIndex].Zones[zIndex].Servers[sIndex])
			}
		}
	}
	for res := range ret {
		if res.Err != nil {
			return res.Err
		}
		for _, cs := range res.Result.([]bsrpc.ChunkServer) {
			res.Key.(*Server).ChunkServers = append(res.Key.(*Server).ChunkServers, cs)
		}
		count += 1
		if count >= size {
			break
		}
	}
	return nil
}

// get servers of zones concurrently
func listZoneServer(pools *[]Pool, size int) error {
	ret := make(chan common.QueryResult, size)
	count := 0
	number := 0
	for pIndex, pool := range *pools {
		for zIndex, zone := range pool.Zones {
			go func(id uint32, addr *Zone) {
				servers, err := bsrpc.GMdsClient.ListZoneServer(id)
				ret <- common.QueryResult{
					Key:    addr,
					Result: servers,
					Err:    err,
				}
			}(zone.Id, &(*pools)[pIndex].Zones[zIndex])
		}
	}
	for res := range ret {
		if res.Err != nil {
			return res.Err
		}
		for _, s := range res.Result.([]bsrpc.Server) {
			var server Server
			server.Id = s.Id
			server.Hostname = s.HostName
			server.InternalIp = s.InternalIp
			server.InternalPort = s.InternalPort
			server.ExternalIp = s.ExternalIp
			server.ExternalPort = s.ExternalPort
			res.Key.(*Zone).Servers = append(res.Key.(*Zone).Servers, server)
			number = number + 1
		}
		count += 1
		if count >= size {
			break
		}
	}
	return listChunkServer(pools, number)
}

// get zones of pools concurrently
func listPoolZone(pools *[]Pool) error {
	size := len(*pools)
	ret := make(chan common.QueryResult, size)
	count := 0
	number := 0
	for index, pool := range *pools {
		go func(id uint32, addr *Pool) {
			zones, err := bsrpc.GMdsClient.ListPoolZone(id)
			ret <- common.QueryResult{
				Key:    addr,
				Result: zones,
				Err:    err,
			}
		}(pool.physicalPoolId, &(*pools)[index])
	}
	for res := range ret {
		if res.Err != nil {
			return res.Err
		}
		for _, z := range res.Result.([]bsrpc.Zone) {
			var zone Zone
			zone.Id = z.Id
			zone.Name = z.Name
			res.Key.(*Pool).Zones = append(res.Key.(*Pool).Zones, zone)
			number = number + 1
		}
		count += 1
		if count >= size {
			break
		}
	}
	return listZoneServer(pools, number)
}

func getPoolSpace(pools *[]PoolInfo) error {
	// get can be recycled space
	_, recycledSize, err := bsrpc.GMdsClient.GetFileAllocatedSize(RECYCLEBIN_DIR)
	if err != nil {
		return err
	}

	// get capacity and used space
	size := len(*pools)
	ret := make(chan common.QueryResult, size)
	for index, pool := range *pools {
		go func(name string, index int) {
			space, err := bsmetric.GetPoolSpace(name)
			ret <- common.QueryResult{
				Key:    index,
				Result: space,
				Err:    err,
			}
		}(pool.Name, index)
	}
	count := 0
	for res := range ret {
		if res.Err != nil {
			return res.Err
		}
		(*pools)[res.Key.(int)].Space.Total = res.Result.(*bsmetric.Space).Total
		(*pools)[res.Key.(int)].Space.Alloc = res.Result.(*bsmetric.Space).Used
		(*pools)[res.Key.(int)].Space.CanRecycled = (*recycledSize)[(*pools)[res.Key.(int)].Id]
		count += 1
		if count >= size {
			break
		}
	}
	return nil
}

func getPoolItemNum(pools *[]PoolInfo) error {
	size := len(*pools)
	ret := make(chan common.QueryResult, size)
	for index, pool := range *pools {
		go func(name string, index int) {
			number, err := bsmetric.GetPoolItemNum(name)
			ret <- common.QueryResult{
				Key:    index,
				Result: number,
				Err:    err,
			}
		}(pool.Name, index)
	}
	count := 0
	for res := range ret {
		if res.Err != nil {
			return res.Err
		}
		(*pools)[res.Key.(int)].ServerNum = res.Result.(*bsmetric.PoolItemNum).ServerNum
		(*pools)[res.Key.(int)].ChunkServerNum = res.Result.(*bsmetric.PoolItemNum).ChunkServerNum
		(*pools)[res.Key.(int)].CopysetNum = res.Result.(*bsmetric.PoolItemNum).CopysetNum
		count += 1
		if count >= size {
			break
		}
	}
	return nil
}

func getPoolPerformance(pools *[]PoolInfo) error {
	size := len(*pools)
	ret := make(chan common.QueryResult, size)
	for index, pool := range *pools {
		go func(name string, index int) {
			performance, err := bsmetric.GetPoolPerformance(name)
			ret <- common.QueryResult{
				Key:    index,
				Result: performance,
				Err:    err,
			}
		}(pool.Name, index)
	}
	count := 0
	for res := range ret {
		if res.Err != nil {
			return res.Err
		}
		(*pools)[res.Key.(int)].Performance = append((*pools)[res.Key.(int)].Performance, res.Result.([]bsmetricomm.Performance)...)
		count += 1
		if count >= size {
			break
		}
	}
	return nil
}

func ListLogicalPool() (interface{}, error) {
	result := []PoolInfo{}
	// get info from mds
	pools, err := bsrpc.GMdsClient.ListLogicalPool()
	if err != nil {
		return nil, fmt.Errorf("ListLogicalPool failed, %s", err)
	}

	for _, pool := range pools {
		var info PoolInfo
		info.Id = pool.Id
		info.Name = pool.Name
		info.PhysicalPoolId = pool.PhysicalPoolId
		info.Type = pool.Type
		info.CreateTime = pool.CreateTime
		info.AllocateStatus = pool.AllocateStatus
		info.ScanEnable = pool.ScanEnable
		result = append(result, info)
	}

	// get info from monitor
	err = getPoolItemNum(&result)
	if err != nil {
		return nil, err
	}

	err = getPoolSpace(&result)
	if err != nil {
		return nil, err
	}

	err = getPoolPerformance(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func ListTopology() (interface{}, error) {
	result := []Pool{}
	logicalPools, err := bsrpc.GMdsClient.ListLogicalPool()
	if err != nil {
		return nil, fmt.Errorf("ListLogicalPool failed, %s", err)
	}
	for _, lp := range logicalPools {
		var pool Pool
		pool.Id = lp.Id
		pool.physicalPoolId = lp.PhysicalPoolId
		pool.Name = lp.Name
		pool.Type = lp.Type
		result = append(result, pool)
	}
	err = listPoolZone(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}