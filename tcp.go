package tcp

import (
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
	"net"
	"time"
)

func init() {
	modules.Register("k6/x/tcp", new(TCP))
}

type (
	// RootModule is the global module instance that will create module
	// instances for each VU.
	RootModule struct{}

	// ModuleInstance represents an instance of the JS module.
	ModuleInstance struct {
		// vu provides methods for accessing internal k6 objects for a VU
		vu modules.VU
		// comparator is the exported type
		tcp *TCP
	}
)

type Socket struct {
	builtinMetrics *metrics.BuiltinMetrics
}

var socket = &Socket{
	builtinMetrics: &metrics.BuiltinMetrics{},
}

var (
	_ modules.Instance = &ModuleInstance{}
	_ modules.Module   = &RootModule{}
)

func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &ModuleInstance{
		vu:  vu,
		tcp: &TCP{vu: vu},
	}
}

type TCP struct {
	vu modules.VU
}

func (mi *ModuleInstance) Exports() modules.Exports {
	return modules.Exports{
		Default: mi.tcp,
	}
}

func (tcp *TCP) Connect(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (tcp *TCP) Write(conn net.Conn, data []byte) error {
	_, err := conn.Write(data)
	if err != nil {
		return err
	}

	metrics.PushIfNotDone(tcp.vu.Context(), tcp.vu.State().Samples, metrics.Sample{
		TimeSeries: metrics.TimeSeries{
			Metric: metrics.BuiltinMetrics.DataReceived,
			Tags:   nil,
		},
		Time:  time.Now(),
		Value: float64(len(data)),
	})
	return nil
}

func (tcp *TCP) Read(conn net.Conn, size int) ([]byte, error) {
	buf := make([]byte, size)
	_, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (tcp *TCP) WriteLn(conn net.Conn, data []byte) error {
	return tcp.Write(conn, append(data, []byte("\n")...))
}

func (tcp *TCP) Close(conn net.Conn) error {
	err := conn.Close()
	if err != nil {
		return err
	}
	return nil
}
