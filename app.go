package main

import (
	"fmt"
	"sync"
)

type App struct {
	lock sync.RWMutex

	version         int
	Commands        CommandLog
	commandHandlers []CommandHandler
	queryHandlers   []QueryHandler
}

func NewApp(log CommandLog) *App {
	return &App{
		version:         0,
		Commands:        log,
		commandHandlers: []CommandHandler{&SkipHandler{}},
		queryHandlers:   []QueryHandler{},
	}
}

func (app *App) Mount(m interface{}) *App {
	if commandHandler, ok := m.(CommandHandler); ok {
		app.commandHandlers = append(app.commandHandlers, commandHandler)
	}

	if queryHandler, ok := m.(QueryHandler); ok {
		app.queryHandlers = append(app.queryHandlers, queryHandler)
	}

	return app
}

func (app *App) Replay(skipErrors bool) error {
	app.lock.Lock()
	defer app.lock.Unlock()

	commands, err := app.Commands.After(app.version)
	if err != nil {
		return fmt.Errorf("failed to replay commands: %w", err)
	}
	for command := range commands {
		for _, handler := range app.commandHandlers {
			err := handler.HandleCommand(command.Message)
			if err == ErrCommandNotAccepted {
				continue
			}
			if err != nil {
				fmt.Printf("failed to replay command: %s", err)
				if !skipErrors {
					return err
				}
			}
		}
		app.version = command.ID
	}
	return nil
}

func (app *App) HandleCommand(message Command) error {
	app.lock.Lock()
	defer app.lock.Unlock()

	for _, handler := range app.commandHandlers {
		err := handler.HandleCommand(message)
		if err == ErrCommandNotAccepted {
			continue
		}
		if err == nil {
			break
		}
		if err != ErrCommandNotAccepted {
			return fmt.Errorf("failed to handle command: %w", err)
		}
	}

	if err := app.Commands.Append(message); err != nil {
		return fmt.Errorf("failed to append command: %w", err)
	}

	app.version += 1

	return nil
}

func (app *App) HandleQuery(query Query) error {
	app.lock.RLock()
	defer app.lock.RUnlock()

	for _, handler := range app.queryHandlers {
		err := handler.HandleQuery(query)
		if err == ErrQueryNotAccepted {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to handle query: %w", err)
		}
		return nil
	}
	return ErrQueryNotAccepted
}

func (app *App) ExposeState(globals map[string]any) {
	for _, handler := range app.queryHandlers {
		if state, ok := handler.(interface{ Inspect() map[string]any }); ok {
			for key, value := range state.Inspect() {
				globals[key] = value
			}
		}
	}
}
