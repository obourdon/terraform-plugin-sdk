package configschema

import (
	"testing"

	"github.com/zclconf/go-cty/cty"

	multierror "github.com/hashicorp/go-multierror"
)

func TestBlockInternalValidate(t *testing.T) {
	tests := map[string]struct {
		Block    *Block
		ErrCount int
	}{
		"empty": {
			&Block{},
			0,
		},
		"valid": {
			&Block{
				Attributes: map[string]*Attribute{
					"foo": {
						Type:     cty.String,
						Required: true,
					},
					"bar": {
						Type:     cty.String,
						Optional: true,
					},
					"baz": {
						Type:     cty.String,
						Computed: true,
					},
					"baz_maybe": {
						Type:     cty.String,
						Optional: true,
						Computed: true,
					},
				},
				BlockTypes: map[string]*NestedBlock{
					"single": {
						Nesting: NestingSingle,
						Block:   Block{},
					},
					"single_required": {
						Nesting:  NestingSingle,
						Block:    Block{},
						MinItems: 1,
						MaxItems: 1,
					},
					"list": {
						Nesting: NestingList,
						Block:   Block{},
					},
					"list_required": {
						Nesting:  NestingList,
						Block:    Block{},
						MinItems: 1,
					},
					"set": {
						Nesting: NestingSet,
						Block:   Block{},
					},
					"set_required": {
						Nesting:  NestingSet,
						Block:    Block{},
						MinItems: 1,
					},
					"map": {
						Nesting: NestingMap,
						Block:   Block{},
					},
				},
			},
			0,
		},
		"attribute with no flags set": {
			&Block{
				Attributes: map[string]*Attribute{
					"foo": {
						Type: cty.String,
					},
				},
			},
			1, // must set one of the flags
		},
		"attribute required and optional": {
			&Block{
				Attributes: map[string]*Attribute{
					"foo": {
						Type:     cty.String,
						Required: true,
						Optional: true,
					},
				},
			},
			1, // both required and optional
		},
		"attribute required and computed": {
			&Block{
				Attributes: map[string]*Attribute{
					"foo": {
						Type:     cty.String,
						Required: true,
						Computed: true,
					},
				},
			},
			1, // both required and computed
		},
		"attribute optional and computed": {
			&Block{
				Attributes: map[string]*Attribute{
					"foo": {
						Type:     cty.String,
						Optional: true,
						Computed: true,
					},
				},
			},
			0,
		},
		"attribute with missing type": {
			&Block{
				Attributes: map[string]*Attribute{
					"foo": {
						Optional: true,
					},
				},
			},
			1, // Type must be set
		},
		"attribute with invalid name": {
			&Block{
				Attributes: map[string]*Attribute{
					"fooBar": {
						Type:     cty.String,
						Optional: true,
					},
				},
			},
			1, // name may not contain uppercase letters
		},
		"block type with invalid name": {
			&Block{
				BlockTypes: map[string]*NestedBlock{
					"fooBar": {
						Nesting: NestingSingle,
					},
				},
			},
			1, // name may not contain uppercase letters
		},
		"colliding names": {
			&Block{
				Attributes: map[string]*Attribute{
					"foo": {
						Type:     cty.String,
						Optional: true,
					},
				},
				BlockTypes: map[string]*NestedBlock{
					"foo": {
						Nesting: NestingSingle,
					},
				},
			},
			1, // "foo" is defined as both attribute and block type
		},
		"nested block with badness": {
			&Block{
				BlockTypes: map[string]*NestedBlock{
					"bad": {
						Nesting: NestingSingle,
						Block: Block{
							Attributes: map[string]*Attribute{
								"nested_bad": {
									Type:     cty.String,
									Required: true,
									Optional: true,
								},
							},
						},
					},
				},
			},
			1, // nested_bad is both required and optional
		},
		"nested list block with dynamically-typed attribute": {
			&Block{
				BlockTypes: map[string]*NestedBlock{
					"bad": {
						Nesting: NestingList,
						Block: Block{
							Attributes: map[string]*Attribute{
								"nested_bad": {
									Type:     cty.DynamicPseudoType,
									Optional: true,
								},
							},
						},
					},
				},
			},
			0,
		},
		"nested set block with dynamically-typed attribute": {
			&Block{
				BlockTypes: map[string]*NestedBlock{
					"bad": {
						Nesting: NestingSet,
						Block: Block{
							Attributes: map[string]*Attribute{
								"nested_bad": {
									Type:     cty.DynamicPseudoType,
									Optional: true,
								},
							},
						},
					},
				},
			},
			1, // NestingSet blocks may not contain attributes of cty.DynamicPseudoType
		},
		"nil": {
			nil,
			1, // block is nil
		},
		"nil attr": {
			&Block{
				Attributes: map[string]*Attribute{
					"bad": nil,
				},
			},
			1, // attribute schema is nil
		},
		"nil block type": {
			&Block{
				BlockTypes: map[string]*NestedBlock{
					"bad": nil,
				},
			},
			1, // block schema is nil
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			errs := multierrorErrors(test.Block.InternalValidate())
			if got, want := len(errs), test.ErrCount; got != want {
				t.Errorf("wrong number of errors %d; want %d", got, want)
				for _, err := range errs {
					t.Logf("- %s", err.Error())
				}
			}
		})
	}
}

func multierrorErrors(err error) []error {
	// A function like this should really be part of the multierror package...

	if err == nil {
		return nil
	}

	switch terr := err.(type) {
	case *multierror.Error:
		return terr.Errors
	default:
		return []error{err}
	}
}
