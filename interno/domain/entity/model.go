package entity

type Model struct {
	Nome      string
	MaxTokens int
}

func NewModel(nome string, maxTokens int) *Model {
	return &Model{
		Nome:      nome,
		MaxTokens: maxTokens,
	}
}

func (m *Model) GetMaxTokens() int {
	return m.MaxTokens
}

func (m *Model) GetModelNome() string {
	return m.Nome
}
