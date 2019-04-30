package ak8s

// Collection contains collections of various K8s resources.
type Collection interface {
	GetKind() string
	GetNames() []string
	Len() int
	Get(string) Resource
	Search(...string) Collection
}

// Resource Def here.
type Resource interface {
	GetName() string
	GetUID() string
}
