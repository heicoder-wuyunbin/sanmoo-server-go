package mpuser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sanmoo-server-go/internal/infrastructure/config"
	"sanmoo-server-go/internal/infrastructure/logger"
	"time"
)

type SubscribeService struct {
	wechatProvider *config.WechatConfigProvider
	tokenCache     *TokenCache
}

type TokenCache struct {
	accessToken string
	expireTime  time.Time
	mu          chan struct{}
}

func NewSubscribeService(wechatProvider *config.WechatConfigProvider) *SubscribeService {
	return &SubscribeService{
		wechatProvider: wechatProvider,
		tokenCache: &TokenCache{
			mu: make(chan struct{}, 1),
		},
	}
}

type WXAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

func (s *SubscribeService) getAccessToken(ctx context.Context) (string, error) {
	s.tokenCache.mu <- struct{}{}
	defer func() { <-s.tokenCache.mu }()

	if s.tokenCache.accessToken != "" && time.Now().Before(s.tokenCache.expireTime) {
		return s.tokenCache.accessToken, nil
	}

	// 从动态配置提供者获取微信配置
	appID := ""
	appSecret := ""
	if s.wechatProvider != nil {
		wechatCfg, err := s.wechatProvider.Get(ctx)
		if err == nil && wechatCfg != nil {
			appID = wechatCfg.AppID
			appSecret = wechatCfg.Secret
		}
	}
	if appID == "" || appSecret == "" {
		return "", fmt.Errorf("微信小程序 AppID/Secret 未配置")
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("request access token failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %w", err)
	}

	var tokenResp WXAccessTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parse response failed: %w", err)
	}

	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("wechat API error: %d - %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	s.tokenCache.accessToken = tokenResp.AccessToken
	s.tokenCache.expireTime = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return tokenResp.AccessToken, nil
}

type SubscribeMessage struct {
	ToUser     string                 `json:"touser"`
	TemplateID string                 `json:"template_id"`
	Page       string                 `json:"page"`
	Data       map[string]TemplateData `json:"data"`
	MiniprogramState string           `json:"miniprogram_state,omitempty"`
	Lang            string            `json:"lang,omitempty"`
}

type TemplateData struct {
	Value string `json:"value"`
}

func (s *SubscribeService) SendSubscribeMessage(ctx context.Context, openID, templateID, page string, data map[string]TemplateData) error {
	token, err := s.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get access token failed: %w", err)
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", token)

	msg := SubscribeMessage{
		ToUser:     openID,
		TemplateID: templateID,
		Page:       page,
		Data:       data,
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("send message failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %w", err)
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse response failed: %w", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("wechat API error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return nil
}

func (s *SubscribeService) SendNewArticleNotification(ctx context.Context, openID, templateID, articleID, articleTitle, author string) error {
	data := map[string]TemplateData{
		"thing1": {Value: articleTitle},
		"thing2": {Value: author},
		"time3":  {Value: time.Now().Format("2006年01月02日 15:04")},
	}
	page := fmt.Sprintf("/pages/article-detail/index?id=%s", articleID)
	return s.SendSubscribeMessage(ctx, openID, templateID, page, data)
}

func (s *SubscribeService) BatchSendNewArticleNotification(ctx context.Context, openIDs []string, templateID, articleID, articleTitle, author string) {
	data := map[string]TemplateData{
		"thing1": {Value: articleTitle},
		"thing2": {Value: author},
		"time3":  {Value: time.Now().Format("2006年01月02日 15:04")},
	}
	page := fmt.Sprintf("/pages/article-detail/index?id=%s", articleID)

	for _, openID := range openIDs {
		go func(id string) {
			err := s.SendSubscribeMessage(ctx, id, templateID, page, data)
			if err != nil {
				logger.Warnf("send subscribe message to %s failed: %v", id, err)
			}
		}(openID)
	}
}