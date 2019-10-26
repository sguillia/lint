package customer

type User struct {
	name    string
	surname string
	age     int
	active  bool
}

func isSaleValid(customer User, seller string, items int) bool {
	if !isCustomerValid(customer) {
		return false
	}
	if seller == "" || seller == customer.name {
		return false
	}
	if customer.age < 13 {
		return false
	}
	if items <= 0 {
		return false
	}
	return customer.active
}

func isCustomerValid(customer User) bool {
	return len(customer.name) != 0 && len(customer.surname) != 0
}
