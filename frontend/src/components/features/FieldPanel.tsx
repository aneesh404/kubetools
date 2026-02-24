import { useMemo, useState } from "react";
import type { FieldDefinition } from "../../types/crd";

interface FieldPanelProps {
  fields: FieldDefinition[];
  optionalFields: FieldDefinition[];
  onUpdateField: (index: number, value: string) => void;
  onAddField: (path: string) => void;
}

export function FieldPanel({
  fields,
  optionalFields,
  onUpdateField,
  onAddField
}: FieldPanelProps) {
  const [newFieldPath, setNewFieldPath] = useState("");
  const [openDescriptions, setOpenDescriptions] = useState<Record<number, boolean>>({});

  const suggestions = useMemo(() => {
    const existing = new Set(fields.map((field) => field.path));
    return optionalFields.filter((field) => !existing.has(field.path));
  }, [fields, optionalFields]);

  return (
    <article className="panel form-panel">
      <div className="panel-header">
        <h3>Custom Resource Form</h3>
      </div>

      <div className="field-list">
        {fields.map((field, index) => {
          const isOpen = openDescriptions[index] ?? false;
          return (
            <div key={`${field.path}-${index}`} className="field-card">
              <div className="field-row">
                <label className="field-label" htmlFor={`field-${index}`}>
                  {field.label ?? field.path}
                </label>
                <button
                  className="info-trigger"
                  type="button"
                  onClick={() =>
                    setOpenDescriptions((state) => ({
                      ...state,
                      [index]: !isOpen
                    }))
                  }
                  aria-label="Toggle field description"
                >
                  i
                </button>
              </div>

              <input
                id={`field-${index}`}
                value={field.value ?? ""}
                onChange={(event) => onUpdateField(index, event.target.value)}
                placeholder={`Set ${field.path}`}
              />

              {isOpen ? <p className="field-description">{field.description}</p> : null}
            </div>
          );
        })}
      </div>

      <div className="add-field-box">
        <h4>Add Field (Autocomplete)</h4>
        <div className="add-field-row">
          <input
            value={newFieldPath}
            onChange={(event) => setNewFieldPath(event.target.value)}
            list="field-suggestions"
            placeholder="spec.volumeSnapshotClassName"
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                event.preventDefault();
                onAddField(newFieldPath);
                setNewFieldPath("");
              }
            }}
          />
          <button
            type="button"
            onClick={() => {
              onAddField(newFieldPath);
              setNewFieldPath("");
            }}
          >
            Add
          </button>
        </div>

        <datalist id="field-suggestions">
          {suggestions.map((field) => (
            <option key={field.path} value={field.path} />
          ))}
        </datalist>
      </div>
    </article>
  );
}
