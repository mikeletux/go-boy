package bus

import (
	"github.com/mikeletux/goboy/pkg/cart"
	"github.com/mikeletux/goboy/pkg/log"
)

type DataBusInterface interface {
	BusRead(address uint16) byte
	BusWrite(address uint16, value byte)
	BusRead16(address uint16) uint16
	BusWrite16(address uint16, value uint16)
	// Probably I need a method to return the ieRegister
}

// Bus represents the whole Gameboy bus
type Bus struct {
	// Cartridge represents the Gameboy game cartridge. It has functions to read and write its memory.
	Cartridge  cart.CartridgeInterface
	logger     log.Logger
	ram        *Ram
	ieRegister byte
}

// NewBus initializes a bus given a type that implements the cart.CartridgeInterface interface.
func NewBus(cartridge cart.CartridgeInterface, logger log.Logger) *Bus {
	return &Bus{
		Cartridge: cartridge,
		logger:    logger,
		ram:       NewRam(logger),
	}
}

// BusRead given an address it returns the value from that bus memory area.
func (b *Bus) BusRead(address uint16) byte {
	switch {
	case address >= RomBank00Start && address <= RomBank01NNEnd: // Cartridge ROM area
		return b.Cartridge.CartRead(address)

	case address >= VramStart && address <= VramEnd: // VRAM area
		// TO-DO
		b.logger.Fatalf("VRAM area bus address 0x%X yet not implemented to read", address)

	case address >= ExternalRamFromCartridgeStart && address <= ExternalRamFromCartridgeEnd: // Cartridge RAM area
		return b.Cartridge.CartRead(address)

	case address >= WorkRam0Start && address <= WorkRam1End: // Working RAM area
		return b.ram.readWorkingRam(address)

	case address >= EchoRamStart && address <= EchoRamEnd: // Reserved echo RAM area
		return 0x0

	case address >= OamStart && address <= OamEnd: // Sprite attribute table area
		// TO-DO
		b.logger.Fatalf("OAM area bus address 0x%X yet not implemented to read", address)

	case address >= NintendoNotUsableMemoryStart && address <= NintendoNotUsableMemoryEnd: // Not usable area
		return 0x0

	case address >= IORegistersStart && address <= IORegistersEnd: // IO Registers area
		// TO-DO
		b.logger.Fatalf("IO Registers area bus address 0x%X yet not implemented to read", address)

	case address >= HighRamStart && address <= HighRamEnd: // High RAM area
		b.ram.readHighRam(address)

	case address == InterruptEnableRegister: // CPU enable register
		return b.ieRegister

	default:
		b.logger.Fatalf("Unknown memory bus address 0x%X to read", address) // Code should never reach here
	}

	return 0
}

// BusWrite writes a byte into bus given a bus memory area.
func (b *Bus) BusWrite(address uint16, value byte) {
	switch {
	case address >= RomBank00Start && address <= RomBank01NNEnd: // Cartridge ROM area
		b.Cartridge.CartWrite(address, value)

	case address >= VramStart && address <= VramEnd: // VRAM area
		// TO-DO
		b.logger.Fatalf("VRAM area bus address 0x%X yet not implemented to write", address)

	case address >= ExternalRamFromCartridgeStart && address <= ExternalRamFromCartridgeEnd: // Cartridge RAM area
		b.Cartridge.CartWrite(address, value)

	case address >= WorkRam0Start && address <= WorkRam1End: // Working RAM area
		b.ram.writeWorkingRam(address, value)

	case address >= OamStart && address <= OamEnd: // Sprite attribute table area
		// TO-DO
		b.logger.Fatalf("OAM area bus address 0x%X yet not implemented to write", address)

	case address >= IORegistersStart && address <= IORegistersEnd: // IO Registers area
		// TO-DO
		b.logger.Debugf("IO Registers area bus address 0x%X yet not implemented to write", address)

	case address >= HighRamStart && address <= HighRamEnd: // High RAM area
		b.ram.writeHighRam(address, value)

	case address == InterruptEnableRegister: // CPU enable register
		b.ieRegister = value

	default:
		b.logger.Fatalf("Unknown memory bus address 0x%X to write", address) // Code should never reach here
	}

}

// BusRead16 given an address it returns the 16 bit value from that area.
func (b *Bus) BusRead16(address uint16) uint16 {
	low := b.BusRead(address)
	high := b.BusRead(address + 1)
	return uint16(high<<8 | low)
}

// BusWrite16 writes 16 bit value into bus given a bus memory area.
func (b *Bus) BusWrite16(address uint16, value uint16) {
	b.BusWrite(address, byte(value&0xFF))        // Low
	b.BusWrite(address+1, byte((value>>8)&0xFF)) // High
}
