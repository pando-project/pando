package metadata

import (
	"errors"
	"fmt"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/bindnode"
)

type Metadata struct {
	PreviousID     *ipld.Link
	Provider       string
	Cache          *bool
	DatabaseName   *string
	CollectionName *string
	Payload        datamodel.Node
	Signature      []byte
}

// ToNode converts this metadata to its representation as an IPLD typed node.
// See: bindnode.Wrap.
func (m *Metadata) ToNode() (n ipld.Node, err error) {
	// TODO: remove the panic recovery once IPLD bindnode is stabilized.
	defer func() {
		if r := recover(); r != nil {
			err = toError(r)
		}
	}()
	n = bindnode.Wrap(m, MetadataPrototype.Type()).Representation()
	return
}

// UnwrapMetadata unwraps the given node as metadata.
//
// Note that the node is reassigned to MetadataPrototype if its prototype is different.
// Therefore, it is recommended to load the node using the correct prototype initially
// function to avoid unnecessary node assignment.
func UnwrapMetadata(node ipld.Node) (*Metadata, error) {
	// When an IPLD node is loaded using `Prototype.Any` unwrap with bindnode will not work.
	// Here we defensively check the prototype and wrap if needed, since:
	//   - linksystem in sti is passed into other libraries, like go-legs, and
	//   - for whatever reason clients of this package may load nodes using Prototype.Any.
	//
	// The code in this repo, however should load nodes with appropriate prototype and never trigger
	// this if statement.
	if node.Prototype() != MetadataPrototype {
		adBuilder := MetadataPrototype.NewBuilder()
		err := adBuilder.AssignNode(node)
		if err != nil {
			return nil, fmt.Errorf("faild to convert node prototype: %w", err)
		}
		node = adBuilder.Build()
	}

	ad, ok := bindnode.Unwrap(node).(*Metadata)
	if !ok || ad == nil {
		return nil, fmt.Errorf("unwrapped node does not match schema.Metadata")
	}
	return ad, nil
}

func toError(r interface{}) error {
	switch x := r.(type) {
	case string:
		return errors.New(x)
	case error:
		return x
	default:
		return fmt.Errorf("unknown panic: %v", r)
	}
}
