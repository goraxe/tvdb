package tvdb

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

const (
	apiKey       = "90D7DF3AE9E4841E"
	testUser     = "34A8615ABE815874"
	simpsonsID   = 71663
	simpsonsIMDB = "tt0096697"
	simpsonsName = "The Simpsons"
	futuramaID   = 73871
)

var ()

var (
	mux     *http.ServeMux
	server  *httptest.Server
	handler *fileHandler
)

type fileHandler struct {
	io.ReadCloser
}

func newFileHandler(filename string) *fileHandler {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	return &fileHandler{
		ReadCloser: f,
	}
}

func (h *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Support zip?
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	io.Copy(w, h)
}

func setup() *Client {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client := NewClient(apiKey)
	client.BaseURL, _ = url.Parse(server.URL)
	return client
}

func teardown() {
	server.Close()
	handler.Close()
}

type values map[string]string

func testFormValues(t *testing.T, r *http.Request, values values) {
	want := url.Values{}
	for k, v := range values {
		want.Add(k, v)
	}

	r.ParseForm()
	if got := r.Form; !reflect.DeepEqual(got, want) {
		t.Errorf("Request parameters: %v, want %v", got, want)
	}
}

func TestLanguages(t *testing.T) {
	client := setup()
	defer teardown()

	handler = newFileHandler("testdata/languages.xml")
	mux.Handle(fmt.Sprintf("/api/%s/languages.xml", apiKey), handler)

	langs, err := client.Languages()
	if err != nil {
		t.Fatal(err)
	}

	if len(langs) != 23 {
		t.Errorf("TestLanguage: Incorrect number of lanuages. Expected '23' got '%d'", len(langs))
	}

	want := Language{ID: 7, Abbr: "en", Name: "English"}
	for _, lang := range langs {
		if lang.Abbr == "en" {
			if !reflect.DeepEqual(lang, want) {
				t.Errorf("Language 'en' does not match.  \n%s", pretty.Compare(want, lang))
			}
			return
		}
	}

	t.Errorf("TestLanguage: Couldn't find english in languges")
}

func TestSearchSeries(t *testing.T) {
	client := setup()
	defer teardown()

	handler = newFileHandler(`testdata/GetSeries.php?seriesname=The%20Simpsons`)
	mux.HandleFunc("/api/GetSeries.php", func(w http.ResponseWriter, r *http.Request) {
		testFormValues(t, r, values{
			"language":   "en",
			"seriesname": "The Simpsons",
		})
		handler.ServeHTTP(w, r)
	})

	series, err := client.SearchSeries("The Simpsons", "en")
	if err != nil {
		t.Fatal(err)
	}

	if len(series) != 2 {
		t.Errorf("TestSearchSeries: Incorrect number of series. Expected '2' got '%d'", len(series))
	}

	want := SeriesSummary{
		Aliases: PipeList(nil),
		seriesShared: seriesShared{
			ID: 71663, Language: "en",
			Name:       "The Simpsons",
			BannerPath: "graphical/71663-g13.jpg",
			Overview:   "Set in Springfield, the average American town, the show focuses on the antics and everyday adventures of the Simpson family; Homer, Marge, Bart, Lisa and Maggie, as well as a virtual cast of thousands. Since the beginning, the series has been a pop culture icon, attracting hundreds of celebrities to guest star. The show has also made name for itself in its fearless satirical take on politics, media and American life in general.",
			FirstAired: "1989-12-17",
			IMDBID:     "tt0096697",
			Zap2itID:   "EP00018693",
			Network:    "FOX",
		},
	}

	if !reflect.DeepEqual(series[0], want) {
		t.Errorf("First series does not match.  \n%s", pretty.Compare(want, series[0]))
	}
}

func TestSeriesByID(t *testing.T) {
	client := setup()
	defer teardown()

	handler = newFileHandler("testdata/series_71663_en.xml")
	mux.Handle(fmt.Sprintf("/api/%s/series/71663/en.xml", apiKey), handler)

	series, err := client.SeriesByID(71663, "en")
	if err != nil {
		t.Fatal(err)
	}

	want := &Series{
		Actors:        PipeList{"Dan Castellaneta", "Hank Azaria", "Harry Shearer", "Marcia Wallace", "Julie Kavner", "Yeardley Smith", "Nancy Cartwright", "Anne Hathaway"},
		AirsDayOfWeek: "Sunday",
		AirsTime:      "8:00 PM",
		ContentRating: "TV-PG",
		Genre:         PipeList{"Animation", "Comedy"},
		Network:       "FOX",
		Rating:        "9.0",
		RatingCount:   "542",
		Runtime:       "30",
		Status:        "Continuing",
		Added:         "",
		AddedBy:       "",
		FanartPath:    "fanart/original/71663-31.jpg",
		LastUpdated:   "1422395198",
		PostersPath:   "",
		seriesShared: seriesShared{
			ID:         71663,
			Language:   "",
			Name:       "The Simpsons",
			BannerPath: "graphical/71663-g13.jpg",
			Overview:   "Set in Springfield, the average American town, the show focuses on the antics and everyday adventures of the Simpson family; Homer, Marge, Bart, Lisa and Maggie, as well as a virtual cast of thousands. Since the beginning, the series has been a pop culture icon, attracting hundreds of celebrities to guest star. The show has also made name for itself in its fearless satirical take on politics, media and American life in general.",
			FirstAired: "1989-12-17",
			IMDBID:     "tt0096697",
			Zap2itID:   "EP00018693",
			Network:    ""},
	}

	if !reflect.DeepEqual(series, want) {
		t.Errorf("Response does not match.  \n%s", pretty.Compare(want, series))
	}
}

func TestSeriesByRemoteID(t *testing.T) {
	client := setup()
	defer teardown()

	handler = newFileHandler(`testdata/GetSeriesByRemoteID.php?imdbid=tt0096697&language=en`)
	mux.HandleFunc("/api/GetSeriesByRemoteID.php", func(w http.ResponseWriter, r *http.Request) {
		testFormValues(t, r, values{
			"language": "en",
			"imdbid":   "tt0096697",
		})
		handler.ServeHTTP(w, r)
	})

	series, err := client.SeriesByRemoteID(IMDB, "tt0096697", "en")
	if err != nil {
		t.Fatal(err)
	}

	want := &SeriesSummary{
		Aliases: nil,
		seriesShared: seriesShared{
			ID: 71663, Language: "en",
			Name:       "The Simpsons",
			BannerPath: "graphical/71663-g13.jpg",
			Overview:   "Set in Springfield, the average American town, the show focuses on the antics and everyday adventures of the Simpson family; Homer, Marge, Bart, Lisa and Maggie, as well as a virtual cast of thousands. Since the beginning, the series has been a pop culture icon, attracting hundreds of celebrities to guest star. The show has also made name for itself in its fearless satirical take on politics, media and American life in general.",
			FirstAired: "1989-12-17",
			IMDBID:     "tt0096697",
			Zap2itID:   "EP00018693",
		},
	}

	if !reflect.DeepEqual(series, want) {
		t.Errorf("Response does not match.  \n%s", pretty.Compare(want, series))
	}
}

func TestSeriesAllByID(t *testing.T) {
	client := setup()
	defer teardown()

	handler = newFileHandler("testdata/series_71663_all_en.xml")
	mux.Handle(fmt.Sprintf("/api/%s/series/71663/all/en.xml", apiKey), handler)

	series, episodes, err := client.SeriesAllByID(71663, "en")
	if err != nil {
		t.Fatal(err)
	}

	want := &Series{
		Actors:        PipeList{"Dan Castellaneta", "Hank Azaria", "Harry Shearer", "Marcia Wallace", "Julie Kavner", "Yeardley Smith", "Nancy Cartwright", "Anne Hathaway"},
		AirsDayOfWeek: "Sunday",
		AirsTime:      "8:00 PM",
		ContentRating: "TV-PG",
		Genre:         PipeList{"Animation", "Comedy"},
		Network:       "FOX",
		Rating:        "9.0",
		RatingCount:   "543",
		Runtime:       "30",
		Status:        "Continuing",
		Added:         "",
		AddedBy:       "",
		FanartPath:    "fanart/original/71663-31.jpg",
		LastUpdated:   "1422643901",
		PostersPath:   "",
		seriesShared: seriesShared{
			ID:         71663,
			Language:   "",
			Name:       "The Simpsons",
			BannerPath: "graphical/71663-g13.jpg",
			Overview:   "Set in Springfield, the average American town, the show focuses on the antics and everyday adventures of the Simpson family; Homer, Marge, Bart, Lisa and Maggie, as well as a virtual cast of thousands. Since the beginning, the series has been a pop culture icon, attracting hundreds of celebrities to guest star. The show has also made name for itself in its fearless satirical take on politics, media and American life in general.",
			FirstAired: "1989-12-17",
			IMDBID:     "tt0096697",
			Zap2itID:   "EP00018693",
			Network:    ""},
	}

	episodeWant := Episode{
		ID: 4350173,
		CombinedEpisodeNumber: "1",
		CombinedSeason:        0,
		DvdChapter:            "",
		DvdDiscID:             "",
		DvdEpisodeNumber:      "",
		DvdSeason:             "",
		Director:              PipeList{"Gabor Csupo"},
		EpImgFlag:             "1",
		EpisodeName:           "Good Night",
		EpisodeNumber:         1,
		FirstAired:            "1987-04-19",
		GuestStars:            "",
		ImdbID:                "",
		Language:              "en",
		Overview:              "Good Night was the first ever Simpsons short to air on The Tracey Ullman Show. The five main family members - Homer, Marge, Bart, Lisa, and Maggie - were first introduced in this short. Homer and Marge attempt to calm their children to sleep, with the opposite results. \n\nMaggie can be heard saying \"good night\". She rarely talks throughout the run of the series.",
		ProductionCode:        "101",
		Rating:                "7.0",
		RatingCount:           "1",
		SeasonNumber:          0,
		Writer:                PipeList{""},
		AbsoluteNumber:        "",
		Filename:              "episodes/71663/4350173.jpg",
		LastUpdated:           "1340731501",
		SeasonID:              19130,
		SeriesID:              71663,
		ThumbAdded:            "",
		ThumbHeight:           "225",
		ThumbWidth:            "300",
	}

	if !reflect.DeepEqual(series, want) {
		t.Errorf("Response does not match.  \n%s", pretty.Compare(want, series))
	}

	if !reflect.DeepEqual(episodes[0], episodeWant) {
		t.Errorf("Episode 0 does not match.  \n%s", pretty.Compare(episodeWant, episodes[0]))
	}
}

func TestEpisodeByID(t *testing.T) {
	client := setup()
	defer teardown()

	handler = newFileHandler("testdata/episodes_4350173_en.xml")
	mux.Handle(fmt.Sprintf("/api/%s/episodes/4350173/en.xml", apiKey), handler)

	episode, err := client.EpisodeByID(4350173, "en")
	if err != nil {
		t.Fatal(err)
	}

	want := &Episode{
		ID: 4350173,
		CombinedEpisodeNumber: "",
		CombinedSeason:        0,
		DvdChapter:            "",
		DvdDiscID:             "",
		DvdEpisodeNumber:      "",
		DvdSeason:             "",
		Director:              PipeList{"Gabor Csupo"},
		EpImgFlag:             "1",
		EpisodeName:           "Good Night",
		EpisodeNumber:         1,
		FirstAired:            "1987-04-19",
		GuestStars:            "",
		ImdbID:                "",
		Language:              "en",
		Overview:              "Good Night was the first ever Simpsons short to air on The Tracey Ullman Show. The five main family members - Homer, Marge, Bart, Lisa, and Maggie - were first introduced in this short. Homer and Marge attempt to calm their children to sleep, with the opposite results. \n\nMaggie can be heard saying \"good night\". She rarely talks throughout the run of the series.",
		ProductionCode:        "101",
		Rating:                "7",
		RatingCount:           "",
		SeasonNumber:          0,
		Writer:                PipeList{""},
		AbsoluteNumber:        "",
		Filename:              "episodes/71663/4350173.jpg",
		LastUpdated:           "1340731501",
		SeasonID:              19130,
		SeriesID:              71663,
		ThumbAdded:            "",
		ThumbHeight:           "225",
		ThumbWidth:            "300",
	}

	if !reflect.DeepEqual(episode, want) {
		t.Errorf("Response does not match.  \n%s", pretty.Compare(want, episode))
	}
}

func TestEpisodeBySeries(t *testing.T) {
	client := setup()

	defaultHandler := newFileHandler("testdata/series_71663_default_1_1_en.xml")
	dvdHandler := newFileHandler("testdata/series_71663_dvd_1_1_en.xml")
	absHandler := newFileHandler("testdata/series_71663_absolute_1_en.xml")

	defer func() {
		teardown()
		defaultHandler.Close()
		dvdHandler.Close()
		absHandler.Close()
	}()

	mux.Handle(fmt.Sprintf("/api/%s/series/71663/default/1/1/en.xml", apiKey), defaultHandler)
	mux.Handle(fmt.Sprintf("/api/%s/series/71663/dvd/1/1/en.xml", apiKey), dvdHandler)
	mux.Handle(fmt.Sprintf("/api/%s/series/71663/absolute/1/en.xml", apiKey), absHandler)

	for order, ep := range map[string]string{"default": "1/1", "dvd": "1/1", "absolute": "1"} {
		episode, err := client.episodeBySeries(71663, ep, "en", order)
		if err != nil {
			t.Fatal(err)
		}

		want := &Episode{
			ID: 55452,
			CombinedEpisodeNumber: "",
			CombinedSeason:        0,
			DvdChapter:            "",
			DvdDiscID:             "",
			DvdEpisodeNumber:      "1.0",
			DvdSeason:             "1",
			Director:              PipeList{"David Silverman"},
			EpImgFlag:             "1",
			EpisodeName:           "Simpsons Roasting on an Open Fire",
			EpisodeNumber:         1,
			FirstAired:            "1989-12-17",
			GuestStars:            "Christopher Collins",
			ImdbID:                "",
			Language:              "en",
			Overview:              "When his Christmas bonus is cancelled, Homer becomes a department-store Santa--and then bets his meager earnings at the track. When all seems lost, Homer and Bart save Christmas by adopting the losing greyhound, Santa's Little Helper.",
			ProductionCode:        "7G08",
			Rating:                "7.2",
			RatingCount:           "",
			SeasonNumber:          1,
			Writer:                PipeList{"Mimi Pond"},
			AbsoluteNumber:        "1",
			Filename:              "episodes/71663/55452.jpg",
			LastUpdated:           "1306809485",
			SeasonID:              2727,
			SeriesID:              71663,
			ThumbAdded:            "",
			ThumbHeight:           "300",
			ThumbWidth:            "400",
		}

		if !reflect.DeepEqual(episode, want) {
			t.Errorf("episodeBySeries repsonse does not match for order '%s' \n%s", order, pretty.Compare(want, episode))
		}
	}
}
