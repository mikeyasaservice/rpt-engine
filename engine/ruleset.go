package engine

import (
	"fmt"
	"os"
	"sync"
)

// Config is used as argument to creating a new ruleset
type Config struct {
	Directory       []string
	FailOnRuleParse bool
	FailOnYamlParse bool
	NoCollapseWS    bool
}

func (c Config) validate() error {
	if c.Directory == nil || len(c.Directory) == 0 {
		return fmt.Errorf("missing root directory for sigma rules")
	}
	for _, dir := range c.Directory {
		info, err := os.Stat(dir)
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist", dir)
		}
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", dir)
		}
	}
	return nil
}

// Ruleset is a collection of rules
type Ruleset struct {
	mu          *sync.RWMutex
	Rules       []*Tree
	root        []string
	Total       int
	Ok          int
	Failed      int
	Unsupported int
	Errors      []error // Add this field to track errors
}

// NewRuleset instanciates a Ruleset object
func NewRuleset(c Config, tags []string) (*Ruleset, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	files, err := NewRuleFileList(c.Directory)
	if err != nil {
		return nil, err
	}
	var fail int
	rules, err := NewRuleList(files, !c.FailOnYamlParse, c.NoCollapseWS, tags)
	if err != nil {
		switch e := err.(type) {
		case ErrBulkParseYaml:
			fail += len(e.Errs)
		default:
			return nil, err
		}
	}
	result := RulesetFromRuleList(rules)
	result.root = c.Directory
	result.Failed += fail
	result.Total += fail
	return result, nil
}

func RulesetFromRuleList(rules []RuleHandle) *Ruleset {
	var fail, unsupp int
	set := make([]*Tree, 0)
	errors := make([]error, 0) // Initialize an error slice

loop:
	for _, raw := range rules {
		if raw.Multipart {
			unsupp++
			continue loop
		}
		tree, err := NewTree(raw)
		if err != nil {
			switch err.(type) {
			case ErrUnsupportedToken, *ErrUnsupportedToken:
				unsupp++
			default:
				fail++
				errors = append(errors, err) // Collect errors
			}
			continue loop
		}
		set = append(set, tree)
	}
	return &Ruleset{
		mu:          &sync.RWMutex{},
		Rules:       set,
		Failed:      fail,
		Ok:          len(set),
		Unsupported: unsupp,
		Total:       len(rules),
		Errors:      errors, // Assign errors
	}
}

func (r *Ruleset) EvalAll(e Event) (Results, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make(Results, 0)
	for _, rule := range r.Rules {
		if res, match := rule.Eval(e); match {
			results = append(results, *res)
		}
	}
	if len(results) > 0 {
		return results, true
	}
	return nil, false
}
