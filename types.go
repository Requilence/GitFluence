package main

type UserStat struct {
	CodeLines   LinesStat             `bson:",omitempty"`
	DocLines    LinesStat             `bson:",omitempty"`
	TestLines   LinesStat             `bson:",omitempty"`
	Resources   LinesStat             `bson:",omitempty"`
	LinesPerExt map[string]*LinesStat `bson:",omitempty"`
	Email       string
	CommitID    string `bson:",omitempty"`
	CommitDays  int    `bson:",omitempty"`
	Username    string `bson:",omitempty"`
}

type LinesStat struct {
	LastMonth  int `bson:",omitempty"`
	Last3Month int `bson:",omitempty"`
	Last6Month int `bson:",omitempty"`
	LastYear   int `bson:",omitempty"`
	Total      int `bson:",omitempty"`
}

func (l *LinesStat) Append(lines LinesStat) {
	l.LastMonth += lines.LastMonth
	l.Last3Month += lines.Last3Month
	l.Last6Month += lines.Last6Month
	l.LastYear += lines.LastYear
	l.Total += lines.Total
}

type FileStat struct {
	IsDoc      bool
	IsTest     bool
	IsBinary   bool
	TotalLines int
	Users      map[string]*UserStat
}

type RepoStat struct {
	CodeLines LinesStat
	TestLines LinesStat
	DocLines  LinesStat
	Resources LinesStat

	Users map[string]*UserStat
	Files map[string]*FileStat
}

type RepoConfig struct {
	URL string
}
type Repo struct {
	Token string
	Host  string
	Owner string
	Name  string
	Stat  RepoStat
}
