package remnawave

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	remapi "github.com/Jolymmiles/remnawave-api-go/v2/api"
	"github.com/google/uuid"
	"remnawave-tg-shop-bot/internal/config"
	"remnawave-tg-shop-bot/utils"
)

// shortHash returns 6-char hex of SHA1 over input
func shortHash(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])[:6]
}

// CreateUserForSubscription creates a fresh user for a new subscription to ensure unique credentials/URL per subscription
func (r *Client) CreateUserForSubscription(ctx context.Context, customerId int64, telegramId int64, trafficLimit int, days int, seq int) (*remapi.User, error) {
	// Build base and add short hash to avoid username collisions: {customerId}_{telegramId}_{seq}_{hash}
	base := fmt.Sprintf("%d_%d_%d", customerId, telegramId, seq)
	h := shortHash(fmt.Sprintf("%s_%d", base, time.Now().UnixNano()))
	username := fmt.Sprintf("%s_%s", base, h)
	expireAt := time.Now().UTC().AddDate(0, 0, days)

	resp, err := r.client.InternalSquad().GetInternalSquads(ctx)
	if err != nil {
		return nil, err
	}

	squads := resp.(*remapi.InternalSquadsResponse).GetResponse()
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

	userCreate, err := r.client.Users().CreateUser(ctx, &createUserRequestDto)
	if err != nil {
		return nil, err
	}
	slog.Info("created subscription user", "telegramId", utils.MaskHalf(strconv.FormatInt(telegramId, 10)), "username", utils.MaskHalf(username), "days", days, "seq", seq)
	return &userCreate.(*remapi.UserResponse).Response, nil
}
