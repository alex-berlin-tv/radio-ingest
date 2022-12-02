package daemon

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/alex-berlin-tv/nexx_omnia_go/notification"
	"github.com/alex-berlin-tv/nexx_omnia_go/omnia"
	"github.com/alex-berlin-tv/radio-ingest/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

// Listens to the Omnia's notification gateway and handles incoming radio
// uploads.
type Daemon struct {
	Omnia omnia.Omnia
	Port  int
}

func NewDaemon(cfg config.Config) Daemon {
	return Daemon{
		Omnia: omnia.NewOmnia(cfg.DomainId, cfg.ApiSecret, cfg.SessionId),
		Port:  cfg.Port,
	}
}

func (d Daemon) Run() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/", d.handleNotification)
	http.ListenAndServe(fmt.Sprintf(":%d", d.Port), router)
}

func (d Daemon) handleNotification(w http.ResponseWriter, r *http.Request) {
	dt, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	data, err := notification.NotificationFromJson(dt)
	if err != nil {
		logrus.Error(err)
	}
	fmt.Printf("%+v\n", data)
}
