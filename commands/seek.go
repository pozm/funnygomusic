package commands

import (
	"funnygomusic/bot"
	"funnygomusic/utils"
	"github.com/diamondburned/arikawa/v3/gateway"
	"log"
	"strconv"
)

func SeekCommand(c *gateway.MessageCreateEvent, b *bot.Botter, args []string) {
	timer, err := strconv.ParseFloat(utils.GetIndex(args, 0), 64)
	if err != nil {
		log.Println("failed to parse timer", err)
		return
	}
	asSecs := uint64(timer * 1000)
	b.Queue.Notify <- bot.NewPlaylistMessage(bot.CurrentSeek).SetSeek(int(asSecs))
}
