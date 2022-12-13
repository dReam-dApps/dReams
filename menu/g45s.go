package menu

const (
	Seals_mint = "dero1qyfq8m3rju62tshju60zuc0ymrajwxqajkdh6pw888ejuv94jlfgjqq58px98"
	ATeam_mint = "dero1qyx9748k9wrt89a6rm0zzlayxgs3ndkmvg6m20shqp8ynh54zf2rgqq8yn9hn"
)

// / Dero Seals
type Seal struct {
	Attributes struct {
		Eyes        string `json:"Eyes"`
		FacialHair  string `json:"Facial Hair"`
		HairAndHats string `json:"Hair And Hats"`
		Shirts      string `json:"Shirts"`
	} `json:"attributes"`
	ID    int     `json:"id"`
	Image string  `json:"image"`
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

// / Dero A-Team
type Agent struct {
	Attributes struct {
		Color  string `json:"Color"`
		IChing string `json:"I-ching"`
	} `json:"attributes"`
	ID    int    `json:"id"`
	Image string `json:"image"`
	Name  string `json:"name"`
}
