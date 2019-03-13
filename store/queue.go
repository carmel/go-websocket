package util

type Queue interface {
	Offer(e interface{}) //向队列中添加元素
	Poll() interface{}   //移除队列中最前面的元素
	Clear() bool         //清空队列
	Size() int           //获取队列的元素个数
	IsEmpty() bool       //判断队列是否是空
}

type RedisQueue struct{}

func (q RedisQueue) Offer(e interface{}) {

}
