package lane

import "github.com/lixianmin/pc/pkg/lane/intern"

/********************************************************************
created:    2020-10-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Acceptor interface {
	Listen()
	GetLinkChan() chan intern.Link
}
