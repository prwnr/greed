package pkg

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Listener struct
type Listener struct {
	Items                   map[string]*StreamListener
	newListenerHandlers     []func(name string)
	listenerChangedHandlers []func(listener StreamListener)
	artisan                 *Artisan
}

// NewListener creates listener with artisan command.
func NewListener() (*Listener, error) {
	artisan := NewArtisan()

	_, _, err := artisan.Exec("list", "streamer")
	if err != nil {
		return nil, errors.New("artisan not detected")
	}

	listener := &Listener{
		artisan: artisan,
	}

	return listener, nil
}

// Listen starts listening via Artisan command call and adds output to the stack.
// Restarts command listening when it returns error code 1.
func (l *Listener) Listen(stream Stream) {
	lis := l.AddStreamListener(stream.Name)
	if lis.stopped {
		return
	}

	messages := stream.GetMessagesList()
	lastID := "0-0"
	if len(messages) > 0 {
		lastID = messages[len(messages)-1]
	}

	id := fmt.Sprintf("--last_id=%s", lastID)
	args := []string{"streamer:listen", stream.Name, id}

	for {
		cmd, err := l.artisan.ExecPipe(func(output string, cmd *exec.Cmd) error {
			if lis.stopped {
				return errors.New("stopped")
			}

			lis.Output = append(lis.Output, fmt.Sprintf("%s: %s", time.Now().Format("01-02-2006 15:04:05"), output))
			if lis.HasNoListeners(output) {
				lis.stopped = true
				l.emitListenerChanged(*lis)
				return errors.New("stopped")
			}

			if lis.IsFailing(output) {
				lis.warning = true
				l.emitListenerChanged(*lis)
			}

			return nil
		}, args...)

		if lis.stopped {
			return
		}

		if err != nil {
			return
		}

		code := cmd.ProcessState.ExitCode()
		if code == 1 {
			lis.error = true
			l.emitListenerChanged(*lis)
			args = []string{"streamer:listen", stream.Name}
			continue
		}
	}
}

func (l *Listener) AddStreamListener(name string) *StreamListener {
	if l.Items == nil {
		l.Items = make(map[string]*StreamListener)
	}

	lis, ok := l.Items[name]
	if ok {
		return lis
	}

	lis = &StreamListener{
		Name:   name,
		Output: nil,
	}

	l.Items[name] = lis
	l.emitNewListener(name)

	return lis
}

func (l *Listener) OnNewListener(handle func(a string)) {
	l.newListenerHandlers = append(l.newListenerHandlers, handle)
}

func (l *Listener) emitNewListener(name string) {
	for _, h := range l.newListenerHandlers {
		h(name)
	}
}

func (l *Listener) OnListenerChange(handle func(listener StreamListener)) {
	l.listenerChangedHandlers = append(l.listenerChangedHandlers, handle)
}

func (l *Listener) emitListenerChanged(listener StreamListener) {
	for _, h := range l.listenerChangedHandlers {
		h(listener)
	}
}

type StreamListener struct {
	Name    string
	Output  []string
	stopped bool
	warning bool
	error   bool
}

func (s StreamListener) ParseOutput() string {
	var content string
	for _, i := range s.Output {
		content += fmt.Sprintf("%s", i)
	}

	return content
}

func (s StreamListener) HasNoListeners(output string) bool {
	return output == fmt.Sprintf("There are no local listeners associated with %s event in configuration.\n", s.Name)
}

func (s StreamListener) IsFailing(output string) bool {
	return strings.Contains(output, "Listener error. Failed processing message")
}

// Status of StreamListener as a formatted string.
func (s StreamListener) Status() string {
	if s.error {
		return "[red]WARNING[red]"
	}

	if s.warning {
		return "[yellow]WARNING[yellow]"
	}

	if s.stopped {
		return "[grey]STOPPED[grey]"
	}

	return "[green]OK[green]"
}