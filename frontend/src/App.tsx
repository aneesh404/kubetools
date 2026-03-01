import { Clock3, ChevronRight, Link2, Plus, Search, X } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { FieldPanel } from "./components/features/FieldPanel";
import { YamlPanel } from "./components/features/YamlPanel";
import { generateYaml, highlightYaml } from "./lib/yaml";
import {
  generateYamlViaApi,
  getTemplates,
  listManifests,
  parseCrd,
  submitCrd,
  importCrdFromUrl,
  validateCrd,
  type ManifestHistoryItem,
  type ValidateCrdResponse
} from "./services/api";
import type { FieldDefinition, TemplateDefinition } from "./types/crd";

type StatusTone = "idle" | "ok" | "warn";
type AppStage = "landing" | "builder";

type SearchOption = {
  key: string;
  type: "template" | "manifest";
  label: string;
  subtitle: string;
  template?: TemplateDefinition;
  manifest?: ManifestHistoryItem;
};

interface StatusState {
  tone: StatusTone;
  text: string;
}

function isSchemaManifest(item: ManifestHistoryItem): boolean {
  return (
    item.kind.toLowerCase() === "customresourcedefinition" ||
    item.title.toLowerCase().startsWith("schema:")
  );
}

function cloneFields(fields: FieldDefinition[]): FieldDefinition[] {
  return fields.map((field) => ({ ...field }));
}

function upsertTemplate(
  list: TemplateDefinition[],
  template: TemplateDefinition
): TemplateDefinition[] {
  const exists = list.find((item) => item.id === template.id);
  if (!exists) {
    return [...list, template];
  }
  return list.map((item) => (item.id === template.id ? template : item));
}

function mergeManifestHistory(
  previous: ManifestHistoryItem[],
  incoming: ManifestHistoryItem[]
): ManifestHistoryItem[] {
  const byID = new Map<string, ManifestHistoryItem>();
  [...incoming, ...previous].forEach((item) => byID.set(item.id, item));
  return [...byID.values()]
    .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
    .slice(0, 80);
}

export default function App() {
  const searchInputRef = useRef<HTMLInputElement | null>(null);
  const [templates, setTemplates] = useState<TemplateDefinition[]>([]);
  const [activeTemplateId, setActiveTemplateId] = useState("");
  const [fields, setFields] = useState<FieldDefinition[]>([]);
  const [visibleFieldPaths, setVisibleFieldPaths] = useState<string[] | null>(null);
  const [manifestHistory, setManifestHistory] = useState<ManifestHistoryItem[]>([]);

  const [appStage, setAppStage] = useState<AppStage>("landing");
  const [status, setStatus] = useState<StatusState>({ tone: "idle", text: "" });
  const [isSyncingApi, setIsSyncingApi] = useState(false);
  const [yamlOutput, setYamlOutput] = useState("");

  const [searchQuery, setSearchQuery] = useState("");
  const [searchOpen, setSearchOpen] = useState(false);
  const [activeSearchIndex, setActiveSearchIndex] = useState(-1);
  const [showHistory, setShowHistory] = useState(false);

  const [isCrdModalOpen, setIsCrdModalOpen] = useState(false);
  const [crdTitle, setCrdTitle] = useState("");
  const [crdRaw, setCrdRaw] = useState("");
  const [crdUrl, setCrdUrl] = useState("");
  const [crdLint, setCrdLint] = useState<ValidateCrdResponse | null>(null);
  const [isLinting, setIsLinting] = useState(false);
  const [isImportingURL, setIsImportingURL] = useState(false);
  const [isUsingCRD, setIsUsingCRD] = useState(false);
  const [isSubmittingCRD, setIsSubmittingCRD] = useState(false);
  const [crdModalError, setCrdModalError] = useState("");
  const [crdModalInfo, setCrdModalInfo] = useState("");

  const activeTemplate = useMemo(() => {
    const selected = templates.find((item) => item.id === activeTemplateId);
    return selected ?? templates[0];
  }, [activeTemplateId, templates]);

  const effectiveFields = useMemo(() => {
    if (!visibleFieldPaths) {
      return fields;
    }
    const visible = new Set(visibleFieldPaths);
    return fields.filter((field) => visible.has(field.path));
  }, [fields, visibleFieldPaths]);

  const computedYaml = useMemo(() => {
    if (!activeTemplate) {
      return "";
    }
    return generateYaml(activeTemplate.apiVersion, activeTemplate.kind, effectiveFields);
  }, [activeTemplate, effectiveFields]);

  const yamlHtml = useMemo(() => highlightYaml(yamlOutput), [yamlOutput]);

  const searchOptions = useMemo<SearchOption[]>(() => {
    const templateOptions = templates.map((template) => ({
      key: `template-${template.id}`,
      type: "template" as const,
      label: template.kind,
      subtitle: `${template.apiVersion} · template`,
      template
    }));

    const historyOptions = manifestHistory
      .filter((manifest) => !isSchemaManifest(manifest))
      .map((manifest) => ({
      key: `manifest-${manifest.id}`,
      type: "manifest" as const,
      label: manifest.title,
      subtitle: `${manifest.resource} · saved`,
      manifest
      }));

    const combined = [...templateOptions, ...historyOptions];
    const query = searchQuery.trim().toLowerCase();

    if (query === "") {
      return [];
    }

    return combined
      .filter((option) => {
        return (
          option.label.toLowerCase().includes(query) || option.subtitle.toLowerCase().includes(query)
        );
      })
      .sort((left, right) => {
        const leftStarts = left.label.toLowerCase().startsWith(query) ? 1 : 0;
        const rightStarts = right.label.toLowerCase().startsWith(query) ? 1 : 0;
        return rightStarts - leftStarts;
      })
      .slice(0, 10);
  }, [manifestHistory, searchQuery, templates]);

  useEffect(() => {
    setYamlOutput(computedYaml);
  }, [computedYaml]);

  useEffect(() => {
    let active = true;

    const load = async () => {
      try {
        const [apiTemplates, history] = await Promise.all([getTemplates(), listManifests()]);
        if (!active) {
          return;
        }

        setTemplates(apiTemplates);
        if (apiTemplates.length > 0) {
          const preferred = apiTemplates.find((item) => item.id === activeTemplateId) ?? apiTemplates[0];
          setActiveTemplateId(preferred.id);
          setFields(cloneFields(preferred.defaultFields));
        } else {
          setActiveTemplateId("");
          setFields([]);
        }
        setManifestHistory(history);
        setStatus(
          apiTemplates.length > 0
            ? { tone: "ok", text: "Backend connected." }
            : { tone: "warn", text: "Backend connected but no templates found in MongoDB." }
        );
      } catch {
        if (!active) {
          return;
        }
        setStatus({
          tone: "warn",
          text: "Backend unavailable. Templates cannot be loaded."
        });
      }
    };

    void load();

    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    if (!isCrdModalOpen) {
      return;
    }

    const raw = crdRaw.trim();
    if (raw === "") {
      setCrdLint(null);
      setCrdModalError("");
      setCrdModalInfo("");
      return;
    }

    let canceled = false;
    const timer = window.setTimeout(async () => {
      setIsLinting(true);
      try {
        const lint = await validateCrd(raw);
        if (!canceled) {
          setCrdLint(lint);
          setCrdModalError("");
          if (lint.valid) {
            setCrdModalInfo("CRD passed static validation.");
          }
        }
      } catch {
        if (!canceled) {
          setCrdModalError("Linting failed. Ensure backend is reachable.");
          setCrdLint(null);
        }
      } finally {
        if (!canceled) {
          setIsLinting(false);
        }
      }
    }, 350);

    return () => {
      canceled = true;
      window.clearTimeout(timer);
    };
  }, [crdRaw, isCrdModalOpen]);

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      const isHotkey = (event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k";
      if (!isHotkey) {
        return;
      }

      const target = event.target as HTMLElement | null;
      const tagName = target?.tagName?.toLowerCase();
      if (tagName === "input" || tagName === "textarea" || target?.isContentEditable) {
        return;
      }

      event.preventDefault();
      setSearchOpen(true);
      searchInputRef.current?.focus();
      searchInputRef.current?.select();
    };

    window.addEventListener("keydown", handler);
    return () => {
      window.removeEventListener("keydown", handler);
    };
  }, []);

  useEffect(() => {
    if (!searchOpen || searchOptions.length === 0) {
      setActiveSearchIndex(-1);
      return;
    }

    setActiveSearchIndex((previous) => {
      if (previous < 0 || previous >= searchOptions.length) {
        return 0;
      }
      return previous;
    });
  }, [searchOpen, searchOptions]);

  function activateTemplate(template: TemplateDefinition): void {
    setActiveTemplateId(template.id);
    setFields(cloneFields(template.defaultFields));
    setSearchQuery(template.kind);
    setAppStage("builder");
    setStatus({ tone: "ok", text: `Loaded ${template.kind} builder.` });
  }

  function selectTemplate(templateId: string): void {
    const nextTemplate = templates.find((item) => item.id === templateId);
    if (!nextTemplate) {
      return;
    }
    activateTemplate(nextTemplate);
  }

  async function loadHistoryItem(item: ManifestHistoryItem): Promise<void> {
    const isCRDHistory =
      item.kind.toLowerCase() === "customresourcedefinition" ||
      item.apiVersion.toLowerCase().startsWith("apiextensions.k8s.io/");

    if (isCRDHistory) {
      try {
        const parsed = await parseCrd(item.yaml);
        setTemplates((previous) => upsertTemplate(previous, parsed.template));
        activateTemplate(parsed.template);
        setSearchOpen(false);
        setShowHistory(false);
        setStatus({ tone: "ok", text: `Loaded ${parsed.template.kind} builder from CRD history.` });
        return;
      } catch {
        setStatus({ tone: "warn", text: "Could not parse CRD history item. Loading saved YAML snapshot instead." });
      }
    }

    const templateMatch = templates.find(
      (template) =>
        template.kind.toLowerCase() === item.kind.toLowerCase() ||
        template.apiVersion.toLowerCase() === item.apiVersion.toLowerCase()
    );

    if (templateMatch) {
      setActiveTemplateId(templateMatch.id);
      setFields(cloneFields(templateMatch.defaultFields));
    } else {
      try {
        const schemaTemplate = await findTemplateFromSchemaHistory(item.kind);
        if (schemaTemplate) {
          setTemplates((previous) => upsertTemplate(previous, schemaTemplate));
          activateTemplate(schemaTemplate);
          setSearchOpen(false);
          setShowHistory(false);
          setStatus({ tone: "ok", text: `Recovered ${schemaTemplate.kind} builder from saved CRD schema.` });
          return;
        }

        const parsed = await parseCrd(item.yaml);
        setTemplates((previous) => upsertTemplate(previous, parsed.template));
        activateTemplate(parsed.template);
        setSearchOpen(false);
        setShowHistory(false);
        setStatus({ tone: "ok", text: `Derived builder form from saved ${item.kind} manifest.` });
        return;
      } catch {
        // Fallback to loading raw YAML snapshot below.
      }
    }

    setYamlOutput(item.yaml);
    setSearchQuery(item.title);
    setAppStage("builder");
    setSearchOpen(false);
    setShowHistory(false);
    setStatus({ tone: "ok", text: `Loaded history item: ${item.title}.` });
  }

  async function findTemplateFromSchemaHistory(kind: string): Promise<TemplateDefinition | null> {
    try {
      const related = await listManifests(kind);
      for (const entry of related) {
        if (!isSchemaManifest(entry)) {
          continue;
        }
        try {
          const parsed = await parseCrd(entry.yaml);
          if (parsed.template.kind.toLowerCase() === kind.toLowerCase()) {
            return parsed.template;
          }
        } catch {
          // Continue scanning candidates.
        }
      }
    } catch {
      // Ignore and fallback to local manifest parsing.
    }
    return null;
  }

  function chooseSearchOption(option: SearchOption): void {
    if (option.type === "template" && option.template) {
      activateTemplate(option.template);
      setSearchOpen(false);
      setActiveSearchIndex(-1);
      return;
    }

    if (option.type === "manifest" && option.manifest) {
      void loadHistoryItem(option.manifest);
      setActiveSearchIndex(-1);
      return;
    }
  }

  async function refreshHistory(query?: string): Promise<ManifestHistoryItem[]> {
    try {
      const history = await listManifests(query);
      setManifestHistory((previous) => mergeManifestHistory(previous, history));
      return history;
    } catch {
      setStatus({ tone: "warn", text: "Could not refresh manifest history." });
      return [];
    }
  }

  function updateField(index: number, value: string): void {
    setFields((previous) =>
      previous.map((field, fieldIndex) =>
        fieldIndex === index
          ? {
              ...field,
              value
            }
          : field
      )
    );
  }

  function addField(path: string): void {
    const normalized = path.trim();
    if (!normalized || !activeTemplate) {
      return;
    }

    if (fields.some((field) => field.path === normalized)) {
      setStatus({ tone: "warn", text: "Field already exists in this resource form." });
      return;
    }

    const optional = activeTemplate.optionalFields.find((field) => field.path === normalized);
    if (optional) {
      setFields((previous) => [...previous, { ...optional, value: "" }]);
      setStatus({ tone: "ok", text: `Added optional field ${normalized}.` });
      return;
    }

    setFields((previous) => [
      ...previous,
      {
        path: normalized,
        label: normalized,
        value: "",
        description: "Custom field. Validate against your CRD schema before applying."
      }
    ]);

    setStatus({ tone: "ok", text: `Added custom field ${normalized}.` });
  }

  async function copyYaml(): Promise<void> {
    try {
      await navigator.clipboard.writeText(yamlOutput);
      setStatus({ tone: "ok", text: "YAML copied to clipboard." });
    } catch {
      setStatus({ tone: "warn", text: "Clipboard is not available in this browser context." });
    }
  }

  async function copyApplyCommand(): Promise<void> {
    const namespaceField = effectiveFields.find((field) => field.path === "metadata.namespace");
    const namespace = namespaceField?.value?.trim() ?? "";
    const namespaceFlag = namespace ? ` -n ${namespace}` : "";
    const command = `kubectl apply${namespaceFlag} -f - <<'EOF'\n${yamlOutput.trimEnd()}\nEOF`;

    try {
      await navigator.clipboard.writeText(command);
      setStatus({ tone: "ok", text: "kubectl apply command copied to clipboard." });
    } catch {
      setStatus({ tone: "warn", text: "Clipboard is not available in this browser context." });
    }
  }

  function resetFields(): void {
    if (!activeTemplate) {
      return;
    }
    setFields(cloneFields(activeTemplate.defaultFields));
    setStatus({ tone: "ok", text: "Fields reset to template defaults." });
  }

  async function syncYamlWithApi(): Promise<void> {
    if (!activeTemplate) {
      return;
    }

    setIsSyncingApi(true);
    try {
      const response = await generateYamlViaApi({
        apiVersion: activeTemplate.apiVersion,
        kind: activeTemplate.kind,
        fields: effectiveFields
      });
      setYamlOutput(response.yaml);
      setStatus({ tone: "ok", text: "YAML generated." });
    } catch {
      setStatus({ tone: "warn", text: "YAML generation failed. Check backend availability." });
    } finally {
      setIsSyncingApi(false);
    }
  }

  function closeAndResetCRDModal(): void {
    setIsCrdModalOpen(false);
    setCrdTitle("");
    setCrdRaw("");
    setCrdUrl("");
    setCrdLint(null);
    setCrdModalError("");
    setCrdModalInfo("");
  }

  async function useCRDFromModal(): Promise<void> {
    const raw = crdRaw.trim();
    if (raw === "") {
      setCrdModalError("Paste your CRD before using it.");
      return;
    }

    if (crdLint && !crdLint.valid) {
      setCrdModalError("Fix validation errors before using this CRD.");
      return;
    }

    setIsUsingCRD(true);
    try {
      const response = await parseCrd(raw);
      setTemplates((previous) => upsertTemplate(previous, response.template));
      activateTemplate(response.template);
      setStatus({ tone: "ok", text: `${response.template.kind} loaded for editing (not saved).` });
      closeAndResetCRDModal();
    } catch {
      setCrdModalError("Could not parse CRD. Check schema and try again.");
    } finally {
      setIsUsingCRD(false);
    }
  }

  async function submitCRDFromModal(): Promise<void> {
    const raw = crdRaw.trim();
    if (raw === "") {
      setCrdModalError("Paste your CRD before submitting.");
      return;
    }

    if (crdLint && !crdLint.valid) {
      setCrdModalError("Fix validation errors before submitting.");
      return;
    }

    setIsSubmittingCRD(true);
    try {
      const response = await submitCrd({
        title: crdTitle,
        raw
      });

      if (!response.validation.valid) {
        setCrdLint(response.validation);
        setCrdModalError("Submission blocked due to validation errors.");
        return;
      }

      setTemplates((previous) => upsertTemplate(previous, response.template));
      activateTemplate(response.template);
      setStatus({ tone: "ok", text: `${response.template.kind} saved and loaded from MongoDB.` });
      closeAndResetCRDModal();
    } catch {
      setCrdModalError("Submission failed. Check backend and MongoDB connectivity.");
    } finally {
      setIsSubmittingCRD(false);
    }
  }

  async function importCRDFromURL(): Promise<void> {
    const url = crdUrl.trim();
    if (url === "") {
      setCrdModalError("Paste a URL first.");
      return;
    }

    setIsImportingURL(true);
    setCrdModalError("");
    setCrdModalInfo("");
    try {
      const response = await importCrdFromUrl(url);
      setCrdRaw(response.raw);
      setCrdLint(response.validation);
      if (response.validation.valid) {
        setCrdModalInfo(`Imported from ${response.sourceUrl}`);
      } else {
        setCrdModalError("Imported document has validation errors. Review and fix before submit.");
      }
    } catch (error) {
      setCrdModalError(error instanceof Error ? error.message : "Failed to import CRD URL.");
    } finally {
      setIsImportingURL(false);
    }
  }

  async function runSearch(): Promise<void> {
    const query = searchQuery.trim();
    if (query === "") {
      setSearchOpen(false);
      setActiveSearchIndex(-1);
      return;
    }

    if (searchOptions.length > 0) {
      chooseSearchOption(searchOptions[0]);
      return;
    }

    const matches = await refreshHistory(query);
    if (matches.length > 0) {
      chooseSearchOption({
        key: `manifest-${matches[0].id}`,
        type: "manifest",
        label: matches[0].title,
        subtitle: `${matches[0].resource} · saved`,
        manifest: matches[0]
      });
      return;
    }

    setStatus({ tone: "warn", text: "No matching resource found." });
  }

  function renderSearchInput(variant: "landing" | "top"): JSX.Element {
    return (
      <div
        className={`resource-search ${variant}`}
        onClick={(event) => {
          const target = event.target as HTMLElement;
          if (target.closest("button")) {
            return;
          }
          searchInputRef.current?.focus();
        }}
      >
        <Search size={18} className="search-icon" />
        <input
          ref={searchInputRef}
          value={searchQuery}
          onFocus={() => {
            const hasValue = searchQuery.trim().length > 0;
            setSearchOpen(hasValue);
            if (hasValue && searchOptions.length > 0) {
              setActiveSearchIndex(0);
            }
          }}
          onBlur={() => {
            window.setTimeout(() => setSearchOpen(false), 120);
          }}
          onChange={(event) => {
            const value = event.target.value;
            setSearchQuery(value);
            const hasValue = value.trim().length > 0;
            setSearchOpen(hasValue);
            setActiveSearchIndex(hasValue ? 0 : -1);
          }}
          onKeyDown={(event) => {
            if (event.key === "ArrowDown") {
              if (searchOptions.length > 0) {
                event.preventDefault();
                setSearchOpen(true);
                setActiveSearchIndex((previous) => (previous + 1 + searchOptions.length) % searchOptions.length);
              }
              return;
            }

            if (event.key === "ArrowUp") {
              if (searchOptions.length > 0) {
                event.preventDefault();
                setSearchOpen(true);
                setActiveSearchIndex((previous) => {
                  if (previous < 0) {
                    return searchOptions.length - 1;
                  }
                  return (previous - 1 + searchOptions.length) % searchOptions.length;
                });
              }
              return;
            }

            if (event.key === "Escape") {
              if (searchOpen) {
                event.preventDefault();
                setSearchOpen(false);
                setActiveSearchIndex(-1);
              }
              return;
            }

            if (event.key === "Enter") {
              event.preventDefault();
              if (searchOpen && activeSearchIndex >= 0 && activeSearchIndex < searchOptions.length) {
                chooseSearchOption(searchOptions[activeSearchIndex]);
                return;
              }
              void runSearch();
            }
          }}
          placeholder="Search manifest type (Deployment, VolumeSnapshot, Widget...)"
          aria-label="Search resource type"
        />
        {variant === "landing" ? (
          <button
            type="button"
            className="search-kbd search-aux-action"
            onMouseDown={(event) => {
              event.preventDefault();
            }}
            onClick={() => {
              setIsCrdModalOpen(true);
              setCrdModalError("");
              setCrdModalInfo("");
            }}
          >
            + Add templates or manifests
          </button>
        ) : (
          <span className="search-kbd" aria-hidden="true">
            ⌘K
          </span>
        )}
        <button
          type="button"
          className="search-submit"
          onMouseDown={(event) => {
            event.preventDefault();
          }}
          onClick={() => {
            void runSearch();
          }}
          aria-label="Submit search"
        >
          <ChevronRight size={16} />
        </button>

        {searchOpen && searchQuery.trim().length > 0 ? (
          <div className="search-dropdown">
            {searchOptions.length === 0 ? (
              <p className="search-empty">No quick matches</p>
            ) : (
              searchOptions.map((option) => (
                <button
                  key={option.key}
                  type="button"
                  className={`search-option ${
                    searchOptions[activeSearchIndex]?.key === option.key ? "active" : ""
                  }`}
                  onMouseEnter={() => {
                    const optionIndex = searchOptions.findIndex((item) => item.key === option.key);
                    if (optionIndex >= 0) {
                      setActiveSearchIndex(optionIndex);
                    }
                  }}
                  onMouseDown={(event) => {
                    event.preventDefault();
                    chooseSearchOption(option);
                  }}
                >
                  <span>{option.label}</span>
                  <small>{option.subtitle}</small>
                </button>
              ))
            )}
          </div>
        ) : null}
      </div>
    );
  }

  return (
    <div className={`app-shell ${appStage === "landing" ? "is-landing" : "is-builder"}`}>
      {appStage === "landing" ? (
        <main className="landing-page">
          <section className="landing-center">
            <h1 className="landing-brand">KubeBuilder</h1>

            <div className="landing-search-row">{renderSearchInput("landing")}</div>

            <p className="landing-subtitle">
              Search for a resource type, start building instantly, and ship validated manifests with confidence.
            </p>
          </section>
        </main>
      ) : (
        <>
          <header className="top-nav card">
            <div className="top-nav-left">
              <div className="brand-lockup">
                <div className="brand-mark">KB</div>
                <button
                  type="button"
                  className="brand-link"
                  onClick={() => {
                    setAppStage("landing");
                    setSearchOpen(false);
                  }}
                >
                  KubeBuilder
                </button>
              </div>
            </div>

            <div className="top-nav-right">
              {renderSearchInput("top")}

              <div className="history-wrap">
                <button
                  type="button"
                  className="ghost-btn"
                  onClick={() => {
                    setShowHistory((previous) => !previous);
                    if (!showHistory) {
                      void refreshHistory();
                    }
                  }}
                >
                  <Clock3 size={15} /> History
                </button>

                {showHistory ? (
                  <div className="history-dropdown">
                    {manifestHistory.filter((item) => !isSchemaManifest(item)).length === 0 ? (
                      <p className="history-empty">No manifests in history yet.</p>
                    ) : (
                      manifestHistory
                        .filter((item) => !isSchemaManifest(item))
                        .map((item) => (
                        <button
                          key={item.id}
                          type="button"
                          onClick={() => {
                            void loadHistoryItem(item);
                          }}
                        >
                          <span>{item.title}</span>
                          <small>{new Date(item.createdAt).toLocaleString()}</small>
                        </button>
                        ))
                    )}
                  </div>
                ) : null}
              </div>

              <button
                type="button"
                className="ghost-btn plus-mini"
                onClick={() => {
                  setIsCrdModalOpen(true);
                  setCrdModalError("");
                  setCrdModalInfo("");
                }}
              >
                <Plus size={15} /> Add CRD
              </button>
            </div>
          </header>

          <section className="builder-shell">
            <article className="builder-header card">
              <div>
                <h2>Builder Workspace</h2>
                <p>Focused editing for generated custom resource YAML.</p>
              </div>
              <span className="resource-tag">
                {activeTemplate?.kind} ({activeTemplate?.apiVersion})
              </span>
            </article>

            <section className="workspace-grid">
              <FieldPanel
                fields={fields}
                optionalFields={activeTemplate?.optionalFields ?? []}
                requiredFieldPaths={activeTemplate?.defaultFields.map((field) => field.path) ?? []}
                onVisiblePathsChange={setVisibleFieldPaths}
                onUpdateField={updateField}
                onAddField={addField}
              />

              <YamlPanel
                html={yamlHtml}
                onCopy={() => void copyYaml()}
                onCopyApply={() => void copyApplyCommand()}
                onReset={resetFields}
                onSyncApi={() => void syncYamlWithApi()}
                syncing={isSyncingApi}
              />
            </section>
          </section>

          <footer className={`status-row ${status.tone}`}>
            <span className="resource-tag">
              Resource: {activeTemplate?.kind} ({activeTemplate?.apiVersion})
            </span>
            <span>{status.text}</span>
          </footer>
        </>
      )}

      {isCrdModalOpen ? (
        <div
          className="modal-backdrop"
          onClick={() => {
            setIsCrdModalOpen(false);
          }}
        >
          <section
            className="modal-card"
            onClick={(event) => {
              event.stopPropagation();
            }}
          >
            <div className="modal-header">
              <h3>Submit Custom CRD</h3>
              <button
                type="button"
                className="modal-close-btn"
                onClick={() => {
                  setIsCrdModalOpen(false);
                }}
                aria-label="Close modal"
              >
                <X size={18} />
              </button>
            </div>

            <input
              value={crdTitle}
              onChange={(event) => setCrdTitle(event.target.value)}
              placeholder="Optional title"
            />

            <div className="modal-url-row">
              <input
                value={crdUrl}
                onChange={(event) => setCrdUrl(event.target.value)}
                placeholder="Or paste a CRD URL (raw.githubusercontent.com, GitHub blob, ...)"
              />
              <button
                type="button"
                className="ghost-btn import-url-btn"
                onClick={() => {
                  void importCRDFromURL();
                }}
                disabled={isImportingURL}
              >
                <Link2 size={14} />
                {isImportingURL ? "Importing..." : "Import URL"}
              </button>
            </div>

            <textarea
              className="modal-textarea"
              value={crdRaw}
              onChange={(event) => setCrdRaw(event.target.value)}
              placeholder="Paste your CRD YAML here..."
            />

            {isLinting ||
            crdModalError ||
            crdModalInfo ||
            (crdLint && (crdLint.errors.length > 0 || crdLint.warnings.length > 0)) ? (
              <div className="lint-box">
                {isLinting ? <p>Running static checks...</p> : null}
                {crdModalInfo ? <p className="lint-ok">{crdModalInfo}</p> : null}
                {crdLint && crdLint.errors.length > 0 ? (
                  <ul className="lint-errors">
                    {crdLint.errors.map((error) => (
                      <li key={error}>{error}</li>
                    ))}
                  </ul>
                ) : null}
                {crdLint && crdLint.warnings.length > 0 ? (
                  <ul className="lint-warnings">
                    {crdLint.warnings.map((warning) => (
                      <li key={warning}>{warning}</li>
                    ))}
                  </ul>
                ) : null}
                {crdModalError ? <p className="lint-error-text">{crdModalError}</p> : null}
              </div>
            ) : null}

            <div className="modal-actions">
              <button
                type="button"
                className="ghost-btn"
                onClick={() => {
                  void (async () => {
                    setIsLinting(true);
                    setCrdModalError("");
                    setCrdModalInfo("");
                    try {
                      const lint = await validateCrd(crdRaw);
                      setCrdLint(lint);
                      if (lint.valid) {
                        setCrdModalInfo("CRD passed static validation.");
                      }
                    } catch {
                      setCrdModalError("Manual validation failed. Backend may be unavailable.");
                    } finally {
                      setIsLinting(false);
                    }
                  })();
                }}
              >
                Validate
              </button>
              <button
                type="button"
                className="primary-btn"
                disabled={
                  isUsingCRD || isSubmittingCRD || isLinting || isImportingURL || (crdLint !== null && !crdLint.valid)
                }
                onClick={() => {
                  void useCRDFromModal();
                }}
              >
                {isUsingCRD ? "Using..." : "Use CRD"}
              </button>
              <button
                type="button"
                className="primary-btn"
                disabled={
                  isUsingCRD || isSubmittingCRD || isLinting || isImportingURL || (crdLint !== null && !crdLint.valid)
                }
                onClick={() => {
                  void submitCRDFromModal();
                }}
              >
                {isSubmittingCRD ? "Saving..." : "Save + Use CRD"}
              </button>
            </div>
          </section>
        </div>
      ) : null}
    </div>
  );
}
