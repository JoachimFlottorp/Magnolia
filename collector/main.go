package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"os"

	"github.com/JoachimFlottorp/yeahapi/internal/config"
	"github.com/JoachimFlottorp/yeahapi/protobuf/collector"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type server struct {
	collector.UnsafeChattersServer
}

const (
	URL = "https://gql.twitch.tv/gql"

	QueryGetChatters = `
		query GetChatViewers($login: String!) {
			channel(name: $login) {
				chatters {
					count,
					broadcasters {
						login
					},
					moderators {
						login
					},
					staff {
						login
					},
					viewers {
						login
					},
					vips {
						login
					}
				}
			}
		}`
)

type GqlOperation struct {
	Query string `json:"query"`
	Variables map[string]string `json:"variables"`
}
type GqlChannel struct {
	Channel struct {
		Chatters struct {
			Count      	int `json:"count"`
			Broadcasters []struct {
				Login string `json:"login"`
			} `json:"broadcasters"`
			Moderators []struct {
				Login string `json:"login"`
			} `json:"moderators"`
			Staff   []struct{
				Login string `json:"login"`
			} `json:"staff"`
			Viewers []struct {
				Login string `json:"login"`
			} `json:"viewers"`
			Vips []struct {
				Login string `json:"login"`
			} `json:"vips"`
		} `json:"chatters"`
	} `json:"channel"`
}

type GqlResponse struct {
	Data json.RawMessage `json:"data"`
	Extensions struct {
		DurationMilliseconds int    `json:"durationMilliseconds"`
		RequestID            string `json:"requestID"`
	} `json:"extensions"`
}


func (s *server) GetChatters(ctx context.Context, in *collector.ChatterRequest) (*collector.ChatterResponse, error) {
	login := in.Login
	ClientID := os.Getenv("GQL_CLIENT_ID")

	body := GqlOperation{
		Query: QueryGetChatters,
		Variables: map[string]string{
			"login": login,
		},
	}

	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(body); err != nil {
		zap.S().Errorw("Failed to encode body", "error", err)
		
		return nil, err
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", URL, b)
	if err != nil {
		zap.S().Errorw("Failed to create request", "error", err)
		
		return nil, err
	}
	req.Header.Set("Client-ID", ClientID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zap.S().Errorw("Failed to do request", "error", err)
		
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		zap.S().Errorw("Failed to get chatters", "status", resp.StatusCode, "body", resp.Body)
		
		return nil, err
	}

	response := GqlResponse{}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		zap.S().Errorw("Failed to decode response", "error", err)
		
		return nil, err
	}

	channel := GqlChannel{}

	if err := json.Unmarshal(response.Data, &channel); err != nil {
		zap.S().Errorw("Failed to unmarshal response", "error", err, "response", response.Data)
		
		return nil, err
	}

	amount 		:= channel.Channel.Chatters.Count
	chatters 	:= collector.ChatterList{}

	normalizeUsers := func (users []struct{ Login string `json:"login"` }) []string {
		a := make([]string, len(users))
		
		for i, user := range users {
			a[i] = user.Login
		}

		return a
	}

	chatters.Broadcaster 	= normalizeUsers(channel.Channel.Chatters.Broadcasters)
	chatters.Moderators 	= normalizeUsers(channel.Channel.Chatters.Moderators)
	chatters.Viewers 		= normalizeUsers(channel.Channel.Chatters.Viewers)
	chatters.Vips 			= normalizeUsers(channel.Channel.Chatters.Vips)
	chatters.Staff 			= normalizeUsers(channel.Channel.Chatters.Staff)
	
	return &collector.ChatterResponse{
		Amount: int32(amount),
		Chatters: &chatters,
	}, nil
}

var c = flag.String("config", "config.json", "Path to config file")

func init() {
	flag.Parse()
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
}

func main() {
	cfgFile, err := os.OpenFile(*c, os.O_RDONLY, 0)
	if err != nil {
		zap.S().Fatalw("Config file is not set", "error", err)
	}

	defer func() {
		err := cfgFile.Close()
		zap.S().Warnw("Failed to close config file", "error", err)
	}();

	conf := &config.Config{}
	err = json.NewDecoder(cfgFile).Decode(conf)
	if err != nil {
		zap.S().Fatalw("Failed to decode config file", "error", err)
	}

	port := "localhost:50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		zap.S().Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	collector.RegisterChattersServer(s, &server{})

	zap.S().Infof("Starting gRPC server on %s", port)
	if err := s.Serve(lis); err != nil {
		zap.S().Fatal(err)
	}
}

