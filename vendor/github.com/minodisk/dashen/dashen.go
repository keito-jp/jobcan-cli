package dashen

import (
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/pkg/errors"
)

var (
	AlreadyListening = errors.New("already listening")
	NotListening     = errors.New("not listening")
	NoWriter         = errors.New("no writer")
)

type Dashen struct {
	Logger          io.Writer
	LoggerVerbose   io.Writer
	MACCallbacksMap MACCallbacksMap
	doneCh          chan struct{}
	mutex           *sync.Mutex
	listening       bool
}

func New() *Dashen {
	return &Dashen{
		MACCallbacksMap: MACCallbacksMap{},
		mutex:           new(sync.Mutex),
		doneCh:          make(chan struct{}),
	}
}

func (d *Dashen) Listen() error {
	d.mutex.Lock()
	if d.listening {
		d.mutex.Unlock()
		return AlreadyListening
	}
	d.listening = true
	d.mutex.Unlock()

	ifaces, err := pcap.FindAllDevs()
	if err != nil {
		return err
	}

	errCh := make(chan error)
	go func(errCh chan error) {
		for {
			select {
			case <-d.doneCh:
				return
			case err := <-errCh:
				d.Log(err)
			}
		}
	}(errCh)

	var wg sync.WaitGroup
	for _, iface := range ifaces {
		wg.Add(1)
		go func(iface pcap.Interface) {
			defer func() {
				wg.Done()
			}()
			if err := Filter(iface); err != nil {
				errCh <- errors.Wrap(err, "skip to open interface")
				return
			}
			if err := d.listen(iface, errCh); err != nil {
				errCh <- errors.Wrap(err, "fail to open interface with error")
				return
			}
		}(iface)
	}
	wg.Wait()

	return nil
}

func (d *Dashen) listen(iface pcap.Interface, errCh chan error) error {
	d.Log("open interface:", iface.Name)
	handle, err := pcap.OpenLive(iface.Name, 10240, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	// handle.Close() blocks this goroutine
	// https://github.com/google/gopacket/issues/253
	// defer handle.Close()

	d.Log("capture packets:", handle.LinkType().String(), iface.Name)
	src := gopacket.NewPacketSource(handle, handle.LinkType())
	in := src.Packets()
	for {
		select {
		case <-d.doneCh:
			return nil
		case packets := <-in:
			// when the type of Layer 2 is LLC,
			// it may be the first packets from the Dash Button
			llcl := packets.Layer(layers.LayerTypeLLC)
			if llcl == nil {
				continue
			}
			if _, ok := llcl.(*layers.LLC); !ok {
				continue
			}
			// get MAC addr
			ethernetl := packets.Layer(layers.LayerTypeEthernet)
			if ethernetl == nil {
				continue
			}
			ethernet, ok := ethernetl.(*layers.Ethernet)
			if !ok {
				continue
			}
			srcMAC := ethernet.SrcMAC.String()
			d.LogVerbose("detect:", srcMAC)
			// scan callback map
			for mac, cbs := range d.MACCallbacksMap {
				if mac != srcMAC {
					continue
				}
				for _, cb := range cbs {
					cb()
				}
			}
		}
	}
}

func (d *Dashen) Close() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.listening {
		return NotListening
	}
	close(d.doneCh)
	d.listening = false
	return nil
}

func (d *Dashen) Subscribe(mac string, callback Callback) {
	cbs, ok := d.MACCallbacksMap[mac]
	if !ok {
		d.MACCallbacksMap[mac] = []Callback{callback}
		return
	}
	ptr := reflect.ValueOf(callback).Pointer()
	for _, cb := range cbs {
		if reflect.ValueOf(cb).Pointer() == ptr {
			return
		}
	}
	d.MACCallbacksMap[mac] = append(cbs, callback)
}

func (d *Dashen) Unsubscribe(mac string, callback Callback) {
	cbs, ok := d.MACCallbacksMap[mac]
	if !ok {
		return
	}
	callbacks := []Callback{}
	for _, cb := range cbs {
		if IsEquals(cb, callback) {
			continue
		}
		callbacks = append(callbacks, cb)
	}
	d.MACCallbacksMap[mac] = callbacks
}

func (d *Dashen) Log(v ...interface{}) (int, error) {
	if d.Logger == nil {
		return 0, NoWriter
	}
	return fmt.Fprintln(d.Logger, v...)
}

func (d *Dashen) LogVerbose(v ...interface{}) (int, error) {
	if d.LoggerVerbose == nil {
		return 0, NoWriter
	}
	return fmt.Fprintln(d.LoggerVerbose, v...)
}
