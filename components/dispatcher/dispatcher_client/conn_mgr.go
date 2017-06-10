package dispatcher_client

import (
	"time"

	"sync/atomic"

	"unsafe"

	"errors"

	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
)

const (
	LOOP_DELAY_ON_DISPATCHER_CLIENT_ERROR = time.Second
)

var (
	_dispatcherClient         *DispatcherClient // DO NOT access it directly
	dispatcherClientDelegate  IDispatcherClientDelegate
	errDispatcherNotConnected = errors.New("dispatcher not connected")
)

func getDispatcherClient() *DispatcherClient { // atomic
	addr := (*uintptr)(unsafe.Pointer(&_dispatcherClient))
	return (*DispatcherClient)(unsafe.Pointer(atomic.LoadUintptr(addr)))
}

func setDispatcherClient(dc *DispatcherClient) { // atomic
	addr := (*uintptr)(unsafe.Pointer(&_dispatcherClient))
	atomic.StoreUintptr(addr, uintptr(unsafe.Pointer(dc)))
}

func assureConnectedDispatcherClient() *DispatcherClient {
	var err error
	dispatcherClient := getDispatcherClient()
	//gwlog.Debug("assureConnectedDispatcherClient: _dispatcherClient", _dispatcherClient)
	for dispatcherClient == nil {
		dispatcherClient, err = connectDispatchClient()
		if err != nil {
			gwlog.Error("Connect to dispatcher failed: %s", err.Error())
			time.Sleep(LOOP_DELAY_ON_DISPATCHER_CLIENT_ERROR)
			continue
		}
		setDispatcherClient(dispatcherClient)
		dispatcherClientDelegate.OnDispatcherClientConnect()

		gwlog.Info("dispatcher_client: connected to dispatcher: %s", dispatcherClient)
	}

	return dispatcherClient
}

func connectDispatchClient() (*DispatcherClient, error) {
	dispatcherConfig := config.GetDispatcher()
	conn, err := netutil.ConnectTCP(dispatcherConfig.Ip, dispatcherConfig.Port)
	if err != nil {
		return nil, err
	}
	return newDispatcherClient(conn), nil
}

type IDispatcherClientDelegate interface {
	OnDispatcherClientConnect()
	HandleDispatcherClientPacket(msgtype proto.MsgType_t, packet *netutil.Packet)
	//HandleDeclareService(entityID common.EntityID, serviceName string)
	//HandleCallEntityMethod(entityID common.EntityID, method string, args []interface{})
}

func Initialize(delegate IDispatcherClientDelegate) {
	dispatcherClientDelegate = delegate

	assureConnectedDispatcherClient()
	go netutil.ServeForever(serveDispatcherClient) // start the recv routine
}

func GetDispatcherClientForSend() *DispatcherClient {
	dispatcherClient := getDispatcherClient()
	return dispatcherClient
}

// serve the dispatcher client, receive RESPs from dispatcher and process
func serveDispatcherClient() {
	gwlog.Debug("serveDispatcherClient: start serving dispatcher client ...")
	for {
		dispatcherClient := assureConnectedDispatcherClient()
		var msgtype proto.MsgType_t
		pkt, err := dispatcherClient.Recv(&msgtype)
		if err != nil {
			gwlog.Error("serveDispatcherClient: RecvMsgPacket error: %s", err.Error())
			dispatcherClient.Close()
			setDispatcherClient(nil)
			time.Sleep(LOOP_DELAY_ON_DISPATCHER_CLIENT_ERROR)
			continue
		}

		gwlog.Debug("%s.RecvPacket: msgtype=%v, payload=%v", msgtype, pkt.Payload())
		dispatcherClientDelegate.HandleDispatcherClientPacket(msgtype, pkt)
	}
}