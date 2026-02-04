package v2

type DcName string

func (n DcName) String() string {
	return string(n)
}

type RackName string

func (n RackName) String() string {
	return string(n)
}

type DcRackName string

func (n DcRackName) String() string {
	return string(n)
}

type CompleteRackName struct {
	DcIndex    int
	RackIndex  int
	DcName     DcName
	RackName   RackName
	DcRackName DcRackName
}
