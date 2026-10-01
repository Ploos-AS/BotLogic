package logic

import (
    "context"
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
        var row map[string]string
        if err := sols.Scan(&row); err != nil {
            return nil, err
        }
        out = append(out, row)
    }
    if err := sols.Err(); err != nil {
        return nil, err
    }
    return out, nil
}
