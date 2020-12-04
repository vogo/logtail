package logtail

import "github.com/vogo/logger"

type Filter struct {
	Matcher Matcher
	Alerter Alerter
}

func (f *Filter) Write(bytes []byte) (int, error) {
	matches := f.Matcher.Match(bytes)
	if len(matches) > 0 {
		for _, m := range matches {
			if err := f.Alerter.Alert(m); err != nil {
				logger.Errorf("send alert error: %+v", err)
			}
		}
	}
	return len(bytes), nil
}

func NewFilter(matcher Matcher, alerter Alerter) *Filter {
	return &Filter{
		Matcher: matcher,
		Alerter: alerter,
	}
}
