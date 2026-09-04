# ATIPICIAL-GO-VM

A cross platform virtual machine implementation for `AEF` compatible programs. 

# Installation

VM is provided as a part of atipicial-go binary, so usual atipicial-go build instructions
are applicable.

# Running the VM

Start the virtual machine:

```
$ ./bin/atipicial-go vm

    _   ____________        __________      _    ____  ___
   / | / / ____/ __ \      / ____/ __ \    | |  / /  |/  /
  /  |/ / __/ / / / /_____/ / __/ / / /____| | / / /|_/ /
 / /|  / /___/ /_/ /_____/ /_/ / /_/ /_____/ |/ / /  / /
/_/ |_/_____/\____/      \____/\____/      |___/_/  /_/


ATIPICIAL-GO-VM >
```

# Usage

```
    _   ____________        __________      _    ____  ___
   / | / / ____/ __ \      / ____/ __ \    | |  / /  |/  /
  /  |/ / __/ / / / /_____/ / __/ / / /____| | / / /|_/ /
 / /|  / /___/ /_/ /_____/ /_/ / /_/ /_____/ |/ / /  / /
/_/ |_/_____/\____/      \____/\____/      |___/_/  /_/


ATIPICIAL-GO-VM > help

Commands:
  aslot           Show arguments slot contents
  break           Place a breakpoint
  clear           clear the screen
  cont            Continue execution of the current loaded script
  estack          Show evaluation stack contents
  exit            Exit the VM prompt
  help            display help
  ip              Show current instruction
  istack          Show invocation stack contents
  loadbase64      Load a base64-encoded script string into the VM
  loadgo          Compile and load a Go file with the manifest into the VM
  loadhex         Load a hex-encoded script string into the VM
  loadaef         Load a AEF-consistent script into the VM
  lslot           Show local slot contents
  ops             Dump opcodes of the current loaded program
  parse           Parse provided argument and convert it into other possible formats
  run             Execute the current loaded script
  sslot           Show static slot contents
  step            Step (n) instruction in the program
  stepinto        Stepinto instruction to take in the debugger
  stepout         Stepout instruction to take in the debugger
  stepover        Stepover instruction to take in the debugger

```

You can get help for each command and its parameters adding `help` as a
parameter to the command:

```
ATIPICIAL-GO-VM > step help

Usage: step [<n>]
<n> is optional parameter to specify number of instructions to run, example:
> step 10

```

## Loading in your script

To load an avm script in AEF format into the VM:

```
ATIPICIAL-GO-VM > loadaef ../contract.aef
READY: loaded 36 instructions
```

Run the script:

```
ATIPICIAL-GO-VM > run
[
    {
        "value": 1,
        "type": "BigInteger"
    }
]
```

You can also directly compile and load `.go` files:

```
ATIPICIAL-GO-VM > loadgo ../contract.go
READY: loaded 36 instructions
```

To make it even more complete, you can directly load hex or base64 strings into the VM:

```
ATIPICIAL-GO-VM > loadhex 0c0c48656c6c6f20776f726c6421
READY: loaded 14 instructions
ATIPICIAL-GO-VM 0 > run
[
    {
        "type": "ByteString",
        "value": "SGVsbG8gd29ybGQh"
    }
]

```

## Running programs with arguments
You can invoke smart contracts with arguments. Take the following ***roll the dice*** smart contract as an example. 

```
package rollthedice

import "github.com/atipicial/atipicial-go/pkg/interop/runtime"

func RollDice(number int) {
    if number == 0 {
        runtime.Log("you rolled 0, better luck next time!")
    }
    if number == 1 {
        runtime.Log("you rolled 1, still better then 0!")
    }
    if number == 2 {
        runtime.Log("you rolled 2, coming closer..") 
    }
    if number == 3 {
        runtime.Log("Sweet you rolled 3. This dice has only 3 sides o_O")
    }
}
```

To invoke this contract we need to specify both the method and the arguments.

The first parameter (called method or operation) is always of type
string. Notice that arguments can have different types. They can be inferred
automatically (please refer to the `run` command help), but if you need to
pass a parameter of a specific type you can specify it in `run`'s arguments:

```
ATIPICIAL-GO-VM > run rollDice int:1
```

> The method is always of type string, hence we don't need to specify the type.

To add more than 1 argument:

```
ATIPICIAL-GO-VM > run someMethod int:1 int:2 string:foo string:bar
```

Currently supported types:
- `bool (bool:false and bool:true)`
- `int (int:1 int:100)`
- `string (string:foo string:this is a string)` 

## Debugging
The `atipicial-go-vm` provides a debugger to inspect your program in-depth.


### Stepping through the program
Step 4 instructions.

```
ATIPICIAL-GO-VM > step 4
at breakpoint 3 (DUPFROMALTSTACK)
ATIPICIAL-GO-VM 3 >
```

Using just `step` will execute 1 instruction at a time.

```
ATIPICIAL-GO-VM 3 > step
at breakpoint 4 (PUSH0)
ATIPICIAL-GO-VM 4 >
```

### Breakpoints

To place breakpoints:

```
ATIPICIAL-GO-VM > break 10
breakpoint added at instruction 10
ATIPICIAL-GO-VM > cont
at breakpoint 10 (SETITEM)
ATIPICIAL-GO-VM 10 > cont
```

## Inspecting stack

Inspecting the evaluation stack:

```
ATIPICIAL-GO-VM > estack
[
    {
        "value": [
            null,
            null,
            null,
            null,
            null,
            null,
            null
        ],
        "type": "Array"
    },
    {
        "value": 4,
        "type": "BigInteger"
    }
]
```

There is one more stack that you can inspect.
- `istack` invocation stack

There are slots that you can inspect.
- `aslot` dumps arguments slot contents.
- `lslot` dumps local slot contents.
- `sslot` dumps static slot contents.

