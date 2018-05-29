package stores

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dave/flux"
	"github.com/dave/frizz/models"
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
)

type App struct {
	Dispatcher flux.DispatcherInterface
	Watcher    flux.WatcherInterface
	Notifier   flux.NotifierInterface

	Empty      *EmptyStore
	Page       *PageStore
	Injector   *InjectorStore
	Connection *ConnectionStore
	Types      *TypesStore

	externalM sync.RWMutex
	external  map[models.Id]flux.StoreInterface
}

func (a *App) Init() {

	n := flux.NewNotifier()
	a.Notifier = n
	a.Watcher = n
	a.external = map[models.Id]flux.StoreInterface{}

	a.Empty = NewEmptyStore(a)
	a.Page = NewPageStore(a)
	a.Injector = NewInjectorStore(a)
	a.Connection = NewConnectionStore(a)
	a.Types = NewTypesStore(a)

	a.Dispatcher = flux.NewDispatcher(
		// Notifier:
		a.Notifier,
		// Stores:
		a.Empty,
		a.Page,
		a.Injector,
		a.Connection,
		a.Types,
	)
}

func (a *App) Dispatch(action flux.ActionInterface) chan struct{} {
	return a.Dispatcher.Dispatch(action)
}

func (a *App) Watch(key interface{}, f func(done chan struct{})) {
	a.Watcher.Watch(key, f)
}

func (a *App) Delete(key interface{}) {
	a.Watcher.Delete(key)
}

func (a *App) Fail(err error) {
	// TODO: improve this
	js.Global.Call("alert", err.Error())
}

func (a *App) Debug(message ...interface{}) {
	js.Global.Get("console").Call("log", message...)
}

var lastLog *struct{}

// LogHide hides the message after 2 seconds
func (a *App) LogHide(args ...interface{}) {
	a.Log(args...)
	if len(args) > 0 {
		// clear message after 2 sec if not changed
		before := lastLog
		go func() {
			<-time.After(time.Second * 2)
			if before == lastLog {
				a.Log()
			}
		}()
	}
}

func (a *App) Log(args ...interface{}) {
	m := dom.GetWindow().Document().GetElementByID("message")
	var message string
	if len(args) > 0 {
		message = strings.TrimSuffix(fmt.Sprintln(args...), "\n")
	}
	if m.InnerHTML() != message {
		if message != "" {
			js.Global.Get("console").Call("log", "Status", strconv.Quote(message))
		}
		requestAnimationFrame()
		m.SetInnerHTML(message)
		requestAnimationFrame()
		lastLog = &struct{}{}
	}
}

func (a *App) Logf(format string, args ...interface{}) {
	a.Log(fmt.Sprintf(format, args...))
}

func (a *App) LogHidef(format string, args ...interface{}) {
	a.LogHide(fmt.Sprintf(format, args...))
}

func requestAnimationFrame() {
	c := make(chan struct{})
	js.Global.Call("requestAnimationFrame", func() { close(c) })
	<-c
}

func (a *App) RegisterExternalStore(id models.Id, app *App) {
	f := GetExternalStoreFunc(id)
	if f == nil {
		return
	}
	a.externalM.Lock()
	defer a.externalM.Unlock()
	a.external[id] = f(app)
}

func (a *App) ExternalStore(id models.Id) flux.StoreInterface {
	a.externalM.Lock()
	defer a.externalM.Unlock()
	return a.external[id]
}

var externalStoreFuncsM sync.RWMutex
var externalStoreFuncs = map[models.Id]StoreFunc{}

type StoreFunc func(a *App) flux.StoreInterface

func RegisterExternalStoreFunc(id models.Id, store StoreFunc) {
	externalStoreFuncsM.Lock()
	defer externalStoreFuncsM.Unlock()
	externalStoreFuncs[id] = store
}

func GetExternalStoreFunc(id models.Id) StoreFunc {
	externalStoreFuncsM.RLock()
	defer externalStoreFuncsM.RUnlock()
	return externalStoreFuncs[id]
}
