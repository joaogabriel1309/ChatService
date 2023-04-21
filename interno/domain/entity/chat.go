package entity

import (
	"errors"

	"github.com/google/uuid"
)

type ConfiguracaoChat struct {
	Model                 *Model
	Temperatura           float32  //precisao da resposta
	TopP                  float32  //quanto conservador ele vai ser na escolha das palavras
	N                     int      //numero de mensagem
	Para                  []string //string para parar o chat
	MaxTokens             int      //quantos tokens pode ter no maximo
	PenalidadeRepeticao   float32  //penalidade de mensagem repetidas...
	PenalidadeFrenquencia float32
}

func NovoChat(IdUsuario string, inicializaSistemaMensagem *Mensagem, configuracaoChat *ConfiguracaoChat) (*Chat, error) {
	chat := &Chat{
		ID:              uuid.New().String(),
		IdUsuario:       IdUsuario,
		MensagemInicial: inicializaSistemaMensagem,
		Status:          "ativado",
		TokensUsados:    0,
		Configuracao:    configuracaoChat,
	}
	chat.AdicionaMensagem(inicializaSistemaMensagem)

	if erro := chat.Validador(); erro != nil {
		return nil, erro
	}
	return chat, nil
}

func (c *Chat) Validador() error {
	if c.IdUsuario == "" {
		return errors.New("Id do usuario vazio")
	}
	if c.Status != "ativo" && c.Status != "terminado" {
		return errors.New("status invalido")
	}
	if c.Configuracao.Temperatura < 0 || c.Configuracao.Temperatura > 2 {
		return errors.New("temperatura invalida")
	}
	return nil
}

type Chat struct {
	ID               string
	IdUsuario        string
	MensagemInicial  *Mensagem
	Mensagem         []*Mensagem
	MensagemApagadas []*Mensagem
	Status           string
	TokensUsados     int
	Configuracao     *ConfiguracaoChat
}

func (c *Chat) AdicionaMensagem(m *Mensagem) error {
	if c.Status == "terminado" {
		return errors.New("Este chat esta terminado, não são permitidas mais mensagens")
	}

	for {
		if c.Configuracao.Model.GetMaxTokens() >= m.GetQtdTokens()+c.TokensUsados {
			c.Mensagem = append(c.Mensagem, m)
			c.AtualizaTokensUsados()
			break
		}
		c.MensagemApagadas = append(c.MensagemApagadas, c.Mensagem[0])
		c.Mensagem = c.Mensagem[1:]
		c.AtualizaTokensUsados()
	}
	return nil
}

func (c *Chat) PegarMensagem() []*Mensagem {
	return c.Mensagem
}

func (c *Chat) TotalMensagem() int {
	return len(c.Mensagem)
}

func (c *Chat) TerminaChat() {
	c.Status = "terminado"
}

func (c *Chat) AtualizaTokensUsados() {
	c.TokensUsados = 0
	for m := range c.Mensagem {
		c.TokensUsados += c.Mensagem[m].GetQtdTokens()
	}
}
