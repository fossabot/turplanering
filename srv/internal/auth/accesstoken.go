package auth

// https://www.lantmateriet.se/sv/Kartor-och-geografisk-information/oppna-data/

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/EmilGedda/turplanering/srv/internal/env"
	"github.com/EmilGedda/turplanering/srv/internal/errc"
)

type Token struct {
	AccessToken string        `json:"access_token"`
	ExpiresAt   time.Time     `json:"expires_at"`
	ExpiresIn   time.Duration `json:"expires_in"`
}

type TokenService interface {
	RevokeToken(token *Token) error
	RefreshToken(token *Token) (*Token, error)
	GetToken() (*Token, error)
}

type NetClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewNetClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 10,
	}
}

type Lantmateriet struct {
	client      NetClient
	url         string
	consumerID  string
	consumerKey string
}

type LantmaterietOption = func(l *Lantmateriet)

func WithClient(client NetClient) LantmaterietOption {
	return func(l *Lantmateriet) { l.client = client }
}

func WithURL(url string) LantmaterietOption {
	return func(l *Lantmateriet) { l.url = url }
}

func WithConsumerID(id string) LantmaterietOption {
	return func(l *Lantmateriet) { l.consumerID = id }
}

func WithConsumerKey(key string) LantmaterietOption {
	return func(l *Lantmateriet) { l.consumerKey = key }
}

const LantmaterietURL = "https://api.lantmateriet.se"

func NewLantmateriet(opts ...LantmaterietOption) (*Lantmateriet, error) {
	// Default construction
	l := &Lantmateriet{
		NewNetClient(),
		LantmaterietURL,
		env.Vars().ConsumerID,
		env.Vars().ConsumerKey,
	}

	for _, opt := range opts {
		opt(l)
	}

	needs := []string{}

	if l.consumerID == "" {
		needs = append(needs, "ConsumerID")
	}
	if l.consumerKey == "" {
		needs = append(needs, "ConsumerKey")
	}

	if len(needs) > 0 {
		return nil, &errc.NotInitializedError{Module: "Lantmäteriet", Needs: needs}
	}

	return l, nil
}

func (l *Lantmateriet) RevokeToken(token *Token) error {
	if token == nil {
		return errors.New("nil given")
	}

	msg := strings.NewReader("token=" + token.AccessToken)
	request, err := http.NewRequest("POST", l.url+"/revoke", msg)
	if err != nil {
		return errc.Wrap(err, "creating request failed")
	}

	request.SetBasicAuth(l.consumerID, l.consumerKey)
	response, err := l.client.Do(request)
	if err != nil {
		return errc.Wrap(err, "revoke token request failed")
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errc.Wrap(err, "read body failed")
	}

	body := string(data)
	if body != "" {
		return errors.New("expected empty body but got: " + body)
	}

	return nil
}

func (l *Lantmateriet) RefreshToken(token *Token) (*Token, error) {
	err := l.RevokeToken(token)
	if err != nil {
		return nil, errc.Wrap(err, "revoke token")
	}
	return l.GetToken()
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type TokenErrResponse struct {
	ErrorDescription *string `json:"error_description"`
	ErrorType        *string `json:"error"`
}

func NewTokenErrResponse(errDesc, errType string) *TokenErrResponse {
	return &TokenErrResponse{
		ErrorDescription: &errDesc,
		ErrorType:        &errType,
	}
}

func (t *TokenErrResponse) Error() string {
	return "Lantmäteriet API returned error: " + *t.ErrorDescription
}

func (l *Lantmateriet) GetToken() (*Token, error) {
	request, err := http.NewRequest("POST", l.url+"/token", strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return nil, errc.Wrap(err, "creating request failed")
	}

	request.SetBasicAuth(l.consumerID, l.consumerKey)
	response, err := l.client.Do(request)
	if err != nil {
		return nil, errc.Wrap(err, "get token request failed")
	}
	defer response.Body.Close()

	jsonData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errc.Wrap(err, "read body failed")
	}

	errResponse := TokenErrResponse{}
	_ = json.Unmarshal(jsonData, &errResponse)
	if errResponse.ErrorType != nil || errResponse.ErrorDescription != nil {
		return nil, &errResponse
	}

	apiToken := &tokenResponse{}
	err = json.Unmarshal(jsonData, apiToken)
	if err != nil {
		err = &errc.UnmarshalError{Err: err, Json: jsonData, Obj: apiToken}
		return nil, errc.Wrap(err, "failed unmarshal on get token response")
	}

	if apiToken.AccessToken == "" || apiToken.ExpiresIn == 0 {
		return nil, errors.New("json received did not contain access_token and expires_in")
	}

	duration := time.Duration(apiToken.ExpiresIn) * time.Second
	return &Token{
		AccessToken: apiToken.AccessToken,
		ExpiresIn:   duration,
		ExpiresAt:   time.Now().Add(duration),
	}, nil
}

//type cachedToken struct {
//    token Token
//    err error
//}
//
//type tokenProducerUpdate int
//
//const (
//    refresh tokenProducerUpdate = iota,
//    invalidate,
//    stop
//)
//
//type TokenCache struct {
//    atomic.Value // cache the token response
//    token *Token
//    cond *sync.Cond
//    refresh <-chan tokenProducerUpdate
//}
//
//type tokenProducer struct {
//    lock *sync.RWMutex
//    cond *sync.Cond
//    refresh chan<- tokenProducerUpdate
//    fetcher TokenService
//}
//
//func NewTokenCache(fetcher TokenService) *TokenCache {
//    lock := sync.RWMutex{}
//    refresh := make(chan struct{})
//    cond := sync.NewCond(lock.RLocker())
//
//    service := &TokenCache{
//        fetcher.GetToken(),
//        cond,
//        refresh,
//    }
//
//    invalidateToken := func(state error) error {
//        lock.Lock()
//        defer cond.Broadcast()
//        defer lock.Unlock()
//        if service.token.AccessToken == "invalidated" { // check type of Err instead
//            return nil
//        }
//
//        err := fetcher.RevokeToken(service.token) // TODO: timeout
//        service.token.AccessToken = "invalidated"
//        if err == nil {
//            service.token.Err = state
//        } else {
//            service.token.Err = err
//        }
//        return err
//    }
//
//    // Token producer
//    go func() {
//        for {
//            select {
//            case request = <-refresh:
//
//                switch request {
//                case stop:
//                    invalidateToken(errors.New("Token producer stopped"))
//                    return
//                case invalidate:
//                    invalidateToken(errors.New("Token Revoked"))
//                    continue
//                case refresh:
//                    fallthrough
//                }
//
//                fallthrough
//
//            case <-time.After(service.token.ExpiresIn - 5 * time.Second):
//                oldToken := service.token
//                newToken := fetcher.GetToken()  // TODO: timeout / Context
//                lock.Lock()
//                service.token = newToken
//                lock.Unlock()
//                cond.Broadcast()
//                fetcher.RevokeToken(oldToken)
//            }
//        }
//    }()
//
//    return service
//}
//
//func (ts *TokenCache) GetToken() Token {
//    ts.cond.L.Lock()
//    defer ts.cond.L.Unlock()
//    return ts.response
//}
//
//func (ts *TokenCache) updateToken(update tokenProducerUpdate) Token {
//    prev := ts.token
//    ts.refresh <- update
//
//    ts.cond.L.lock()
//    for ts.token.Err == nil && ts.response == prev {
//        ts.cond.Wait()
//    }
//    defer ts.cond.L.Unlock()
//
//    return ts.token
//}
//
//func (ts *TokenCache) RefreshToken() Token {
//    return ts.updateToken(refresh)
//}
//
//func (ts *TokenCache) RevokeToken() (err error) {
//    return ts.updateToken(invalidate).Err
//}
//
//func (ts *TokenCache) Stop() error {
//    err := ts.updateToken(invalidate)
//    if err != nil {
//        ts.updateToken(stop)
//        return err
//    }
//    return ts.updateToken(stop)
//}
