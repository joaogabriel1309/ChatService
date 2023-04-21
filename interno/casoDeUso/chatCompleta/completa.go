package chatcompleta

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/joaogabriel1309/ChatService/interno/domain/entity"
	"github.com/joaogabriel1309/ChatService/interno/domain/gateway"
	"github.com/sashabaranov/go-openai"
)

type ChatCompletaConfiguraEntrada struct {
	Model                 string
	ModelMaxTokens        int
	Temperatura           float32
	TopP                  float32
	N                     int
	Parar                 []string
	MaxTokens             int
	PenalidadeRepeticao   float32
	PenalidadeFrenquencia float32
	MensagemInicial       string
}

// informacoes recebidas do usuario
type ChatCompletaEntrada struct {
	IdChat          string
	IdUsuario       string
	MensagemUsuario string
	configuracao    ChatCompletaConfiguraEntrada
}

type ChatCompletaSaida struct {
	IdChat    string
	IdUsuario string
	Resposta  string
}

type ChatCompletaCasoDeUso struct {
	chatGateway gateway.ChatGateway
	IAClient    *openai.Client
	Stream      chan ChatCompletaEntrada
}

func NovoChatCompletaCasoDeUso(chatGateway gateway.ChatGateway, iAClient *openai.Client, stream chan ChatCompletaEntrada) *ChatCompletaCasoDeUso {
	return &ChatCompletaCasoDeUso{
		chatGateway: chatGateway,
		IAClient:    iAClient,
		Stream:      stream,
	}
}

func (uc *ChatCompletaCasoDeUso) Executa(ctx context.Context, entrada ChatCompletaEntrada) (*ChatCompletaSaida, error) {
	chat, erro := uc.chatGateway.ProcurarChatPorId(ctx, entrada.IdChat)
	if erro != nil {
		if erro.Error() == "chat nao existe" {
			chat, erro = CriarNovoChat(entrada)
			if erro != nil {
				return nil, errors.New("erro na ciracao do chat" + erro.Error())
			}
			//gravar as mensagem em banco...
			erro = uc.chatGateway.CriarChat(ctx, chat)
			if erro != nil {
				return nil, errors.New("erro na gravacao do banco")
			}
			//erro.Error()
		} else {
			return nil, errors.New("erro na procura do chat")
		}
	}
	mensagemUsuario, erro := entity.NovaMensagem("usuario", entrada.MensagemUsuario, chat.Configuracao.Model)
	if erro != nil {
		return nil, errors.New("erro na criacao do mensagem do usuario" + erro.Error())
	}
	erro = chat.AdicionaMensagem(mensagemUsuario)
	if erro != nil {
		return nil, errors.New("erro no envio da mensagem do usuario" + erro.Error())
	}

	//requisicao da mensagem para o chatGPT
	mensagem := []openai.ChatCompletionMessage{}
	for _, msg := range chat.Mensagem {
		mensagem = append(mensagem, openai.ChatCompletionMessage{
			Role:    msg.Indicacao,
			Content: msg.Conteudo,
		})
	}

	resp, erro := uc.IAClient.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:            chat.Configuracao.Model.Nome,
			Messages:         mensagem,
			MaxTokens:        chat.Configuracao.MaxTokens,
			Temperature:      chat.Configuracao.Temperatura,
			TopP:             chat.Configuracao.TopP,
			PresencePenalty:  chat.Configuracao.PenalidadeRepeticao,
			FrequencyPenalty: chat.Configuracao.PenalidadeFrenquencia,
			Stop:             chat.Configuracao.Para,
			Stream:           true,
		},
	)
	if erro != nil {
		return nil, errors.New("erro no recebimento da mensagem" + erro.Error())
	}

	var TodasResportas strings.Builder
	for {
		respostas, erro := resp.Recv()
		if errors.Is(erro, io.EOF) {
			break
		}
		if erro != nil {
			return nil, errors.New("erro no stream respota" + erro.Error())
		}

		TodasResportas.WriteString(respostas.Choices[0].Delta.Content)
		r := ChatCompletaEntrada{
			IdChat:          chat.ID,
			IdUsuario:       chat.IdUsuario,
			MensagemUsuario: TodasResportas.String(),
		}
		uc.Stream <- r
	}

	axuliar, erro := entity.NovaMensagem("auxiliar", TodasResportas.String(), chat.Configuracao.Model)
	if erro != nil {
		return nil, errors.New("erro na criacao da mensagem auxiliar" + erro.Error())
	}
	erro = chat.AdicionaMensagem(axuliar)
	if erro != nil {
		return nil, errors.New("erro ao adicionar a mensagem auxiliar" + erro.Error())
	}
	erro = uc.chatGateway.SalvarChat(ctx, chat)
	if erro != nil {
		return nil, errors.New("erro ao salvar a mensagem auxiliar" + erro.Error())
	}
	return &ChatCompletaSaida{
		IdChat:    chat.ID,
		IdUsuario: chat.IdUsuario,
		Resposta:  TodasResportas.String(),
	}, nil
}

func CriarNovoChat(entrada ChatCompletaEntrada) (*entity.Chat, error) {
	model := entity.NewModel(entrada.configuracao.Model, entrada.configuracao.MaxTokens)
	chatConfig := &entity.ConfiguracaoChat{
		Temperatura:           entrada.configuracao.Temperatura,
		TopP:                  entrada.configuracao.TopP,
		N:                     entrada.configuracao.N,
		Para:                  entrada.configuracao.Parar,
		MaxTokens:             entrada.configuracao.MaxTokens,
		PenalidadeRepeticao:   entrada.configuracao.PenalidadeRepeticao,
		PenalidadeFrenquencia: entrada.configuracao.PenalidadeRepeticao,
		Model:                 model,
	}

	inicializaMensagem, erro := entity.NovaMensagem("sistema", entrada.configuracao.MensagemInicial, model)
	if erro != nil {
		return nil, errors.New("erro na mensagem inicial do sistema" + erro.Error())
	}

	chat, erro := entity.NovoChat(entrada.IdUsuario, inicializaMensagem, chatConfig)
	if erro != nil {
		return nil, errors.New("erro na criacao do chat" + erro.Error())
	}
	return chat, nil
}
