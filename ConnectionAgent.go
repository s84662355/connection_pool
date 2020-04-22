package connection_pool

import (
	"fmt"
	"time"
)

/**
 * Class ConnectionAgentStruct
 *
 * @since 2.0
 */
type ConnectionAgentStruct struct {

	/**
	 * @var int
	 */
	id int

	/**
	 * @var AgentPoolStruct
	 */
	pool *AgentPoolStruct

	/**
	 * @var long
	 */
	lastTime int64

	/**
	 * @var string
	 */
	poolname string

	/**
	 * @var ConnectionInterface
	 */
	connection ConnectionInterface
}

/**
 * @return int
 */
func (l *ConnectionAgentStruct) GetLastTime() int64 {
	return l.lastTime
}

/**
 * Update last time
 */
func (l *ConnectionAgentStruct) UpdateLastTime() {

	l.lastTime = time.Now().Unix()
}

/**
 * Release Connection
 * 返回连接池
 */
func (l *ConnectionAgentStruct) Release() {
	l.pool.Release(l)
}

/**
 * Close Connection
 * 关闭连接
 */
func (l *ConnectionAgentStruct) Close() {
	l.connection.Close()
}

/**
 * Create Connection
 * 创建连接
 */
func (l *ConnectionAgentStruct) Create() {
	l.connection.Create()
}

/**
 * Reconnect Connection
 * 重新连接
 */
func (l *ConnectionAgentStruct) Reconnect() bool {
	return l.connection.Reconnect()
}

/**
 * Target
 * 获取连接接口
 */
func (l *ConnectionAgentStruct) Target() ConnectionInterface {
	return l.connection
}

func (l *ConnectionAgentStruct) New(params map[string]interface{}) *ConnectionAgentStruct {
	Connection := new(ConnectionAgentStruct)

	id, ok := params["id"]
	if ok {
		Connection.id, _ = id.(int)
	}

	poolname, ok := params["poolname"]
	if ok {
		Connection.poolname, _ = poolname.(string)
	}

	connection, _ := params["connection"]
	Connection.connection, _ = connection.(ConnectionInterface)

	pool, _ := params["pool"]
	Connection.pool, _ = pool.(*AgentPoolStruct)

	fmt.Println("创建连接")

	return Connection
}

var ConnectionAgent = new(ConnectionAgentStruct)

type ConnectionInterface interface {
	Close()
	Create()
	Reconnect() bool
}
