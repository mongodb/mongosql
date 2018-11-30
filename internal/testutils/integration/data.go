package integration

import "github.com/10gen/sqlproxy/internal/testutils/data"

var datasetsByName = map[string]data.Dataset{
	"blackbox": blackboxDataset,
	"internal": internalDataset,
	"tableau":  tableauDataset,
	"tdvt":     tdvtDataset,
}

var blackboxDataset = data.DatasetGroup{
	tdvtDataset,
	data.NewBSONDataset34("DecimalRei"),
	data.NewBSONDataset34("DecimalUTStarcom"),
	data.NewBSONDataset("Batters"),
	data.NewBSONDataset("DateTime"),
	data.NewBSONDataset("Election"),
	data.NewBSONDataset("Fischeriris"),
	data.NewBSONDataset("Loan"),
	data.NewBSONDataset("NumericBins"),
	data.NewBSONDataset("SeattleCrime"),
	data.NewBSONDataset("Securities"),
	data.NewBSONDataset("SpecialData"),
	data.NewBSONDataset("Starbucks"),
	data.NewBSONDataset("xy"),
	data.NewBSONDataset("bigarray"),
	data.NewBSONDataset("bigcoll"),
	data.NewBSONDataset("bignestedarray"),
	data.NewBSONDataset("bigobjarray"),
}

var internalDataset = data.DatasetGroup{
	data.YMLDataset{File: "internal.yml"},
	data.NewBSONDataset("test1"),
	data.NewBSONDataset("test2"),
	data.NewBSONDataset("test3"),
	data.NewBSONDataset("test4"),
}

var tableauDataset = data.DatasetGroup{
	data.NewBSONDataset("attendees-new"),
	data.NewBSONDataset("flights201406"),
}

var tdvtDataset = data.DatasetGroup{
	data.NewBSONDataset34("DecimalStaples"),
	data.NewBSONDataset("Calcs"),
}
