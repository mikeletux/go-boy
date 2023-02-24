package cpu

func nopExecFunc(c *CPU) {
	c.emulateCpuCycles(4)
	return
}

func xorExecFunc(c *CPU) {
	c.registers.A ^= c.registers.A // This is for sure wrong. Double check
	if c.registers.A == 0x0 {
		c.registers.SetFZ(true)
	}
	c.emulateCpuCycles(4)
}

func (c *CPU) gotoAddr(address uint16, pushPC bool) {
	if checkCondition(c) {
		if pushPC {
			c.stackPush16(c.registers.PC)
			c.emulateCpuCycles(2)
		}
		c.registers.PC = address
		c.emulateCpuCycles(1)
	}
}

func jpExecFunc(c *CPU) {
	c.gotoAddr(c.FetchedData, false)
}

func callExecFunc(c *CPU) {
	c.gotoAddr(c.FetchedData, true)
}

func retExecFunc(c *CPU) {
	if c.CurrentInstruction.Condition != ctNone {
		c.emulateCpuCycles(1)
	}

	if checkCondition(c) {
		low := c.stackPop()
		c.emulateCpuCycles(1)
		high := c.stackPop()
		c.emulateCpuCycles(1)

		c.registers.PC = uint16(high<<8) | uint16(low)
		c.emulateCpuCycles(1)
	}
}

func retiExecFunc(c *CPU) {
	c.EnableMasterInterruptions = true
	retExecFunc(c)
}

func jrExecFunc(c *CPU) {
	rel := int8(c.FetchedData & 0xFF) // This byte must be signed
	addr := c.registers.PC + uint16(rel)
	c.gotoAddr(addr, false)
}

func popExecFunc(c *CPU) {
	low := uint16(c.stackPop()) // Read the least significant byte
	c.emulateCpuCycles(1)
	high := uint16(c.stackPop()) // Read the most significant byte
	c.emulateCpuCycles(1)
	c.registers.SetDataToRegisters(c.CurrentInstruction.RegisterType1, high<<8|low)

	if c.CurrentInstruction.RegisterType1 == rtAF {
		c.registers.SetDataToRegisters(c.CurrentInstruction.RegisterType1, (high<<8|low)&0xFFF0)
	}
}

func pushExecFunc(c *CPU) {
	value, err := c.registers.FetchDataFromRegisters(c.CurrentInstruction.RegisterType1)
	if err != nil {
		c.logger.Fatal(err)
	}

	c.stackPush(byte(value>>8) & 0xFF) // Push the most significant byte
	c.emulateCpuCycles(1)
	c.stackPush(byte(value) & 0xFF) // Push the least significant byte
	c.emulateCpuCycles(1)
}

func checkCondition(c *CPU) bool {
	fz := c.registers.GetFZ()
	fc := c.registers.GetFC()

	switch c.CurrentInstruction.Condition {
	case ctNone:
		return true
	case ctZ:
		return fz
	case ctNZ:
		return !fz
	case ctC:
		return fc
	case ctNC:
		return !fc
	}
	return true // This never should be reached
}

func diExecFunc(c *CPU) {
	c.EnableMasterInterruptions = false
}

func ldExecFunc(c *CPU) {
	if c.DestinationIsMemory {
		// We need to write in memory
		if c.CurrentInstruction.RegisterType2 >= rtAF { // This means we need to write twice in memory.
			c.bus.BusWrite16(c.MemoryDestination, c.FetchedData)
		} else {
			c.bus.BusWrite(c.MemoryDestination, byte(c.FetchedData))
		}
		return
	}

	if c.CurrentInstruction.AddressingMode == amHLnSPR {
		c.registers.SetFZ(false)
		c.registers.SetFN(false)
		reg2Value, err := c.registers.FetchDataFromRegisters(c.CurrentInstruction.RegisterType2)
		if err != nil {
			c.logger.Fatalf("error when executing LD HL SP(r) operation: %s", err)
		}

		c.registers.SetFH((reg2Value&0xF)+(c.FetchedData&0xF) >= 0x10)    // If lower 4 bits of result overflow, set H.
		c.registers.SetFC((reg2Value&0xFF)+(c.FetchedData&0xFF) >= 0x100) // If upper 4 bits of result overflow, set C.

		c.registers.SetDataToRegisters(c.CurrentInstruction.RegisterType1, reg2Value+c.FetchedData)
	}

	c.registers.SetDataToRegisters(c.CurrentInstruction.RegisterType1, c.FetchedData) // Normal case.
}

func ldhExecFunc(c *CPU) {
	if c.CurrentInstruction.RegisterType1 == rtA {
		c.registers.A = c.bus.BusRead(0xFF00 | c.FetchedData)
	} else {
		c.bus.BusWrite(0xFF00|c.FetchedData, c.registers.A)
	}

	c.emulateCpuCycles(1)
}
