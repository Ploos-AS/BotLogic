package logic

import (
    "context"
    "fmt"
    "sync"

    "github.com/ichiban/prolog"
)

type Engine struct {
    mu sync.Mutex
    p  *prolog.Interpreter
}

func New() *Engine {
    return &Engine{p: prolog.New(nil, nil)}
}

func (e *Engine) Consult(src string) error {
    e.mu.Lock()
    defer e.mu.Unlock()
    return e.p.Exec(src)
}

func (e *Engine) Query(ctx context.Context, q string) ([]map[string]string, error) {
    e.mu.Lock()
    defer e.mu.Unlock()
    sols, err := e.p.QueryContext(ctx, q)
    if err != nil {
        return nil, err
    }
    defer sols.Close()
    var out []map[string]string
    for sols.Next() {
        var m map[string]prolog.Term
        if err := sols.Scan(&m); err != nil {
            return nil, err
        }
        row := map[string]string{}
        for k, v := range m {
            row[k] = fmt.Sprint(v)
        }
        out = append(out, row)
    }
    if err := sols.Err(); err != nil {
        return nil, err
    }
    return out, nil
}
