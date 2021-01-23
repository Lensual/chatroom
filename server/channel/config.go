package channel

type Config struct {
	Name       string `json:"name"`
	SampleRate int    `json:"sample_rate"`
	SampleFmt  string `json:"sample_fmt"`
	Capacity   int    `json:"capacity"`
}
