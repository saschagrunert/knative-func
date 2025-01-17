package dist

import (
	"fmt"
	"sort"
	"strings"

	"github.com/buildpacks/lifecycle/api"

	"github.com/buildpacks/pack/internal/stringset"
	"github.com/buildpacks/pack/internal/style"
)

type BuildpackDescriptor struct {
	WithAPI    *api.Version `toml:"api"`
	WithInfo   ModuleInfo   `toml:"buildpack"`
	WithStacks []Stack      `toml:"stacks"`
	WithOrder  Order        `toml:"order"`
}

func (b *BuildpackDescriptor) EscapedID() string {
	return strings.ReplaceAll(b.Info().ID, "/", "_")
}

func (b *BuildpackDescriptor) EnsureStackSupport(stackID string, providedMixins []string, validateRunStageMixins bool) error {
	if len(b.Stacks()) == 0 {
		return nil // Order buildpack, no validation required
	}

	bpMixins, err := b.findMixinsForStack(stackID)
	if err != nil {
		return err
	}

	if !validateRunStageMixins {
		var filtered []string
		for _, m := range bpMixins {
			if !strings.HasPrefix(m, "run:") {
				filtered = append(filtered, m)
			}
		}
		bpMixins = filtered
	}

	_, missing, _ := stringset.Compare(providedMixins, bpMixins)
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("buildpack %s requires missing mixin(s): %s", style.Symbol(b.Info().FullName()), strings.Join(missing, ", "))
	}
	return nil
}

func (b *BuildpackDescriptor) Kind() string {
	return "buildpack"
}

func (b *BuildpackDescriptor) API() *api.Version {
	return b.WithAPI
}

func (b *BuildpackDescriptor) Info() ModuleInfo {
	return b.WithInfo
}

func (b *BuildpackDescriptor) Order() Order {
	return b.WithOrder
}

func (b *BuildpackDescriptor) Stacks() []Stack {
	return b.WithStacks
}

func (b *BuildpackDescriptor) findMixinsForStack(stackID string) ([]string, error) {
	for _, s := range b.Stacks() {
		if s.ID == stackID || s.ID == "*" {
			return s.Mixins, nil
		}
	}
	return nil, fmt.Errorf("buildpack %s does not support stack %s", style.Symbol(b.Info().FullName()), style.Symbol(stackID))
}
