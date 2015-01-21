package m

import (
	"testing"
)

func TestSrc(t *testing.T) {

	<-Src([]string{"/usr/share/**"}).Pipe(Dist("/public"))
}
