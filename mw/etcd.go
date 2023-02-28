package mw

import (
	"context"
	"wecom/zaplog"

	"code.in.wezhuiyi.com/nlp-platform/service-tools/golang/pkg/util/mw"
	etcd "go.etcd.io/etcd/v3/clientv3"
	"go.etcd.io/etcd/v3/mvcc/mvccpb"
)

var EC *mw.EtcdClient

func InitETCD(auth etcd.Config) {
	ETCDClient := mw.NewEtcdClient(auth)
	ETCDClient.MustConnect()
	EC = ETCDClient
}

type ETCDClient struct {
	Auth   etcd.Config
	Client *etcd.Client
}

func (x *ETCDClient) TryConnect() error {
	client, err := etcd.New(x.Auth)
	if err != nil {
		return err
	}
	x.Client = client
	return nil
}

func (x *ETCDClient) MustConnect() {
	client, err := etcd.New(x.Auth)
	if err != nil {
		zaplog.Panic(err.Error())
	}
	x.Client = client
}

func (x *ETCDClient) Put(key, value string) error {
	_, err := x.Client.Put(context.Background(), key, value)
	return err
}

// 超过ttl秒后，key 会被删除
func (x *ETCDClient) PutWithTTL(key, value string, ttl int64) bool {
	grant, err := x.Client.Grant(context.TODO(), ttl)
	if err != nil {
		zaplog.Error(err.Error())
		return false
	}
	_, err = x.Client.Put(context.Background(), key, value, etcd.WithLease(grant.ID))
	if err != nil {
		zaplog.Error(err.Error())
		return false
	}
	return true
}

func (x *ETCDClient) Get(key string) *mvccpb.KeyValue {
	rsp, err := x.Client.Get(context.Background(), key)
	if err != nil {
		zaplog.Panic(err.Error())
	}
	if rsp == nil || len(rsp.Kvs) == 0 {
		return nil
	}
	return rsp.Kvs[0]
}

func (x *ETCDClient) GetWithRange(fromKey, endKey string) []*mvccpb.KeyValue {
	return x.GetWithOps(fromKey, etcd.WithRange(endKey))
}

func (x *ETCDClient) GetWithFromKey(fromKey string) []*mvccpb.KeyValue {
	return x.GetWithOps(fromKey, etcd.WithFromKey())
}

func (x *ETCDClient) GetWithPrefix(keyPrefix string) []*mvccpb.KeyValue {
	return x.GetWithOps(keyPrefix, etcd.WithPrefix())
}

func (x *ETCDClient) GetWithOps(key string, opts ...etcd.OpOption) []*mvccpb.KeyValue {
	rsp, err := x.Client.Get(context.Background(), key, opts...)
	if err != nil {
		zaplog.Panic(err.Error())
	}
	if rsp != nil {
		return rsp.Kvs
	}
	return []*mvccpb.KeyValue{}
}

// WithPrefix WithFilterPut WithFilterDelete ...
func (x *ETCDClient) Watch(key string, opts ...etcd.OpOption) etcd.WatchChan {
	return x.Client.Watch(context.Background(), key, opts...)
}

func (x *ETCDClient) Delete(key string, opts ...etcd.OpOption) (*etcd.DeleteResponse, error) {
	return x.Client.Delete(context.Background(), key, opts...)
}
