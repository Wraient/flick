package internal

type Directory struct {
	Path string
	Name string
	Parent string
	Files []Files
	Id   string
}

type Files struct {
	Id   string
	Name string
	Dir  bool
	Parent string
	Size int64
}

type EpisodeEntry struct {
	Name    string
	ID      string
	Season  int
	Parent  string
	Episode int
}

type Show struct {
	Id           string
	Name         string
	EpisodesList []EpisodeEntry
}

type User struct {
	Watching   EpisodeEntry
	Player     Player
	Resume     bool
}

type Player struct {
	SocketPath string
	PlaybackTime int
	Started      bool
	Duration int
	Speed float64
}
