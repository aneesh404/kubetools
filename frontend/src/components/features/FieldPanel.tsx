import { useMemo, useState } from "react";
import type { FieldDefinition } from "../../types/crd";

interface FieldPanelProps {
  fields: FieldDefinition[];
  optionalFields: FieldDefinition[];
  requiredFieldPaths: string[];
  onUpdateField: (index: number, value: string) => void;
  onAddField: (path: string) => void;
}

interface GroupedField {
  index: number;
  field: FieldDefinition;
  relativePath: string;
  isRequired: boolean;
}

interface FieldGroup {
  key: string;
  title: string;
  fields: GroupedField[];
}

function deriveGroupKey(path: string): string {
  const segments = path.split(".").filter((segment) => segment.trim() !== "");
  if (segments.length === 0) {
    return "root";
  }
  if (segments[0] === "spec") {
    if (segments.length >= 3) {
      return `spec.${segments[1]}`;
    }
    return "spec";
  }
  if (segments[0] === "metadata") {
    return "metadata";
  }
  return segments[0];
}

function deriveGroupTitle(groupKey: string): string {
  if (groupKey.startsWith("spec.")) {
    return groupKey.slice("spec.".length);
  }
  return groupKey;
}

function deriveRelativePath(fieldPath: string, groupKey: string): string {
  if (fieldPath === groupKey) {
    return "(value)";
  }
  const prefix = `${groupKey}.`;
  if (fieldPath.startsWith(prefix)) {
    return fieldPath.slice(prefix.length);
  }
  return fieldPath;
}

export function FieldPanel({
  fields,
  optionalFields,
  requiredFieldPaths,
  onUpdateField,
  onAddField
}: FieldPanelProps) {
  const [newFieldPath, setNewFieldPath] = useState("");
  const [openDescriptions, setOpenDescriptions] = useState<Record<number, boolean>>({});
  const [collapsedGroups, setCollapsedGroups] = useState<Record<string, boolean>>({});

  const suggestions = useMemo(() => {
    const existing = new Set(fields.map((field) => field.path));
    return optionalFields.filter((field) => !existing.has(field.path));
  }, [fields, optionalFields]);

  const requiredSet = useMemo(() => new Set(requiredFieldPaths), [requiredFieldPaths]);

  const groupedFields = useMemo<FieldGroup[]>(() => {
    const groups = new Map<string, FieldGroup>();
    fields.forEach((field, index) => {
      const groupKey = deriveGroupKey(field.path);
      const existing = groups.get(groupKey);
      const item: GroupedField = {
        index,
        field,
        relativePath: deriveRelativePath(field.path, groupKey),
        isRequired: requiredSet.has(field.path)
      };
      if (!existing) {
        groups.set(groupKey, {
          key: groupKey,
          title: deriveGroupTitle(groupKey),
          fields: [item]
        });
        return;
      }
      existing.fields.push(item);
    });
    return [...groups.values()];
  }, [fields, requiredSet]);

  return (
    <article className="panel form-panel">
      <div className="panel-header">
        <h3>Custom Resource Form</h3>
      </div>

      <div className="field-list">
        {groupedFields.map((group) => (
          <section key={group.key} className="field-group">
            <div className="field-group-header">
              <div className="field-group-header-left">
                <h4>{group.title}</h4>
                <span>{group.key}</span>
              </div>
              <button
                type="button"
                className={`group-toggle ${collapsedGroups[group.key] ? "collapsed" : "expanded"}`}
                disabled={group.fields.every((item) => item.isRequired)}
                onClick={() =>
                  setCollapsedGroups((state) => ({
                    ...state,
                    [group.key]: !(state[group.key] ?? false)
                  }))
                }
                aria-label={`Toggle ${group.title} fields`}
                aria-expanded={!collapsedGroups[group.key]}
              >
                <span className="group-toggle-track">
                  <span className="group-toggle-thumb" />
                </span>
              </button>
            </div>

            <div className="field-group-list">
              {(collapsedGroups[group.key]
                ? group.fields.filter((item) => item.isRequired)
                : group.fields
              ).map(({ field, index, relativePath, isRequired }) => {
                const isOpen = openDescriptions[index] ?? false;
                return (
                  <div key={`${field.path}-${index}`} className="field-card">
                    <div className="field-row">
                      <label className="field-label" htmlFor={`field-${index}`}>
                        {field.label ?? relativePath}
                        {isRequired ? <span className="field-required-chip">Required</span> : null}
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

                    <p className="field-path-hint">{field.path}</p>
                    {isOpen ? <p className="field-description">{field.description}</p> : null}
                  </div>
                );
              })}
            </div>
          </section>
        ))}
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
