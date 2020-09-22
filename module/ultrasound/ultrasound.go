package ultrasound

import (
	"context"
	"fmt"

	"github.com/dantin/cubit/log"
	"github.com/dantin/cubit/module/xep0030"
	"github.com/dantin/cubit/router"
	"github.com/dantin/cubit/util/runqueue"
	"github.com/dantin/cubit/xmpp"
)

const ultrasoundNamespace = "hc:apm:ultrasound"

// Config represents customized ultrasound module configuration.
type Config struct {
	PageSize int `yaml:"page_size"`
}

// Ultrasound represents a ultrasound module.
type Ultrasound struct {
	cfg      *Config
	router   router.Router
	runQueue *runqueue.RunQueue
}

// New returns a ultrasound IQ handler module.
func New(config *Config, disco *xep0030.DiscoInfo, router router.Router) *Ultrasound {
	v := &Ultrasound{
		cfg:      config,
		router:   router,
		runQueue: runqueue.New("ultrasound"),
	}
	if disco != nil {
		disco.RegisterServerFeature(ultrasoundNamespace)
	}
	return v
}

// MatchesIQ returns whether or not an IQ should be processed by the ultrasound module.
func (x *Ultrasound) MatchesIQ(iq *xmpp.IQ) bool {
	return iq.IsGet() && iq.Elements().ChildNamespace("query", ultrasoundNamespace) != nil && iq.ToJID().IsServer()
}

// ProcessIQ process a ultrasound IQ talking according action over the associated stream.
func (x *Ultrasound) ProcessIQ(ctx context.Context, iq *xmpp.IQ) {
	x.runQueue.Run(func() {
		x.processIQ(ctx, iq)
	})
}

// Shutdown shuts down ultrasound module.
func (x *Ultrasound) Shutdown() error {
	c := make(chan struct{})
	x.runQueue.Stop(func() { close(c) })
	<-c
	return nil
}

func (x *Ultrasound) processIQ(ctx context.Context, iq *xmpp.IQ) {
	q := iq.Elements().ChildNamespace("query", ultrasoundNamespace)
	if q == nil || q.Elements().Count() != 0 {
		_ = x.router.Route(ctx, iq.BadRequestError())
		return
	}
	x.sendVideoStream(ctx, iq)
}

func (x *Ultrasound) sendVideoStream(ctx context.Context, iq *xmpp.IQ) {
	userJID := iq.FromJID()
	username := userJID.Node()
	log.Infof("retrieving video stream for %s", username)

	result := iq.ResultIQ()
	query := xmpp.NewElementNamespace("query", ultrasoundNamespace)

	rooms := xmpp.NewElementName("rooms")
	query.AppendElement(rooms)
	for i := 0; i < x.cfg.PageSize; i++ {
		room := xmpp.NewElementName("room")
		name := xmpp.NewElementName("name")
		name.SetText(fmt.Sprintf("name %d", i))
		room.AppendElement(name)

		videoStream := xmpp.NewElementName("video_stream")
		videoStream.SetText(fmt.Sprintf("video stream %d", i))
		room.AppendElement(videoStream)

		rooms.AppendElement(room)
	}
	result.AppendElement(query)
	_ = x.router.Route(ctx, result)
}
