package daemon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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
	Omnia      omnia.Omnia
	Port       int
	recordPath string
}

// Returns a new [Daemon] instance based on the given configuration.
func NewDaemon(cfg config.Config) Daemon {
	return Daemon{
		Omnia: omnia.NewOmnia(cfg.DomainId, cfg.ApiSecret, cfg.SessionId),
		Port:  cfg.Port,
	}
}

// Listen for notifications and writes them to a JSON file.
func (d *Daemon) Record(path string) {
	d.recordPath = path
	d.startRouter(d.recordHandler)
}

// Run the daemon.
func (d Daemon) Run() {
	d.startRouter(d.defaultNotificationHandler)
}

// Test the notification handling with a pre-recorded notification body. Takes
// the path to the JSON file with the data as argument.
//
// Use the [Daemon.Record] method for record new notifications.
func (d Daemon) TestRun(path string) error {
	dt, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return d.onNotification(dt)
}

func (d Daemon) startRouter(handler func(http.ResponseWriter, *http.Request)) {
	rtr := chi.NewRouter()
	rtr.Use(middleware.Logger)
	rtr.Post("/", handler)
	logrus.Infof("Will listen for Omnia on :%d", d.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", d.Port), rtr)
}

func (d Daemon) defaultNotificationHandler(w http.ResponseWriter, r *http.Request) {
	dt, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.Fatal(err)
	}
	if err := d.onNotification(dt); err != nil {
		logrus.Fatal(err)
	}
}

func (d Daemon) recordHandler(w http.ResponseWriter, r *http.Request) {
	dt, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	var rec = make(Record)
	if err := json.Unmarshal(dt, &rec); err != nil {
		logrus.Error(err)
	}
	rec.SaveToJson(d.recordPath)
	os.Exit(0)
}

func (d Daemon) onNotification(body []byte) error {
	fmt.Println(string(body))
	data, err := notification.NotificationFromJson(body)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", data)
	return nil
}
