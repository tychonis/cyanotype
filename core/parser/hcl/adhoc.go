package hcl

// Adhoc function supporting bomhub.
func (c *Core) SaveCatalog(endpoint string) error {
	return c.Catalog.Save(endpoint)
}
