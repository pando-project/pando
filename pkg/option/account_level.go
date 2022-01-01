package option

var defaultAccountLevel = []int{1, 10, 100, 500}

// AccountLevel is used for rank the accounts
type AccountLevel struct {
	// account balance
	Threshold []int `yaml:"Threshold"`
}
