package ak8s

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ActionOption defines an Action.
type ActionOption int

// ActionOption Constants.
const (
	ListOption   ActionOption = 0
	GetOption    ActionOption = 1
	DeleteOption ActionOption = 2
)

// ActionsMap maps ActionOptions to K8sActions.
type ActionsMap map[ActionOption]K8sAction

// K8sAction is any k8s option needed to perform actions or tasks.
type K8sAction interface {
	GetType() ActionOption
}

// ListAction contains Options for performing Listing type Actions.
type ListAction struct {
	ListOptions v1.ListOptions
}

// GetType implements K8sAction.
func (a *ListAction) GetType() ActionOption {
	return ListOption
}

// Get returns the v1 Option.
func (a *ListAction) Get() v1.ListOptions {
	return a.ListOptions
}

// GetAction contains Options for performing Listing type Actions.
type GetAction struct {
	GetOptions v1.GetOptions
}

// GetType implements K8sAction.
func (a *GetAction) GetType() ActionOption {
	return ListOption
}

// Get returns the v1 Option.
func (a *GetAction) Get() v1.GetOptions {
	return a.GetOptions
}

// DeleteAction contains Options for performing Delete type Actions.
type DeleteAction struct {
	DeleteOptions v1.DeleteOptions
}

// GetType implements K8sAction.
func (a *DeleteAction) GetType() ActionOption {
	return DeleteOption
}

// Get returns the v1 Option.
func (a *DeleteAction) Get() v1.DeleteOptions {
	return a.DeleteOptions
}

func makeActionMap() ActionsMap {
	actionsMap := make(map[ActionOption]K8sAction, 2)
	actionsMap[ListOption] = &ListAction{
		ListOptions: v1.ListOptions{},
	}
	actionsMap[GetOption] = &GetAction{
		GetOptions: v1.GetOptions{},
	}
	actionsMap[DeleteOption] = &DeleteAction{
		DeleteOptions: v1.DeleteOptions{},
	}
	return actionsMap
}
