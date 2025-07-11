package elicit

import (
	"encoding/json"
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func Answer(request mcp.ElicitRequest, autoApproveTool bool) (mcp.ElicitResult, error) {
	var tool types.ToolCallConfirm
	if err := json.Unmarshal(request.Meta, &tool); err == nil && tool.Type == types.ToolCallConfirmType && autoApproveTool {
		return mcp.ElicitResult{
			Action: "accept",
			Content: map[string]any{
				"answer": "always",
			},
		}, nil
	}

	p := tea.NewProgram(initialModel(request))
	m, err := p.Run()
	if err != nil {
		return mcp.ElicitResult{}, err
	}
	if m, ok := m.(model); ok {
		if m.err != nil {
			return mcp.ElicitResult{}, m.err
		}

		return mcp.ElicitResult{
			Action:  m.GetAction(),
			Content: m.Values(),
		}, m.Err()
	}

	// This should not happen, but if it does, we return a cancel action.
	return mcp.ElicitResult{
		Action: "cancel",
	}, nil
}

type (
	errMsg error
)

type model struct {
	form    *huh.Form
	action  string
	keys    []string
	request mcp.ElicitRequest
	err     error
}

func (m model) GetAction() string {
	if len(m.keys) == 0 {
		if b, ok := m.form.GetFocusedField().GetValue().(bool); ok && !b && m.action == "accept" {
			return "reject"
		}
	}
	return m.action
}

func (m model) Err() error {
	if m.err == nil {
		// might set m.err if there are validation errors
		_ = m.Values()
	}
	return m.err
}

func (m *model) Values() map[string]any {
	if m.err != nil {
		return nil
	}
	result := make(map[string]any, len(m.keys))
	for _, key := range m.keys {
		prop := m.request.RequestedSchema.Properties[key]
		if prop.Type == "boolean" {
			result[key] = m.form.GetBool(key)
		} else {
			v, err := validateFieldValue(prop, m.form.GetString(key))
			if err != nil {
				m.err = fmt.Errorf("invalid value for field %s: %w", key, err)
			}
			result[key] = v
		}
	}
	return result
}

func validateFieldValue(prop mcp.PrimitiveProperty, value string) (any, error) {
	switch prop.Type {
	case "string":
		if prop.MinLength != nil {
			if err := huh.ValidateMinLength(*prop.MinLength)(value); err != nil {
				return nil, err
			}
		}
		if prop.MaxLength != nil {
			if err := huh.ValidateMaxLength(*prop.MaxLength)(value); err != nil {
				return nil, err
			}
		}
		return value, nil
	case "number":
		i, err := strconv.ParseInt(value, 0, 0)
		if err == nil {
			if prop.Minimum != nil {
				if min, err := prop.Minimum.Int64(); err == nil {
					if i < min {
						return nil, fmt.Errorf("value must be greater than or equal to %d", min)
					}
				}
				if min, err := prop.Minimum.Float64(); err == nil {
					if f := float64(i); f < min {
						return nil, fmt.Errorf("value must be greater than or equal to %f", min)
					}
				}
			}
			if prop.Maximum != nil {
				if max, err := prop.Maximum.Int64(); err == nil {
					if i > max {
						return nil, fmt.Errorf("value must be less than or equal to %d", max)
					}
				}
				if max, err := prop.Maximum.Float64(); err == nil {
					if f := float64(i); f > max {
						return nil, fmt.Errorf("value must be less than or equal to %f", max)
					}
				}
			}
			return i, nil
		}

		f, fErr := strconv.ParseFloat(value, 0)
		if fErr != nil {
			return nil, fmt.Errorf("invalid number: %s", value)
		}
		if prop.Minimum != nil {
			if min, err := prop.Minimum.Float64(); err == nil {
				if f < min {
					return nil, fmt.Errorf("value must be greater than or equal to %f", min)
				}
			}
		}
		if prop.Maximum != nil {
			if max, err := prop.Maximum.Float64(); err == nil {
				if f > max {
					return nil, fmt.Errorf("value must be less than or equal to %f", max)
				}
			}
		}
		return f, nil
	case "integer":
		i, err := strconv.ParseInt(value, 0, 0)
		if err != nil {
			return nil, fmt.Errorf("invalid integer: %s", value)
		}

		if prop.Minimum != nil {
			if min, err := prop.Minimum.Int64(); err == nil {
				if i < min {
					return nil, fmt.Errorf("value must be greater than or equal to %d", min)
				}
			}
		}
		if prop.Maximum != nil {
			if max, err := prop.Maximum.Int64(); err == nil {
				if i > max {
					return nil, fmt.Errorf("value must be less than or equal to %d", max)
				}
			}
		}
		return i, nil
	case "boolean":
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b, nil
		}
		return nil, fmt.Errorf("invalid boolean: %s", value)
	case "enum":
		return value, nil
	}

	return "", fmt.Errorf("unknown type: %s", prop.Type)
}

func initialModel(request mcp.ElicitRequest) model {
	var (
		fields []huh.Field
		keys   []string
	)

	for name, prop := range request.RequestedSchema.Properties {
		keys = append(keys, name)

		switch prop.Type {
		case "enum":
			field := huh.NewSelect[string]().
				Key(name).
				Title(prop.Title).
				Description(prop.Description).
				Validate(func(v string) error {
					_, err := validateFieldValue(prop, v)
					return err
				})
			enums := prop.Enum
			names := prop.EnumNames
			if len(enums) == len(names) {
				opts := make([]huh.Option[string], len(enums))
				for i, e := range enums {
					opts[i] = huh.NewOption(names[i], e)
				}
				field = field.Options(opts...)
			} else {
				field = field.Options(huh.NewOptions(enums...)...)
			}
			fields = append(fields, field)
		case "boolean":
			confirm := huh.NewConfirm().
				Key(name).
				Title(prop.Title).
				Description(prop.Description).
				Value(prop.Default).
				Validate(func(v bool) error {
					_, err := validateFieldValue(prop, fmt.Sprint(v))
					return err
				})
			fields = append(fields, confirm)
		default:
			text := huh.NewInput().
				Key(name).
				Title(prop.Title).
				Description(prop.Description).
				Validate(func(v string) error {
					_, err := validateFieldValue(prop, v)
					return err
				})
			fields = append(fields, text)
		}
	}

	if len(keys) == 0 {
		fields = append(fields, huh.NewConfirm())
	}

	form := huh.
		NewForm(huh.NewGroup(fields...).
			Title(request.Message)).
		WithTheme(huh.ThemeBase())
	form.SubmitCmd = func() tea.Msg {
		return quitMsg{}
	}
	form.CancelCmd = func() tea.Msg {
		return interruptMsg{}
	}

	return model{
		form:    form,
		keys:    keys,
		request: request,
	}
}

func (m model) Init() tea.Cmd {
	return m.form.Init()
}

type quitMsg struct{}
type interruptMsg struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case quitMsg:
		m.action = "accept"
		cmds = append(cmds, tea.Quit)
	case interruptMsg:
		m.action = "cancel"
		cmds = append(cmds, tea.Quit)
	case errMsg:
		m.err = msg
		m.action = "cancel"
		cmds = append(cmds, tea.Quit)
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return "\n" + m.form.View()
}
