package resources

import "github.com/tucats/ego/app-cli/ui"

func (r *ResHandle) Insert(v interface{}) error {
	var err error

	if r.Err != nil {
		return r.Err
	}

	sql := r.insertSQL()
	items := r.explode(v)

	ui.Log(ui.ResourceLogger, "resource.insert",
		"sql", sql)
	ui.Log(ui.ResourceLogger, "resource.parms",
		"list", items)

	_, err = r.Database.Exec(sql, items...)

	return err
}
