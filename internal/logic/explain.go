package logic

import (
	"context"
	"errors"
)

type Explanation struct {
	Result bool              `json:"result"`
	Proofs []map[string]string `json:"proofs,omitempty"`
}

// Explain evaluates an explicit proof-producing predicate.
//
// Rulesets opt in by defining a predicate whose solutions contain a Proof
// variable, for example:
//
//   trusted(alice).
//   may_voice(X, fact(trusted(X))) :- trusted(X).
//
// Query with: may_voice(alice, Proof).
//
// BotLogic deliberately does not invent proofs for arbitrary Prolog goals.
// The returned proof is produced by the ruleset itself and is therefore
// deterministic, inspectable and testable.
func (e *Engine) Explain(ctx context.Context, query string) (Explanation, error) {
	rows, err := e.Query(ctx, query)
	if err != nil {
		return Explanation{}, err
	}
	for _, row := range rows {
		if _, ok := row["Proof"]; !ok {
			return Explanation{}, errors.New("explain query must bind Proof")
		}
	}
	return Explanation{Result: len(rows) > 0, Proofs: rows}, nil
}
