package connection_pool

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

/**
 * Class AgentPoolStruct
 * 连接池
 * @since 2.0
 */
type AgentPoolStruct struct {

	/**
	 * Minimum active number of connections
	 * 连接池最少连接数
	 * @var int
	 */
	minActive int

	/**
	 * Maximum active number of connections
	 * 连接池最大连接数
	 * @var int
	 */
	maxActive int

	/**
	 * Maximum waiting time(second), if there is not limit to 0
	 * 获取一个连接的最大等待时间
	 * 单位秒
	 * @var int64
	 */
	maxWaitTime int64

	/**
	 * Maximum idle time(second)
	 * 一个连接最长闲置时间
	 * 单位秒
	 * @var int64
	 */
	maxIdleTime int64

	/**
	 * Maximum wait close time
	 * 连接池关闭时间
	 * 单位秒
	 * @var int64
	 */
	maxCloseTime int64

	/**
	 *
	 * 存放连接的列表
	 * @var list_queue list.List
	 *
	 *
	 */
	list_queue *list.List

	/**
	 *
	 * 已经创建的连接数量
	 * Current count
	 *
	 * @var int
	 */
	count int

	lock sync.Mutex

	factory PoolInterface
}

/**
 *
 * 初始化连接池
 * 不能并发执行
 */
func (l *AgentPoolStruct) InitPool() {
	for i := 0; i < l.minActive; i++ {
		c := l.create()
		if c != nil {
			l.Release(c)
		}

	}

}

/**
 * @return *ConnectionAgentStruct
 * 获取连接
 */
func (l *AgentPoolStruct) GetConnection() *ConnectionAgentStruct {

	passtime := time.Now().Unix() + l.maxWaitTime
	var connection *ConnectionAgentStruct = nil

	for passtime-time.Now().Unix() >= 0 {
		connection = l.getConnectionByChannel()

		if connection != nil {
			connection.UpdateLastTime()
			fmt.Println("获取连接成功")
			return connection
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	fmt.Println("获取失败")
	return connection
}

/**
 * @param *ConnectionAgentStruct connection
 * 返回连接池
 */
func (l *AgentPoolStruct) Release(connection *ConnectionAgentStruct) {
	l.lock.Lock()
	defer func() {
		l.lock.Unlock()
	}()
	fmt.Println("归队")
	l.list_queue.PushBack(connection)
}

/**
 * Get connection by channel
 *
 * @return *ConnectionAgentStruct
 * 获取连接
 */
func (l *AgentPoolStruct) getConnectionByChannel() *ConnectionAgentStruct {
	l.lock.Lock()

	defer func() {
		l.lock.Unlock()
		if err := recover(); err != nil {
			panic(fmt.Sprintf("连接池获取连接失败 error", err))
		}
	}()

	var connection *ConnectionAgentStruct = nil
	if l.list_queue.Len() > 0 {

		fmt.Println("在队列获取连接")

		element := l.list_queue.Front()

		connection, _ = element.Value.(*ConnectionAgentStruct)
		l.list_queue.Remove(element)

		nowtime := time.Now().Unix()

		lastTime := connection.GetLastTime()

		if nowtime-lastTime > l.maxIdleTime {
			fmt.Println("连接过期")
			connection.Close()
			l.count--
			return nil
		}

		connection.UpdateLastTime()
		return connection
	}

	if l.count < l.maxActive {
		connection = l.create()
		return connection
	}

	return nil
}

/**
 * @return *ConnectionAgentStruct
 * 创建连接
 *
 */
func (l *AgentPoolStruct) create() *ConnectionAgentStruct {
	l.count++
	defer func() {
		if err := recover(); err != nil {
			l.count--
			panic(fmt.Sprintf("Create connection error", err))
		}
	}()
	connection := l.createConnection()

	if connection != nil {
		connection.UpdateLastTime()
	}

	fmt.Println("AbstractPool  Count %d", l.count)

	return connection
}

/**
 * @return int
 * 关闭连接池
 * 不建议并发执行
 */
func (l *AgentPoolStruct) Close() int {
	passtime := time.Now().Unix() + l.maxCloseTime
	var connection *ConnectionAgentStruct = nil

	for passtime-time.Now().Unix() > 0 {
		l.lock.Lock()
		if l.list_queue.Len() > 0 {
			element := l.list_queue.Front()
			connection, _ = element.Value.(*ConnectionAgentStruct)
			l.list_queue.Remove(element)
			connection.Close()
			l.count--
			fmt.Println("关闭连接池", l.count)
		}
		if l.count == 0 {
			return l.count
		}
		l.lock.Unlock()
		time.Sleep(time.Duration(10) * time.Millisecond)
	}
	return l.count
}

/**
 * @return *ConnectionAgentStruct
 * 创建连接
 *
 */
func (l *AgentPoolStruct) createConnection() *ConnectionAgentStruct {
	c := l.factory.CreateConnection(l)
	return c
}

func (l *AgentPoolStruct) New(params map[string]interface{}) *AgentPoolStruct {
	pool := new(AgentPoolStruct)

	pool.minActive = 5
	pool.maxActive = 10
	pool.maxWaitTime = 1
	pool.maxIdleTime = 10
	pool.maxCloseTime = 5

	minActive, ok := params["minActive"]
	if ok {
		pool.minActive, _ = minActive.(int)
	}

	maxActive, ok := params["maxActive"]
	if ok {
		pool.maxActive, _ = maxActive.(int)
	}

	maxWaitTime, ok := params["maxWaitTime"]
	if ok {
		pool.maxWaitTime, _ = maxWaitTime.(int64)
	}

	maxIdleTime, ok := params["maxIdleTime"]
	if ok {
		pool.maxIdleTime, _ = maxIdleTime.(int64)
	}

	maxCloseTime, ok := params["maxCloseTime"]
	if ok {
		pool.maxCloseTime, _ = maxCloseTime.(int64)
	}

	factory, _ := params["factory"]

	pool.factory, _ = factory.(PoolInterface)

	pool.count = 0
	pool.list_queue = list.New()

	return pool
}

var AgentPool = new(AgentPoolStruct)

type PoolInterface interface {
	CreateConnection(pool *AgentPoolStruct) *ConnectionAgentStruct
}
