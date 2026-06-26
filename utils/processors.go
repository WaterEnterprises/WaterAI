// --- START OF FILE utils/processors.go ---
package utils

import (
	"fmt"
)

// --- Interfaces ---

type SystemPromptBuilder interface {
	UpdateWebDevRules(rules string)
}

type Processor interface {
	InstallDependencies() error
	GetProcessorMessage() string
	StartUpProject() error
}

// --- Registry ---

var ProcessorRegistry = map[string]func(*WorkspaceManager, *TerminalManager, SystemPromptBuilder, string) Processor{
	"nextjs-shadcn":        NewNextShadcnProcessor,
	"react-tailwind-python": NewReactTailwindProcessor,
	"react-vite-shadcn":    NewReactViteProcessor,
}

// --- Base Logic ---

type BaseProcessor struct {
	WM           *WorkspaceManager
	TM           *TerminalManager
	PromptBuilder SystemPromptBuilder
	ProjectName   string
	TemplateName  string
	Rules         string
}

func (p *BaseProcessor) CopyTemplate() error {
	cmd := fmt.Sprintf("cp -rf /app/templates/%s %s", p.TemplateName, p.ProjectName)
	res := p.TM.Exec("system", cmd, 60)
	if !res.Success {
		return fmt.Errorf("copy failed: %s", res.Output)
	}
	return nil
}

func (p *BaseProcessor) GetProcessorMessage() string {
	return p.Rules
}

// --- Implementations ---

// 1. NextJS ShadCN
type NextShadcnProcessor struct {
	BaseProcessor
}

func NewNextShadcnProcessor(wm *WorkspaceManager, tm *TerminalManager, sp SystemPromptBuilder, name string) Processor {
	rules := fmt.Sprintf("Project directory `%s` created. Next.js App Router. Use `bun run dev`.", name)
	return &NextShadcnProcessor{
		BaseProcessor{WM: wm, TM: tm, PromptBuilder: sp, ProjectName: name, TemplateName: "nextjs-shadcn", Rules: rules},
	}
}

func (p *NextShadcnProcessor) StartUpProject() error {
	if err := p.CopyTemplate(); err != nil {
		return err
	}
	if err := p.InstallDependencies(); err != nil {
		return err
	}
	if p.PromptBuilder != nil {
		p.PromptBuilder.UpdateWebDevRules(p.GetProcessorMessage())
	}
	return nil
}

func (p *NextShadcnProcessor) InstallDependencies() error {
	cmd := fmt.Sprintf("cd %s && bun install", p.ProjectName)
	res := p.TM.Exec("system", cmd, 300)
	if !res.Success {
		return fmt.Errorf("install failed: %s", res.Output)
	}
	return nil
}

// 2. React Tailwind Python
type ReactTailwindProcessor struct {
	BaseProcessor
}

func NewReactTailwindProcessor(wm *WorkspaceManager, tm *TerminalManager, sp SystemPromptBuilder, name string) Processor {
	rules := fmt.Sprintf("Backend/Frontend project `%s`. backend:8080, frontend:3030.", name)
	return &ReactTailwindProcessor{
		BaseProcessor{WM: wm, TM: tm, PromptBuilder: sp, ProjectName: name, TemplateName: "react-tailwind-python", Rules: rules},
	}
}

func (p *ReactTailwindProcessor) StartUpProject() error {
	if err := p.CopyTemplate(); err != nil {
		return err
	}
	if err := p.InstallDependencies(); err != nil {
		return err
	}
	if p.PromptBuilder != nil {
		p.PromptBuilder.UpdateWebDevRules(p.GetProcessorMessage())
	}
	return nil
}

func (p *ReactTailwindProcessor) InstallDependencies() error {
	// Front
	res := p.TM.Exec("system", fmt.Sprintf("cd %s/frontend && bun install", p.ProjectName), 300)
	if !res.Success { return fmt.Errorf("frontend install failed: %s", res.Output) }
	// Back
	res = p.TM.Exec("system", fmt.Sprintf("cd %s/backend && pip install -r requirements.txt", p.ProjectName), 300)
	if !res.Success { return fmt.Errorf("backend install failed: %s", res.Output) }
	return nil
}

// 3. React Vite ShadCN
type ReactViteProcessor struct {
	BaseProcessor
}

func NewReactViteProcessor(wm *WorkspaceManager, tm *TerminalManager, sp SystemPromptBuilder, name string) Processor {
	rules := fmt.Sprintf("Vite React project `%s`. Use `bun run dev`.", name)
	return &ReactViteProcessor{
		BaseProcessor{WM: wm, TM: tm, PromptBuilder: sp, ProjectName: name, TemplateName: "react-vite-shadcn", Rules: rules},
	}
}

func (p *ReactViteProcessor) StartUpProject() error {
	if err := p.CopyTemplate(); err != nil {
		return err
	}
	if err := p.InstallDependencies(); err != nil {
		return err
	}
	if p.PromptBuilder != nil {
		p.PromptBuilder.UpdateWebDevRules(p.GetProcessorMessage())
	}
	return nil
}

func (p *ReactViteProcessor) InstallDependencies() error {
	cmd := fmt.Sprintf("cd %s && bun install", p.ProjectName)
	res := p.TM.Exec("system", cmd, 300)
	if !res.Success {
		return fmt.Errorf("install failed: %s", res.Output)
	}
	return nil
}