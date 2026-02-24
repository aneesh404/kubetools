import type { TemplateDefinition } from "../../types/crd";

interface TemplateSelectorProps {
  templates: TemplateDefinition[];
  activeTemplateId: string;
  onSelect: (id: string) => void;
}

export function TemplateSelector({
  templates,
  activeTemplateId,
  onSelect
}: TemplateSelectorProps) {
  return (
    <div className="template-grid">
      {templates.map((template) => {
        const isActive = template.id === activeTemplateId;
        return (
          <button
            key={template.id}
            className={`template-chip ${isActive ? "active" : ""}`}
            onClick={() => onSelect(template.id)}
            type="button"
          >
            {template.title}
          </button>
        );
      })}
    </div>
  );
}
