package ultrasound

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dantin/cubit/log"
	"github.com/dantin/cubit/model"
	"github.com/dantin/cubit/module/xep0030"
	"github.com/dantin/cubit/router"
	"github.com/dantin/cubit/storage/repository"
	"github.com/dantin/cubit/util/runqueue"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
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
	roomRep  repository.Room
}

// New returns a ultrasound IQ handler module.
func New(config *Config, disco *xep0030.DiscoInfo, router router.Router, userRep repository.User, roomRep repository.Room) *Ultrasound {
	v := &Ultrasound{
		cfg:      config,
		router:   router,
		runQueue: runqueue.New("ultrasound"),
		userRep:  userRep,
		roomRep:  roomRep,
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
	room := e.ChildNamespace("room", ultrasoundNamespace)
	qc := e.ChildNamespace("qc", ultrasoundNamespace)
	return (iq.IsGet() && (profile != nil || rooms != nil || room != nil || qc != nil))
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
		x.sendRooms(ctx, iq)
	} else if room := e.ChildNamespace("room", ultrasoundNamespace); room != nil {
		x.sendRoom(ctx, iq)
	} else if qc := e.ChildNamespace("qc", ultrasoundNamespace); qc != nil {
		x.sendQCStream(ctx, iq)
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
		_ = x.router.Route(ctx, iq.ItemNotFoundError())
		return
	}

	if user.Role == model.Unknown {
		log.Errorf("empty role for user %s", username)
		_ = x.router.Route(ctx, iq.InternalServerError())
		return
	}

	result := iq.ResultIQ()
	profile := xmpp.NewElementNamespace("profile", ultrasoundNamespace)
	profile.SetText(user.Role.String())
	result.AppendElement(profile)

	_ = x.router.Route(ctx, result)
}

func (x *Ultrasound) sendRooms(ctx context.Context, iq *xmpp.IQ) {
	userJID := iq.FromJID()
	username := userJID.Node()

	log.Debugf("retrieving video stream for %s", username)

	user, err := x.userRep.FetchUser(ctx, username)
	if err != nil {
		log.Error(err)
		_ = x.router.Route(ctx, iq.ItemNotFoundError())
		return
	}
	if user.Role != model.Admin {
		_ = x.router.Route(ctx, iq.ForbiddenError())
		return
	}

	req := iq.Elements().ChildNamespace("rooms", ultrasoundNamespace)
	if req == nil {
		log.Errorf("no rooms element.")
		_ = x.router.Route(ctx, iq.InternalServerError())
		return
	}
	var page, size int
	pageValue := req.Attributes().Get("page")
	if len(pageValue) > 0 {
		if page, err = strconv.Atoi(pageValue); err != nil {
			page = 0
		}
	} else {
		page = 0
	}
	sizeValue := req.Attributes().Get("size")
	if len(sizeValue) > 0 {
		if size, err = strconv.Atoi(sizeValue); err != nil {
			size = 4
		}
	} else {
		size = 4
	}
	rooms, err := x.roomRep.FetchRooms(ctx, page, size)
	if err != nil {
		log.Error(err)
		_ = x.router.Route(ctx, iq.InternalServerError())
		return
	}
	total, err := x.roomRep.CountRooms(ctx)
	if err != nil {
		log.Error(err)
		_ = x.router.Route(ctx, iq.InternalServerError())
		return
	}

	result := iq.ResultIQ()
	roomsNode := xmpp.NewElementNamespace("rooms", ultrasoundNamespace)
	roomsNode.SetAttribute("page", fmt.Sprintf("%d", page))
	roomsNode.SetAttribute("pages", fmt.Sprintf("%d", total/size))
	roomsNode.SetAttribute("size", fmt.Sprintf("%d", size))
	for _, room := range rooms {
		roomNode := xmpp.NewElementName("room")
		roomJID, err := jid.New(room.Username, userJID.Domain(), userJID.Resource(), true)
		if err != nil {
			log.Error(err)
			_ = x.router.Route(ctx, iq.InternalServerError())
			return
		}
		roomNode.SetAttribute("jid", roomJID.String())
		roomNode.SetAttribute("room_id", fmt.Sprintf("%d", room.ID))
		roomNode.SetAttribute("name", room.Name)

		for _, video := range room.Streams {
			videoNode := xmpp.NewElementName("video_stream")
			videoNode.SetAttribute("type", video.Type.String())
			videoNode.SetText(video.Stream)
			roomNode.AppendElement(videoNode)
		}
		roomsNode.AppendElement(roomNode)
	}
	result.AppendElement(roomsNode)
	_ = x.router.Route(ctx, result)
}

func (x *Ultrasound) sendRoom(ctx context.Context, iq *xmpp.IQ) {
	userJID := iq.FromJID()
	username := userJID.Node()

	log.Debugf("retrieving video stream for %s", username)

	user, err := x.userRep.FetchUser(ctx, username)
	if err != nil {
		log.Error(err)
		_ = x.router.Route(ctx, iq.ItemNotFoundError())
		return
	}
	if user.Role != model.Usr {
		_ = x.router.Route(ctx, iq.ForbiddenError())
		return
	}

	room, err := x.roomRep.FetchRoom(ctx, username)
	if err != nil {
		log.Error(err)
		_ = x.router.Route(ctx, iq.InternalServerError())
		return
	}

	result := iq.ResultIQ()
	roomNode := xmpp.NewElementNamespace("room", ultrasoundNamespace)
	roomJID, err := jid.New(room.Username, userJID.Domain(), userJID.Resource(), true)
	if err != nil {
		log.Error(err)
		_ = x.router.Route(ctx, iq.InternalServerError())
		return
	}
	roomNode.SetAttribute("jid", roomJID.String())
	roomNode.SetAttribute("room_id", fmt.Sprintf("%d", room.ID))
	roomNode.SetAttribute("name", room.Name)

	for _, video := range room.Streams {
		videoNode := xmpp.NewElementName("video_stream")
		videoNode.SetAttribute("type", video.Type.String())
		videoNode.SetText(video.Stream)
		roomNode.AppendElement(videoNode)
	}

	result.AppendElement(roomNode)
	_ = x.router.Route(ctx, result)
}

func (x *Ultrasound) sendQCStream(ctx context.Context, iq *xmpp.IQ) {
	userJID := iq.FromJID()
	username := userJID.Node()

	log.Debugf("retrieving video stream for %s", username)

	user, err := x.userRep.FetchUser(ctx, username)
	if err != nil {
		log.Error(err)
		_ = x.router.Route(ctx, iq.ItemNotFoundError())
		return
	}
	if user.Role != model.Usr {
		_ = x.router.Route(ctx, iq.ForbiddenError())
		return
	}

	video, err := x.roomRep.FetchQCStream(ctx, username)
	if err != nil {
		log.Error(err)
		_ = x.router.Route(ctx, iq.InternalServerError())
		return
	}

	result := iq.ResultIQ()
	qcNode := xmpp.NewElementNamespace("qc", ultrasoundNamespace)
	qcNode.SetText(video.Stream)
	result.AppendElement(qcNode)
	_ = x.router.Route(ctx, result)
}
