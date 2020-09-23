package ultrasound

import (
	"context"
	"fmt"

	"github.com/dantin/cubit/log"
	"github.com/dantin/cubit/module/xep0030"
	"github.com/dantin/cubit/router"
	"github.com/dantin/cubit/storage/repository"
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
	userRep  repository.User
}

// New returns a ultrasound IQ handler module.
func New(config *Config, disco *xep0030.DiscoInfo, router router.Router, userRep repository.User) *Ultrasound {
	v := &Ultrasound{
		cfg:      config,
		router:   router,
		runQueue: runqueue.New("ultrasound"),
		userRep:  userRep,
	}
	if disco != nil {
		disco.RegisterServerFeature(ultrasoundNamespace)
	}
	return v
}

// MatchesIQ returns whether or not an IQ should be processed by the ultrasound module.
func (x *Ultrasound) MatchesIQ(iq *xmpp.IQ) bool {
	e := iq.Elements()
	profile := e.ChildNamespace("profile", ultrasoundNamespace)
	rooms := e.ChildNamespace("rooms", ultrasoundNamespace)
	return (iq.IsGet() && (profile != nil || rooms != nil))
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
	e := iq.Elements()
	if profile := e.ChildNamespace("profile", ultrasoundNamespace); profile != nil {
		x.sendProfile(ctx, iq)
	} else if rooms := e.ChildNamespace("rooms", ultrasoundNamespace); rooms != nil {
		x.sendVideoStream(ctx, iq)
	} else {
		_ = x.router.Route(ctx, iq.BadRequestError())
	}
}

func (x *Ultrasound) sendProfile(ctx context.Context, iq *xmpp.IQ) {
	userJID := iq.FromJID()
	username := userJID.Node()

	log.Debugf("fetch %s's profile", username)

	user, err := x.userRep.FetchUser(ctx, username)
	if err != nil {
		log.Error(err)
		_ = x.router.Route(ctx, iq.InternalServerError())
		return
	}

	if len(user.Role) == 0 {
		log.Errorf("empty role field for user %s", username)
		_ = x.router.Route(ctx, iq.InternalServerError())
		return
	}

	result := iq.ResultIQ()
	query := xmpp.NewElementNamespace("query", ultrasoundNamespace)
	profile := xmpp.NewElementName("profile")
	query.AppendElement(profile)
	profile.SetText(user.Role)
	result.AppendElement(query)

	_ = x.router.Route(ctx, result)
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
