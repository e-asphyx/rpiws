package rpiws

/*

#include "ws2811.h"

*/
import "C"
import (
	"errors"
	"reflect"
	"runtime"
	"unsafe"
)

const (
	WS2811_TARGET_FREQ = C.WS2811_TARGET_FREQ
	RPI_PWM_CHANNELS   = C.RPI_PWM_CHANNELS
)

var (
	ErrHardware = errors.New("Hardware error")
)

type Led uint32

type Driver struct {
	device  unsafe.Pointer // Private
	Freq    uint32
	Dmanum  int32
	Channel [RPI_PWM_CHANNELS]Channel
}

type Channel struct {
	Gpionum    int32
	Invert     int32
	Count      int32
	Brightness int32
	leds       unsafe.Pointer // use Leds() for access
}

func (driver *Driver) cptr() *C.ws2811_t {
	return (*C.ws2811_t)(unsafe.Pointer(driver))
}

func (driver *Driver) Init() error {
	if ret, err := C.ws2811_init(driver.cptr()); ret < 0 {
		if err != nil {
			// errno is set
			return err
		}

		return errors.New("Initialization error")
	}

	return nil
}

func (driver *Driver) Fini() error {
	C.ws2811_fini(driver.cptr())
	return nil
}

func (driver *Driver) Render() error {
	if err := driver.Wait(); err != nil {
		return err
	}

	if C.ws2811_render(driver.cptr()) < 0 {
		return ErrHardware
	}

	return nil
}

func (driver *Driver) Ready() (bool, error) {
	ret := C.ws2811_dma_ready(driver.cptr())

	if ret < 0 {
		return false, ErrHardware
	}

	return ret != 0, nil
}

func (driver *Driver) Wait() error {
	var err error

	for {
		var ready bool
		ready, err = driver.Ready()
		if ready || err != nil {
			break
		}
		runtime.Gosched()
	}

	return err
}

func (channel *Channel) Leds() []Led {
	if channel.Count == 0 {
		return nil
	}

	sh := reflect.SliceHeader{
		Data: uintptr(channel.leds),
		Len:  int(channel.Count),
		Cap:  int(channel.Count),
	}

	return *(*[]Led)(unsafe.Pointer(&sh))
}

// Helper functions

func (led Led) R() uint8 {
	return uint8((uint32(led) >> 16) & 0xff)
}

func (led Led) G() uint8 {
	return uint8((uint32(led) >> 8) & 0xff)
}

func (led Led) B() uint8 {
	return uint8(uint32(led) & 0xff)
}

func RGB(r, g, b uint8) Led {
	return Led((uint32(r) << 16) | (uint32(g) << 8) | uint32(b))
}
