package daemon

import (
	"bytes"
	"fmt"
	"sort"
	"text/template"
	"time"

	"github.com/alex-berlin-tv/nexx_omnia_go/notification"
	"github.com/alex-berlin-tv/nexx_omnia_go/omnia"
	"github.com/alex-berlin-tv/nexx_omnia_go/omnia/enums"
	"github.com/alex-berlin-tv/nexx_omnia_go/omnia/params"
	"github.com/alex-berlin-tv/radio-ingest/stackfield"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/markusmobius/go-dateparser"
)

const radioChannelId = "31543"

const newRadioIngestMessage string = `*Neue Radiodatei hochgeladen*

:pencil2: Der:die Produzent:in hat folgende Metadaten angegeben:
- Titel: _{{.Notification.Data.General.Title}}_
- Produzent.in: _{{.Notification.Data.General.SubTitle}}_
- Sendungsname: _{{.Notification.Data.General.RefNr}}_
- Beabsichtigte Veröffentlichung am: _{{.Notification.Data.General.Description}}_

{{if .ErrResults -}}
:alert: Während der Verarbeitung traten folgende(r) Fehler auf:
{{range .ErrResults -}}
- {{.}}
{{end -}}
{{end -}}
{{if .OkResults -}}
:robot: Basierend auf diesen Angaben wurde die Sendung wie folgt aufbereitet:
{{range .OkResults -}}
- {{.}}
{{end -}}
{{end -}}
{{if .ManualTasks -}}
:tick: Folgende manuelle Schritte sind notwendig:
{{range .ManualTasks -}}
- {{.}}
{{end -}}
{{end -}}

Waveform: {{.Waveform}}
`

/*
{{ if .Data.errorOccurred -}}
:warning: Es sind Fehler während der automatischen Aufbereitung aufgetreten:
{{- end }}
*/

type taskResults []taskResult

func (r taskResults) okResults() []string {
	var rsl []string
	for _, rs := range r {
		if rs.Omit || !rs.Success {
			continue
		}
		rsl = append(rsl, rs.Result)
	}
	return rsl
}

func (r taskResults) errResults() []string {
	var rsl []string
	for _, rs := range r {
		if rs.Omit || rs.Success {
			continue
		}
		rsl = append(rsl, rs.Result)
	}
	return rsl
}

func (r taskResults) manualTasks() []string {
	var rsl []string
	for _, rs := range r {
		if rs.Omit {
			continue
		}
		rsl = append(rsl, rs.ManualTasks...)
	}
	return rsl
}

type taskResult struct {
	// States whether an action was an success.
	Success bool
	// Some tasks can be omitted in the user output.
	Omit bool
	// Message to the user stating the action taken and the result of it.
	Result string
	// Instructs the user of the necessary manual tasks.
	ManualTasks []string
}

// Handles new radio uploads.
type RadioUpload struct {
	Omnia        omnia.Omnia
	Stackfield   stackfield.Room
	Notification notification.Notification
}

func NewRadioUpload(omnia omnia.Omnia, stackfield stackfield.Room, ntf notification.Notification) RadioUpload {
	return RadioUpload{
		Omnia:        omnia,
		Stackfield:   stackfield,
		Notification: ntf,
	}
}

func (u RadioUpload) Name() string {
	return "New Radio Upload"
}

func (u RadioUpload) Matches() bool {
	return u.Notification.Data.PublishingData.Origin == "uploadlink" &&
		u.Notification.Trigger.Event == "metadata" &&
		u.Notification.Item.StreamType == "audio"
}

func (u RadioUpload) OnNotification() error {
	var rsl taskResults
	rsl = append(rsl, u.handleShow())
	rsl = append(rsl, u.handleDate())
	rsl = append(rsl, u.handleChannel())
	rsl = append(rsl, u.handleSubtitleField())
	rsl = append(rsl, u.handleDescriptionField())
	if err := u.sendMessage(rsl); err != nil {
		return err
	}
	return nil
}

func (u RadioUpload) sendMessage(rsl taskResults) error {
	tpl, err := template.New("message").Parse(newRadioIngestMessage)
	if err != nil {
		return err
	}
	dt := struct {
		Notification notification.Notification
		OkResults    []string
		ErrResults   []string
		ManualTasks  []string
		Waveform     string
	}{
		Notification: u.Notification,
		OkResults:    rsl.okResults(),
		ErrResults:   rsl.errResults(),
		ManualTasks:  rsl.manualTasks(),
		Waveform:     u.Notification.Data.ImageData.Waveform,
	}
	var msg bytes.Buffer
	if err := tpl.Execute(&msg, dt); err != nil {
		return err
	}
	fmt.Println(msg.String())
	if err := u.Stackfield.Send(msg.String()); err != nil {
		return err
	}
	return nil
}

func (u RadioUpload) handleShow() taskResult {
	show, err := u.showByName(u.Notification.Data.General.RefNr)
	if err != nil {
		return taskResult{
			Success: false,
			Omit:    false,
			Result:  fmt.Sprintf("Für den angegebenen Sendungsnamen '%s' konnte keine Sendung gefunden werden", u.Notification.Data.General.RefNr),
			ManualTasks: []string{
				fmt.Sprintf("Passende Sendung für '%s' finden und entsprechend setzen", u.Notification.Data.General.RefNr),
				"Inhalt des Felds Referenznummer löschen",
			},
		}
	}
	rsl, err := u.Omnia.Update(enums.AudioStreamType, u.Notification.Data.General.ID, params.Custom{
		"show": fmt.Sprint(show.General.Id),
	})
	if err != nil && *rsl.Metadata.ErrorHint != "novalidcontext" {
		return taskResult{
			Success: false,
			Omit:    false,
			Result:  fmt.Sprintf("Beitrag konnte nicht mit Sendung '%s' verknüpft werden, %s", show.General.Title, err),
			ManualTasks: []string{
				fmt.Sprintf("Mit Sendung '%s' verbinden", show.General.Title),
				"Inhalt des Felds Referenznummer löschen",
			},
		}
	}
	_, err = u.Omnia.Update(enums.AudioStreamType, u.Notification.Data.General.ID, params.Custom{
		"refnr": "",
	})
	if err != nil {
		return taskResult{
			Success:     false,
			Omit:        false,
			Result:      "Referenznummer konnte nicht zurückgesetzt werden",
			ManualTasks: []string{"Inhalt des Felds Referenznummer löschen"},
		}
	}
	return taskResult{
		Success:     true,
		Result:      fmt.Sprintf("Wurde der Sendung '%s' zugeordnet", show.General.Title),
		ManualTasks: []string{},
	}
}

func (u RadioUpload) showByName(name string) (*omnia.MediaResultItem, error) {
	rsp, err := u.Omnia.All(enums.ShowStreamType, params.Basic{})
	if err != nil {
		return nil, fmt.Errorf("failed to get list of shows, %s", err)
	}
	rsl := *rsp.Result
	var shows = make(map[string]omnia.MediaResultItem)
	var showNames = []string{}
	for _, show := range rsl {
		shows[show.General.Title] = show
		showNames = append(showNames, show.General.Title)
	}
	matches := fuzzy.RankFindNormalizedFold(name, showNames)
	sort.Sort(matches)
	if len(matches) != 0 {
		tmp := shows[matches[0].Source]
		return &tmp, nil
	}
	return nil, fmt.Errorf("not found")
}

func (u RadioUpload) handleDate() taskResult {
	cfg := dateparser.Configuration{
		DateOrder:   dateparser.DMY,
		Languages:   []string{"de"},
		CurrentTime: time.Now(),
	}
	_, rsl, err := dateparser.Search(&cfg, u.Notification.Data.General.Description)
	if err != nil || len(rsl) == 0 {
		return taskResult{
			Success: false,
			Omit:    false,
			Result:  "Das Sendedatum konnte nicht aus dem Beschreibungsfeld entnommen werden",
			ManualTasks: []string{
				"Das Sendedatum setzen",
			},
		}
	}
	date := rsl[0].Date.Time
	_, err = u.Omnia.Update(enums.AudioStreamType, u.Notification.Data.General.ID, params.Custom{
		"releasedate": fmt.Sprint(date.Unix()),
	})
	if err != nil {
		return taskResult{
			Success: false,
			Omit:    false,
			Result:  fmt.Sprintf("Veröffentlichungsdatum konnte nicht gesetzt werden, %s", err),
			ManualTasks: []string{
				fmt.Sprintf("Veröffentlichungsdatum auf %s setzen", date),
			},
		}
	}
	_, err = u.Omnia.Update(enums.AudioStreamType, u.Notification.Data.General.ID, params.Custom{
		"description": "",
	})
	if err != nil {
		return taskResult{
			Success: false,
			Omit:    false,
			Result:  "Beschreibung konnte nicht gelöscht werden",
		}
	}
	return taskResult{
		Success: true,
		Omit:    false,
		Result:  fmt.Sprintf("Veröffentlichungsdatum wurde auf %s gesetzt", date),
	}
}

func (u RadioUpload) handleChannel() taskResult {
	_, err := u.Omnia.Update(enums.AudioStreamType, u.Notification.Data.General.ID, params.Custom{
		"channel": radioChannelId,
	})
	if err != nil {
		return taskResult{
			Success:     false,
			Omit:        false,
			Result:      "Channel konnte nicht auf Radio gesetzt werden",
			ManualTasks: []string{"Channel auf Radio setzten"},
		}
	}
	return taskResult{
		Success: true,
		Omit:    true,
		Result:  "Channel wurde auf Radio gesetzt",
	}
}

func (u RadioUpload) handleSubtitleField() taskResult {
	_, err := u.Omnia.Update(enums.AudioStreamType, u.Notification.Data.General.ID, params.Custom{
		"alttitle": u.Notification.Data.General.SubTitle,
	})
	if err != nil {
		return taskResult{
			Success: false,
			Result:  "Produzent:innen konnten nicht in das Feld »Alternativer Titel« übertragen werden",
			ManualTasks: []string{
				"Produzent:innen in das Feld »Alternativer Titel« übertragen",
				"Inhalt des Felds »Untertitel« löschen",
			},
		}
	}
	_, err = u.Omnia.Update(enums.AudioStreamType, u.Notification.Data.General.ID, params.Custom{
		"subtitle": "",
	})
	if err != nil {
		return taskResult{
			Success: false,
			Result:  "Inhalt des Felds »Untertitel« konnte nicht gelöscht werden",
			ManualTasks: []string{
				"Inhalt des Felds »Untertitel« löschen",
			},
		}
	}
	return taskResult{
		Success: true,
		Omit:    true,
		Result:  "Inhalt aus dem Feld »Untertitel« nach »Alternativer Titel« übertragen",
	}
}

func (u RadioUpload) handleDescriptionField() taskResult {
	_, err := u.Omnia.Update(enums.AudioStreamType, u.Notification.Data.General.ID, params.Custom{
		"altdescription": u.Notification.Data.General.Description,
	})
	if err != nil {
		return taskResult{
			Success: false,
			Result:  "Angaben aus dem Feld »Beschreibung« konnten nicht in das Feld »Alternative Beschreibung« übertragen werden",
			ManualTasks: []string{
				"Inhalt von »Beschreibung« nach »Alternative Beschreibung« übertragen",
			},
		}
	}
	_, err = u.Omnia.Update(enums.AudioStreamType, u.Notification.Data.General.ID, params.Custom{
		"description": "",
	})
	if err != nil {
		return taskResult{
			Success: false,
			Result:  "Inhalt des Felds »Beschreibung« konnte nicht gelöscht werden",
			ManualTasks: []string{
				"Inhalt des Felds »Beschreibung« löschen",
			},
		}
	}
	return taskResult{
		Success: true,
		Omit:    true,
		Result:  "Inhalt aus dem Feld »Beschreibung« nach »Alternative Beschreibung« übertragen",
	}
}
