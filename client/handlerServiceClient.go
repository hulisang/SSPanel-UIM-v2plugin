package client

import (
	"context"
	"fmt"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/model"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/utility"
	"google.golang.org/grpc"
	"strings"
	"v2ray.com/core"
	"v2ray.com/core/app/proxyman"
	"v2ray.com/core/app/proxyman/command"
	"v2ray.com/core/common/net"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/common/uuid"
	"v2ray.com/core/proxy/mtproto"
	"v2ray.com/core/proxy/shadowsocks"
	"v2ray.com/core/proxy/vmess"
	"v2ray.com/core/proxy/vmess/inbound"
	"v2ray.com/core/proxy/vmess/outbound"
	"v2ray.com/core/transport/internet"
	"v2ray.com/core/transport/internet/headers/noop"
	"v2ray.com/core/transport/internet/headers/srtp"
	"v2ray.com/core/transport/internet/headers/tls"
	"v2ray.com/core/transport/internet/headers/utp"
	"v2ray.com/core/transport/internet/headers/wechat"
	"v2ray.com/core/transport/internet/headers/wireguard"
	"v2ray.com/core/transport/internet/kcp"
	"v2ray.com/core/transport/internet/websocket"
)

var KcpHeadMap = map[string]*serial.TypedMessage{
	"wechat-video": serial.ToTypedMessage(&wechat.VideoConfig{}),
	"srtp":         serial.ToTypedMessage(&srtp.Config{}),
	"utp":          serial.ToTypedMessage(&utp.Config{}),
	"wireguard":    serial.ToTypedMessage(&wireguard.WireguardConfig{}),
	"dtls":         serial.ToTypedMessage(&tls.PacketConfig{}),
	"noop":         serial.ToTypedMessage(&noop.Config{}),
}
var CipherTypeMap = map[string]shadowsocks.CipherType{
	"aes-256-cfb":            shadowsocks.CipherType_AES_256_CFB,
	"aes-128-cfb":            shadowsocks.CipherType_AES_128_CFB,
	"aes-128-gcm":            shadowsocks.CipherType_AES_128_GCM,
	"aes-256-gcm":            shadowsocks.CipherType_AES_256_GCM,
	"chacha20":               shadowsocks.CipherType_CHACHA20,
	"chacha20-ietf":          shadowsocks.CipherType_CHACHA20_IETF,
	"chacha20-ploy1305":      shadowsocks.CipherType_CHACHA20_POLY1305,
	"chacha20-ietf-poly1305": shadowsocks.CipherType_CHACHA20_POLY1305,
}

type HandlerServiceClient struct {
	command.HandlerServiceClient
	InboundTag string
}

func NewHandlerServiceClient(client *grpc.ClientConn, inboundTag string) *HandlerServiceClient {
	return &HandlerServiceClient{
		HandlerServiceClient: command.NewHandlerServiceClient(client),
		InboundTag:           inboundTag,
	}
}

// user
func (h *HandlerServiceClient) DelUser(email string) error {
	req := &command.AlterInboundRequest{
		Tag:       h.InboundTag,
		Operation: serial.ToTypedMessage(&command.RemoveUserOperation{Email: email}),
	}
	return h.AlterInbound(req)
}

func (h *HandlerServiceClient) AddUser(user model.UserModel) error {
	req := &command.AlterInboundRequest{
		Tag:       h.InboundTag,
		Operation: serial.ToTypedMessage(&command.AddUserOperation{User: h.ConvertVmessUser(user)}),
	}
	return h.AlterInbound(req)
}

func (h *HandlerServiceClient) AlterInbound(req *command.AlterInboundRequest) error {
	_, err := h.HandlerServiceClient.AlterInbound(context.Background(), req)
	return err
}

//streaming
func GetKcpStreamConfig(headkey string) *internet.StreamConfig {
	var streamsetting internet.StreamConfig
	head, _ := KcpHeadMap["noop"]
	if _, ok := KcpHeadMap[headkey]; ok {
		head, _ = KcpHeadMap[headkey]
	}
	streamsetting = internet.StreamConfig{
		ProtocolName: "mkcp",
		TransportSettings: []*internet.TransportConfig{
			&internet.TransportConfig{
				ProtocolName: "mkcp",
				Settings: serial.ToTypedMessage(
					&kcp.Config{
						HeaderConfig: head,
					}),
			},
		},
	}
	return &streamsetting
}

func GetWebSocketStreamConfig(path string, host string) *internet.StreamConfig {
	var streamsetting internet.StreamConfig
	streamsetting = internet.StreamConfig{
		ProtocolName: "websocket",
		TransportSettings: []*internet.TransportConfig{
			&internet.TransportConfig{
				ProtocolName: "websocket",
				Settings: serial.ToTypedMessage(&websocket.Config{
					Path: path,
					Header: []*websocket.Header{
						&websocket.Header{
							Key:   "Hosts",
							Value: host,
						},
					},
				}),
			},
		},
	}
	return &streamsetting
}

// different type inbounds
func (h *HandlerServiceClient) AddVmessInbound(port uint16, address string, streamsetting *internet.StreamConfig) error {
	var addinboundrequest command.AddInboundRequest
	addinboundrequest = command.AddInboundRequest{
		Inbound: &core.InboundHandlerConfig{
			Tag: h.InboundTag,
			ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
				PortRange:      net.SinglePortRange(net.Port(port)),
				Listen:         net.NewIPOrDomain(net.ParseAddress(address)),
				StreamSettings: streamsetting,
			}),
			ProxySettings: serial.ToTypedMessage(&inbound.Config{
				User: []*protocol.User{
					{
						Level: 0,
						Email: "rico93@xxx.com",
						Account: serial.ToTypedMessage(&vmess.Account{
							Id:      protocol.NewID(uuid.New()).String(),
							AlterId: 16,
						}),
					},
				},
			}),
		},
	}
	return h.AddInbound(&addinboundrequest)
}

func (h *HandlerServiceClient) AddVmessOutbound(tag string, port uint16, address string, streamsetting *internet.StreamConfig, user *protocol.User) error {
	var addoutboundrequest command.AddOutboundRequest
	addoutboundrequest = command.AddOutboundRequest{
		Outbound: &core.OutboundHandlerConfig{
			Tag: tag,
			SenderSettings: serial.ToTypedMessage(&proxyman.SenderConfig{
				StreamSettings: streamsetting,
			}),
			ProxySettings: serial.ToTypedMessage(&outbound.Config{
				Receiver: []*protocol.ServerEndpoint{
					{
						Address: net.NewIPOrDomain(net.ParseAddress(address)),
						Port:    uint32(port),
						User: []*protocol.User{
							user,
						},
					},
				},
			}),
		},
	}
	return h.AddOutbound(&addoutboundrequest)
}

func (h *HandlerServiceClient) AddSSOutbound(user model.UserModel, dist *model.DisNodeInfo) error {
	var addoutboundrequest command.AddOutboundRequest
	addoutboundrequest = command.AddOutboundRequest{
		Outbound: &core.OutboundHandlerConfig{
			Tag: dist.Server_raw + fmt.Sprintf("%d", user.UserID),
			ProxySettings: serial.ToTypedMessage(&shadowsocks.ClientConfig{
				Server: []*protocol.ServerEndpoint{
					{
						Address: net.NewIPOrDomain(net.ParseAddress(dist.Server["server_address"].(string))),
						Port:    uint32(dist.Port),
						User: []*protocol.User{
							h.ConverSSUser(user),
						},
					},
				},
			}),
		},
	}
	return h.AddOutbound(&addoutboundrequest)
}
func (h *HandlerServiceClient) AddMTInbound(port uint16, address string, streamsetting *internet.StreamConfig) error {
	var addinboundrequest command.AddInboundRequest
	addinboundrequest = command.AddInboundRequest{
		Inbound: &core.InboundHandlerConfig{
			Tag: "tg-in",
			ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
				PortRange: net.SinglePortRange(net.Port(port)),
				Listen:    net.NewIPOrDomain(net.ParseAddress(address)),
			}),
			ProxySettings: serial.ToTypedMessage(&mtproto.ServerConfig{
				User: []*protocol.User{
					{
						Level: 0,
						Email: "rico93@xxx.com",
						Account: serial.ToTypedMessage(&mtproto.Account{
							Secret: utility.MD5(utility.GetRandomString(16)),
						}),
					},
				},
			}),
		},
	}
	return h.AddInbound(&addinboundrequest)
}
func (h *HandlerServiceClient) AddSSInbound(user model.UserModel) error {
	var addinboundrequest command.AddInboundRequest
	addinboundrequest = command.AddInboundRequest{
		Inbound: &core.InboundHandlerConfig{
			Tag: user.PrefixedId,
			ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
				PortRange: net.SinglePortRange(net.Port(user.Port)),
				Listen:    net.NewIPOrDomain(net.ParseAddress("0.0.0.0")),
			}),
			ProxySettings: serial.ToTypedMessage(&shadowsocks.ServerConfig{
				User:    h.ConverSSUser(user),
				Network: []net.Network{net.Network_TCP, net.Network_UDP},
			}),
		},
	}
	return h.AddInbound(&addinboundrequest)
}
func (h *HandlerServiceClient) AddInbound(req *command.AddInboundRequest) error {
	_, err := h.HandlerServiceClient.AddInbound(context.Background(), req)
	return err
}
func (h *HandlerServiceClient) AddOutbound(req *command.AddOutboundRequest) error {
	_, err := h.HandlerServiceClient.AddOutbound(context.Background(), req)
	return err
}
func (h *HandlerServiceClient) RemoveInbound(tag string) error {
	req := command.RemoveInboundRequest{
		Tag: tag,
	}
	_, err := h.HandlerServiceClient.RemoveInbound(context.Background(), &req)
	return err
}
func (h *HandlerServiceClient) RemoveOutbound(tag string) error {
	req := command.RemoveOutboundRequest{
		Tag: tag,
	}
	_, err := h.HandlerServiceClient.RemoveOutbound(context.Background(), &req)
	return err
}

func (h *HandlerServiceClient) ConvertVmessUser(userModel model.UserModel) *protocol.User {
	return &protocol.User{
		Level: 0,
		Email: userModel.Email,
		Account: serial.ToTypedMessage(&vmess.Account{
			Id:      userModel.Uuid,
			AlterId: userModel.AlterId,
			SecuritySettings: &protocol.SecurityConfig{
				Type: protocol.SecurityType(protocol.SecurityType_value[strings.ToUpper("AUTO")]),
			},
		}),
	}
}
func (h *HandlerServiceClient) ConverSSUser(userModel model.UserModel) *protocol.User {
	return &protocol.User{
		Level: 0,
		Email: userModel.Email,
		Account: serial.ToTypedMessage(&shadowsocks.Account{
			Password:   userModel.Passwd,
			CipherType: CipherTypeMap[strings.ToLower(userModel.Method)],
			Ota:        shadowsocks.Account_Auto,
		}),
	}
}

func (h *HandlerServiceClient) ConverMTUser(userModel model.UserModel) *protocol.User {
	return &protocol.User{
		Level: 0,
		Email: userModel.Email,
		Account: serial.ToTypedMessage(&mtproto.Account{
			Secret: utility.MD5(userModel.Uuid),
		}),
	}
}
