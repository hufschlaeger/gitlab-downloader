package domain

type Project struct {
	ID   int
	Name string
}

type Release struct {
	ProjectID int
	Tag       string
	Assets    Assets
}

type Assets struct {
	Links   []Link
	Sources []Source
}

type Link struct {
	Name string
	URL  string
}

type Source struct {
	Format string
	URL    string
}

type DownloadRequest struct {
	ProjectName string
	ReleaseTag  string
	OutputPath  string
	ExtIndex    int
	Token       string
}
