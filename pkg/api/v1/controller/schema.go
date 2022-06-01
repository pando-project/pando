package controller

import (
	"bytes"
	"fmt"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// SchemaRegister persistence a schema into database and then it
// will be loaded in memory so that we can build a go struct with reflect
// after that, we will transform the schema to an IPLD type system and create a bind-node
// then you can use it anywhere
// I hope it'll work well
// TODO implement schema persistence and load and make a go struct for using
//type Map struct {
//	Keys   []string
//	Values map[string]datamodel.Node
//}
//type List []datamodel.Node
//type List__MinerLocationModel []MinerLocationModel
//type MinerLocationsModel struct {
//	Epoch          int
//	Date           string
//	MinerLocations List__MinerLocationModel
//}
//type MinerLocationModel struct {
//	Miner        string
//	Region       string
//	Long         float64
//	Lat          float64
//	NumLocations int
//	Country      string
//	City         string
//	SubDiv1      string
//}
func (c *Controller) SchemaRegister(registerInfo []byte) error {
	typeSystem, err := ipld.LoadSchemaBytes(registerInfo)
	if err != nil {
		return err
	}

	goTypes := bytes.NewBuffer([]byte{})
	_, _ = fmt.Fprintln(goTypes, `package p`)
	_, _ = fmt.Fprintln(goTypes, `import "github.com/ipld/go-ipld-prime/datamodel"`)
	_, _ = fmt.Fprintln(goTypes, `var _ datamodel.Link // always used`)
	err = bindnode.ProduceGoTypes(goTypes, typeSystem)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", goTypes.String(), 0)
	if err != nil {
		return err
	}
	err = ast.Print(fset, node)
	if err != nil {
		return err
	}
	var typesPosition [][]int
	ast.Inspect(node, func(n ast.Node) bool {
		_, ok := n.(*ast.TypeSpec)
		if ok {
			//fmt.Println(tp.Type)
			start := fset.Position(n.Pos()).Line
			end := fset.Position(n.End()).Line
			typesPosition = append(typesPosition, []int{start, end})
			return false
		}
		return true
	})

	var goTypesInLines []string
	for {
		line, err := goTypes.ReadString('\n')
		if err != nil {
			fmt.Printf(err.Error())
			break
		}
		goTypesInLines = append(goTypesInLines, line)
	}

	var goTypesInParas []string
	for _, pos := range typesPosition {
		goTypesInParas = append(goTypesInParas, strings.Join(goTypesInLines[pos[0]-1:pos[1]], ""))
	}

	for i, goType := range goTypesInParas {
		fmt.Printf("type %d:\n", i)
		fmt.Println(goType)
		//reflect.StructOf()
	}

	return nil
}
