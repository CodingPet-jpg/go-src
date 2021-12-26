package strconv

// enable fast path for small integers
const fastSmalls = true

// ^uint(0) in
// 32 bit machine   0xFFFF FFFF
// 64 bit machine   0xFFFF FFFF FFFF FFFF
const host32bit = ^uint(0) >> 32
const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
const smallsString = "00010203040506070809" +
	"10111213141516171819" +
	"20212223242526272829" +
	"30313233343536373839" +
	"40414243444546474849" +
	"50515253545556575859" +
	"60616263646566676869" +
	"70717273747576777879" +
	"80818283848586878889" +
	"90919293949596979899"

func small() {

}

func formatBits() {

}