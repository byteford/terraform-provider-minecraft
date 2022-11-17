package minecraft

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/seeruk/minecraft-rcon/rcon"
)

type Client struct {
	client *rcon.Client
}

type Player struct {
	Position struct {
		X float64
		Y float64
		Z float64
	}
}

func New(address string, password string) (*Client, error) {
	addressParts := strings.Split(address, ":")
	host := addressParts[0]
	port, err := strconv.Atoi(addressParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port %s", addressParts[1])
	}

	client, err := rcon.NewClient(host, port, password)
	if err != nil {
		return nil, err
	}

	return &Client{client}, nil
}

// Get a player.
func (c Client) GetPlayer(ctx context.Context, name string) (Player, error) {
	var p Player
	command := fmt.Sprintf("data get entity %s", name)
	res, err := c.client.SendCommand(command)
	if err != nil {
		return p, err
	}
	reg := regexp.MustCompile(`.+ has the following entity data: `)
	res = reg.ReplaceAllString(res, "${1}")
	m := make(map[string]string)
	res = res[1 : len(res)-1]
	arr := strings.Split(res, " ")
	fmt.Printf("%s\n", res)
	for i := 0; i < len(arr); i++ {
		fmt.Println(arr[i])
		if arr[i][len(arr[i])-1] == ':' {
			value := arr[i+1]
			num := i + 1
			if arr[i+1][0] == '[' {
				num = num + 1
				for arr[num][len(arr[num])-2:] != "]," {
					fmt.Printf("next %s\n", arr[num])
					value = value + arr[num]
					num = num + 1
				}
				value = value + arr[num]
			}
			m[arr[i][:len(arr[i])-1]] = value
			i = num
		}
	}
	pos := strings.Split(strings.ReplaceAll(m["Pos"][1:len(m["Pos"])-2], "d", ""), ",")
	p.Position.X, err = strconv.ParseFloat(pos[0], 64)
	if err != nil {
		return p, err
	}
	p.Position.Y, err = strconv.ParseFloat(pos[1], 64)
	if err != nil {
		return p, err
	}
	p.Position.Z, err = strconv.ParseFloat(pos[2], 64)
	if err != nil {
		return p, err
	}
	return p, nil
}
func (c Client) CreatePlayer(ctx context.Context, name string) error {
	command := fmt.Sprintf("player %s spawn", name)
	_, err := c.client.SendCommand(command)
	if err != nil {
		return err
	}

	return nil
}
func (c Client) MovePlayer(ctx context.Context, name string, x, y, z float64) error {
	command := fmt.Sprintf("tp %s %f %f %f", name, x, y, z)
	_, err := c.client.SendCommand(command)
	if err != nil {
		return err
	}

	return nil
}
func (c Client) KickPlayer(ctx context.Context, id string) error {
	command := fmt.Sprintf("player %s kill", id)
	_, err := c.client.SendCommand(command)
	if err != nil {
		return err
	}

	return nil
}
func (c Client) GetBlockMaterial(ctx context.Context, x, y, z int) (string, error) {
	command := fmt.Sprintf("info block %d %d %d grep Material", x, y, z)
	res, err := c.client.SendCommand(command)
	if err != nil {
		return "", err
	}
	sl := strings.Split(res, " ")
	mat := sl[len(sl)-1]

	if strings.Contains(mat, ":") {
		return mat, nil
	}

	return fmt.Sprintf("minecraft:%s", mat), nil
}

// Creates a block.
func (c Client) CreateBlock(ctx context.Context, material string, x, y, z int) error {
	command := fmt.Sprintf("setblock %d %d %d %s replace", x, y, z, material)
	_, err := c.client.SendCommand(command)
	if err != nil {
		return err
	}

	return nil
}
func (c Client) FillBlock(ctx context.Context, material string, sx, sy, sz, ex, ey, ez int) error {
	command := fmt.Sprintf("fill %d %d %d %d %d %d %s hollow", sx, sy, sz, ex, ey, ez, material)
	_, err := c.client.SendCommand(command)
	if err != nil {
		return err
	}

	return nil
}

// Deletes a block.
func (c Client) DeleteBlock(ctx context.Context, x, y, z int) error {
	command := fmt.Sprintf("setblock %d %d %d minecraft:air replace", x, y, z)
	_, err := c.client.SendCommand(command)
	if err != nil {
		return err
	}

	return nil
}

// Creates an entity.
func (c Client) CreateEntity(ctx context.Context, entity string, position string, id string) error {
	command := fmt.Sprintf("summon minecraft:%s %s {CustomName:'{\"text\":\"%s\"}'}", entity, position, id)
	_, err := c.client.SendCommand(command)
	if err != nil {
		return err
	}

	return nil
}

// Deletes an entity.
func (c Client) DeleteEntity(ctx context.Context, entity string, position string, id string) error {
	// Remove the entity.
	command := fmt.Sprintf("kill @e[type=minecraft:%s,nbt={CustomName:'{\"text\":\"%s\"}'}]", entity, id)
	_, err := c.client.SendCommand(command)
	if err != nil {
		return err
	}

	// Remove the entity from inventories.
	command = fmt.Sprintf("clear @a minecraft:%s{display:{Name:'{\"text\":\"%s\"}'}}", entity, id)
	_, err = c.client.SendCommand(command)
	if err != nil {
		return err
	}

	return nil
}
