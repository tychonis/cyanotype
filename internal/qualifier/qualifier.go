package qualifier

import "github.com/tychonis/cyanotype/model"

func ImplicitProcess(item *model.Item) string {
	return item.Qualifier + ".__process__"
}

func ImplicitCoProcess(item *model.Item) string {
	return item.Qualifier + ".__coprocess__"
}

func ImplicitCoItem(item *model.Item) string {
	return item.Qualifier + ".__coitem__"
}
