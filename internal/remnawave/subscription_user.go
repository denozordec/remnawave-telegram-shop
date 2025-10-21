package remnawave

import (
	"context"
	"errors"
	"fmt"
	remapi "github.com/Jolymmiles/remnawave-api-go/v2/api"
	"github.com/google/uuid"
	"log/slog"
	"remnawave-tg-shop-bot/internal/config"
	"remnawave-tg-shop-bot/utils"
	"strconv"
	"strings"
	"time"
)

// CreateUserForSubscription creates a fresh user for a new subscription to ensure unique credentials/URL per subscription
func (r *Client) CreateUserForSubscription(ctx context.Context, customerId int64, telegramId int64, trafficLimit int, days int, seq int) (*remapi.UserDto, error) {
	// Build unique username per subscription: {customerId}_{telegramId}_{seq}
	username := fmt.Sprintf("%d_%d_%d", customerId, telegramId, seq)
	expireAt := time.Now().UTC().AddDate(0, 0, days)

	resp, err := r.client.InternalSquadControllerGetInternalSquads(ctx)
	if err != nil {
		return nil, err
	}

	squads := resp.(*remapi.GetInternalSquadsResponseDto).GetResponse()
	squadId := make([]uuid.UUID, 0, len(config.SquadUUIDs()))
	for _, squad := range squads.GetInternalSquads() {
		if config.SquadUUIDs() != nil && len(config.SquadUUIDs()) > 0 {
			if _, isExist := config.SquadUUIDs()[squad.UUID]; !isExist {
				continue
			} else {
				squadId = append(squadId, squad.UUID)
			}
		} else {
			squadId = append(squadId, squad.UUID)
		}
	}

	createUserRequestDto := remapi.CreateUserRequestDto{
		Username:             username,
		ActiveInternalSquads: squadId,
		Status:               remapi.NewOptCreateUserRequestDtoStatus(remapi.CreateUserRequestDtoStatusACTIVE),
		TelegramId:           remapi.NewOptNilInt(int(telegramId)),
		ExpireAt:             expireAt,
		TrafficLimitStrategy: remapi.NewOptCreateUserRequestDtoTrafficLimitStrategy(remapi.CreateUserRequestDtoTrafficLimitStrategyMONTH),
		TrafficLimitBytes:    remapi.NewOptInt(trafficLimit),
	}
	if config.RemnawaveTag() != "" {
		createUserRequestDto.Tag = remapi.NewOptNilString(config.RemnawaveTag())
	}

	if ctx.Value("username") != nil {
		createUserRequestDto.Description = remapi.NewOptString(ctx.Value("username").(string))
	}

	userCreate, err := r.client.UsersControllerCreateUser(ctx, &createUserRequestDto)
	if err != nil {
		return nil, err
	}
	slog.Info("created subscription user", "telegramId", utils.MaskHalf(strconv.FormatInt(telegramId, 10)), "username", utils.MaskHalf(username), "days", days, "seq", seq)
	return &userCreate.(*remapi.UserResponseDto).Response, nil
}
