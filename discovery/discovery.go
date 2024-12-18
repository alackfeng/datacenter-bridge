package discovery

import "fmt"

// Service -
type Service struct {
	Zone    string `json:"Zone" comment:"consul service zone"`
	Id      string `json:"ID" comment:"consul service id"`
	Service string `json:"Service" comment:"consul service name"`
	Host    string `json:"Host" comment:"consul service host"`
	Tag     string `json:"Tag" comment:"consul service tag"`
}

// Key -
func (s Service) Key() string {
	return fmt.Sprintf("%s_%s", s.Zone, s.Service)
}

// String -
func (s Service) String() string {
	return fmt.Sprintf("[%s]%s(%s)", s.Service, s.Host, s.Tag)
}

// Discovery - 服务发现.
type Discovery interface {
}
