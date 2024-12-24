package cah

import (
	"strings"

	"github.com/friendsofgo/errors"
	"github.com/botlabs-gg/yagpdb/v2/bot"
	"github.com/botlabs-gg/yagpdb/v2/commands"
	"github.com/botlabs-gg/yagpdb/v2/lib/cardsagainstdiscord"
	"github.com/botlabs-gg/yagpdb/v2/lib/dcmd"
	"github.com/botlabs-gg/yagpdb/v2/lib/dstate"
	"github.com/sirupsen/logrus"
)

func (p *Plugin) AddCommands() {

	cmdCreate := &commands.YAGCommand{
		Name:        "Create",
		CmdCategory: commands.CategoryFun,
		Aliases:     []string{"c"},
		Description: "Creates a Cards Against Humanity game in this channel, add packs after commands, or * for all packs. (-v for vote mode without a card czar).",
		Arguments: []*dcmd.ArgDef{
			{Name: "packs", Type: dcmd.String, Default: "main", Help: "Packs separated by space, or * for all of them."},
		},
		ArgSwitches: []*dcmd.ArgDef{
			{Name: "v", Help: "Vote mode - players vote instead of having a card czar."},
		},
		RunFunc: func(data *dcmd.Data) (interface{}, error) {
			voteMode := data.Switch("v").Bool()
			pStr := data.Args[0].Str()
			packs := strings.Fields(pStr)

			_, err := p.Manager.CreateGame(data.GuildData.GS.ID, data.GuildData.CS.ID, data.Author.ID, data.Author.Username, voteMode, packs...)
			if err == nil {
				logrus.Info("[cah] Created a new game in ", data.GuildData.CS.ID, ":", data.GuildData.GS.ID)
				return nil, nil
			}

			if cahErr := cardsagainstdiscord.HumanizeError(err); cahErr != "" {
				return cahErr, nil
			}

			return nil, err
		},
	}

	cmdEnd := &commands.YAGCommand{
		Name:        "End",
		CmdCategory: commands.CategoryFun,
		Description: "Ends a Cards Against Humanity game that is ongoing in this channel.",
		RunFunc: func(data *dcmd.Data) (interface{}, error) {
			isAdmin, err := bot.AdminOrPermMS(data.GuildData.GS.ID, data.ChannelID, data.GuildData.MS, 0)
			if err == nil && isAdmin {
				err = p.Manager.RemoveGame(data.ChannelID)
			} else {
				err = p.Manager.TryAdminRemoveGame(data.Author.ID)
			}

			if err != nil {
				if cahErr := cardsagainstdiscord.HumanizeError(err); cahErr != "" {
					return cahErr, nil
				}

				return "", err
			}

			return "Stopped the game", nil
		},
	}

	cmdKick := &commands.YAGCommand{
		Name:         "Kick",
		CmdCategory:  commands.CategoryFun,
		RequiredArgs: 1,
		Arguments: []*dcmd.ArgDef{
			{Name: "user", Type: dcmd.UserID},
		},
		Description: "Kicks a player from the ongoing Cards Against Humanity game in this channel.",
		RunFunc: func(data *dcmd.Data) (interface{}, error) {
			userID := data.Args[0].Int64()
			err := p.Manager.AdminKickUser(data.Author.ID, userID)
			if err != nil {
				if cahErr := cardsagainstdiscord.HumanizeError(err); cahErr != "" {
					return cahErr, nil
				}

				return "", err
			}

			return "User removed", nil
		},
	}

	cmdPacks := &commands.YAGCommand{
		Name:         "Packs",
		CmdCategory:  commands.CategoryFun,
		RequiredArgs: 0,
		Description:  "Lists all available packs.",
		RunFunc: func(data *dcmd.Data) (interface{}, error) {
			resp := "Available packs: \n\n"
			for _, v := range cardsagainstdiscord.Packs {
				resp += "`" + v.Name + "` - " + v.Description + "\n"
			}

			return resp, nil
		},
	}

	cmdMove := &commands.YAGCommand{
		Name:         "Move",
		CmdCategory:  commands.CategoryFun,
		RequiredArgs: 1,
		Arguments: []*dcmd.ArgDef{
			{Name: "channel", Type: dcmd.ChannelOrThread},
		},
		Description: "Move the ongoing Cards Against Humanity game to another channel.",
		RunFunc: func(data *dcmd.Data) (interface{}, error) {
			channelID := data.Args[0].Int64()
			
			if checkNewChannel := p.Manager.FindGameFromChannelOrUser(channelID); checkNewChannel != nil {
				return "There is already a game in the new channel", nil
			}

			err := p.Manager.MoveGameTo(data.ChannelID, channelID)
			if !err {
				return "It seems that there is no game in the current channel", errors.New("Failed to move game")
			}
			
			logrus.Info("[cah] Moved a game from ", data.GuildData.CS.ID, " to ", channelID," :", data.GuildData.GS.ID)
			return "Cah moved", nil
		},
	}


	container, _ := commands.CommandSystem.Root.Sub("cah")
	container.NotFound = commands.CommonContainerNotFoundHandler(container, "")
	container.Description = "Play cards against humanity!"

	container.AddCommand(cmdCreate, cmdCreate.GetTrigger())
	container.AddCommand(cmdEnd, cmdEnd.GetTrigger())
	container.AddCommand(cmdKick, cmdKick.GetTrigger())
	container.AddCommand(cmdPacks, cmdPacks.GetTrigger())
	container.AddCommand(cmdMove, cmdMove.GetTrigger())
	commands.RegisterSlashCommandsContainer(container, true, func(gs *dstate.GuildSet) ([]int64, error) {
		return nil, nil
	})
}
