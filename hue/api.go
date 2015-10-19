package hue

import (
	"fmt"
	"net/http"
)

const (
	ErrCodesLinkButtonNotPressed = 101
)

var (
	ErrLinkButtonNotPressed = fmt.Errorf("link button not pressed")
)

func DiscoverBridges() ([]BridgeInfo, error) {
	var bridges []BridgeInfo
	url := fmt.Sprintf("%v/api/nupnp", meethueURL)
	err := do(http.DefaultClient, "GET", url, nil, &bridges)
	if err != nil {
		return nil, fmt.Errorf("DiscoverBridges: %v", err)
	}
	return bridges, nil
}

func RegisterUser(hostname string) (string, error) {
	var result struct {
		Success struct {
			Username string
		}
		Error struct {
			Description string
			Type        int
		}
	}
	url := fmt.Sprintf("http://%v/api", hostname)
	err := do(http.DefaultClient, "GET", url, nil, &result)

	if err != nil {
		return "", fmt.Errorf("RegisterUser: %v", err)
	} else if result.Error.Type == ErrCodesLinkButtonNotPressed {
		return "", ErrLinkButtonNotPressed
	} else if result.Error.Description != "" {
		return "", fmt.Errorf("RegisterUser: error: type: %v; description: %v", result.Error.Type, result.Error.Description)
	} else {
		return result.Success.Username, nil
	}
}

func (c *Client) GetLights() ([]Light, error) {
	var lightMap map[string]Light
	err := c.get("/lights", &lightMap)
	if err != nil {
		return nil, err
	}
	lights := make([]Light, 0, len(lightMap))
	for id, light := range lightMap {
		light.ID = id
		lights = append(lights, light)
	}
	return lights, nil
}

func (c *Client) GetLight(id string) (Light, error) {
	var light Light
	err := c.get("/lights/"+id, &light)
	return light, err
}

func (c *Client) UpdateLight(light Light) error {
	return c.put("/lights/"+light.ID, light, nil)
}

func (c *Client) UpdateLightState(light Light) error {
	return c.put("/lights/"+light.ID+"/state", light.State, nil)
}

func (c *Client) GetGroups() ([]Group, error) {
	var groupMap map[string]Group
	err := c.get("/groups", &groupMap)
	if err != nil {
		return nil, err
	}
	groups := make([]Group, 0, len(groupMap))
	for id, group := range groupMap {
		group.ID = id
		groups = append(groups, group)
	}
	return groups, nil
}

func (c *Client) GetGroup(id string) (Group, error) {
	var group Group
	err := c.get("/groups/"+id, &group)
	return group, err
}

func (c *Client) UpdateGroup(group Group) error {
	return c.put("/groups/"+group.ID, group, nil)
}

func (c *Client) UpdateGroupState(group Group) error {
	return c.put("/groups/"+group.ID+"/action", group.Action, nil)
}
