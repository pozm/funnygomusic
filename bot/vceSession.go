package bot

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/voice"
	"github.com/diamondburned/arikawa/v3/voice/udp"
	"github.com/diamondburned/arikawa/v3/voice/voicegateway"
	"github.com/pkg/errors"
	"log"
	"time"
)

type VoiceSessionUser struct {
	Name     string         `json:"name"`
	ID       discord.UserID `json:"id"`
	Muted    bool           `json:"muted"`
	Deafen   bool           `json:"deafen"`
	Speaking bool           `json:"speaking"`
}

func NewVoiceSessionUser(vste *discord.VoiceState, b *Botter) VoiceSessionUser {
	uname := b.MyUsername
	if vste.UserID != b.MyId {
		u, err := b.BState.User(vste.UserID)
		if err != nil {
			uname = "unknown"
		} else {
			uname = u.DisplayOrUsername()
		}
	}
	return VoiceSessionUser{
		Name:     uname,
		ID:       vste.UserID,
		Muted:    vste.Mute,
		Deafen:   vste.Deaf,
		Speaking: false,
	}
}

type VoiceSessionHndlr struct {
	vs        *voice.Session
	b         *Botter
	GuildId   discord.GuildID    `json:"guild_id"`
	ChannelID discord.ChannelID  `json:"channel_id"`
	Users     []VoiceSessionUser `json:"users"`
}

func (v *VoiceSessionHndlr) AttachVoiceSession(vs *voice.Session) {
	v.vs = vs
	//vs.AddHandler(func(spk *voicegateway.SpeakingEvent) {
	//	log.Println("spk evt", spk)
	//	for k, user := range v.Users {
	//		if user.ID == spk.UserID {
	//			log.Println("spk = ", spk.Speaking, voicegateway.Microphone)
	//			v.Users[k].Speaking = spk.Speaking == voicegateway.Microphone
	//		}
	//	}
	//})
}
func (v *VoiceSessionHndlr) JoinedChannel(chn *discord.ChannelID, gld *discord.GuildID) {
	v.GuildId = *gld
	v.ChannelID = *chn
}

func (v *VoiceSessionHndlr) UpdateUsers() {
	v.Users = nil
	vstates, err := v.b.BState.VoiceStates(v.GuildId)
	if err != nil {
		log.Println("unable to getg voice states: ", err)
		return
	}
	for _, state := range vstates {
		if state.ChannelID != state.ChannelID {
			continue
		}
		v.Users = append(v.Users, NewVoiceSessionUser(&state, v.b))
	}
}

func JoinUsersVc(b *Botter, gld discord.GuildID, uid discord.UserID) error {
	vs, err := voice.NewSession(b.BState)
	if err != nil {
		return errors.Wrap(err, "cannot make new voice session")
	}
	b.V.AttachVoiceSession(vs)
	vs.SetUDPDialer(udp.DialFuncWithFrequency(
		FrameDuration*time.Millisecond, // correspond to -frame_duration
		TimeIncrement,
	))
	uservs, err := b.BState.VoiceState(gld, uid)
	b.V.GuildId = gld
	b.V.ChannelID = uservs.ChannelID

	if err != nil {
		return errors.Wrap(err, "cannot get voice state")
	}
	vs.JoinChannel(b.Ctx, uservs.ChannelID, false, false)
	go b.V.UpdateUsers()
	return nil
}

func (v *VoiceSessionHndlr) Open() bool {
	return v.vs != nil
}
func (v *VoiceSessionHndlr) Leave() {
	v.vs.Leave(v.b.Ctx)
	v.ChannelID = discord.ChannelID(0)
	v.GuildId = discord.GuildID(0)
	v.Users = nil
	v.vs = nil
}
func (v *VoiceSessionHndlr) HasUser(uid discord.UserID) bool {
	for _, user := range v.Users {
		if user.ID == uid {
			return true
		}
	}
	return false
}
func (v *VoiceSessionHndlr) DeleteUser(uid discord.UserID) {
	for k, user := range v.Users {
		if user.ID == uid {
			v.Users = append(v.Users[:k], v.Users[k+1:]...)
			return
		}
	}
}
func (v *VoiceSessionHndlr) GetUser(uid discord.UserID) *VoiceSessionUser {
	for _, user := range v.Users {
		if user.ID == uid {
			return &user
		}
	}
	return nil
}
func (v *VoiceSessionHndlr) UpdateUser(uid discord.UserID, user *VoiceSessionUser) {
	for k, u := range v.Users {
		if u.ID == uid {
			v.Users[k] = *user
			return
		}
	}
}
func (v *VoiceSessionHndlr) AddUser(user *VoiceSessionUser) {
	v.Users = append(v.Users, *user)
}
func (v *VoiceSessionHndlr) Speaking(isSpeaking bool) {
	if !v.Open() {
		return
	}
	if isSpeaking {
		v.vs.Speaking(v.b.Ctx, voicegateway.Microphone)
	} else {
		v.vs.Speaking(v.b.Ctx, voicegateway.NotSpeaking)
	}
}

func (v *VoiceSessionHndlr) GetSession() *voice.Session {
	return v.vs
}
