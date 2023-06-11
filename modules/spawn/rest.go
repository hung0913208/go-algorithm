package spawn

import (
	"fmt"
	"net/url"
)

func (self *spawnImpl) DoGet(
	function string, kwargs url.Values,
) (interface{}, error) {
	return nil, fmt.Errorf("not support `%s`", function)
}

func (self *spawnImpl) DoPost(
	function string, kwargs url.Values,
) (interface{}, error) {
	return nil, fmt.Errorf("not support `%s`", function)
}
