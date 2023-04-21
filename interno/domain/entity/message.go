package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	tiktoken_go "github.com/j178/tiktoken-go"
)

type Mensagem struct {
	ID          string
	Indicacao   string //origem da mensagem...
	Conteudo    string
	Tokens      int
	Model       *Model
	DataCriacao time.Time
}

func NovaMensagem(indicacao, conteudo string, Model *Model) (*Mensagem, error) {
	TotalTokens := tiktoken_go.CountTokens(Model.GetModelNome(), conteudo)
	//TotalTokens := tiktoken_go.CountTokens(Model.GetModelNome(), conteudo)
	msg := &Mensagem{
		ID:          uuid.New().String(),
		Indicacao:   indicacao,
		Tokens:      TotalTokens,
		Conteudo:    conteudo,
		Model:       Model,
		DataCriacao: time.Now(),
	}

	if err := msg.Validador(); err != nil {
		return nil, err
	}

	return msg, nil
}

func (m *Mensagem) Validador() error {
	if m.Indicacao != "usuario" && m.Indicacao != "sistema" && m.Indicacao != "assistente" {
		return errors.New("Indicativo Invalido!")
	}

	if m.Conteudo == "" {
		return errors.New("Conteudo vazio")
	}

	if m.DataCriacao.IsZero() {
		return errors.New("Data da criacao vazia")
	}
	return nil
}

func (m *Mensagem) GetQtdTokens() int {
	return m.Tokens
}
