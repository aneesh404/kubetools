interface YamlPanelProps {
  html: string;
  onCopy: () => void;
  onReset: () => void;
  onSyncApi: () => void;
  syncing: boolean;
}

export function YamlPanel({
  html,
  onCopy,
  onReset,
  onSyncApi,
  syncing
}: YamlPanelProps) {
  return (
    <article className="panel yaml-panel">
      <div className="panel-header">
        <h3>Generated YAML</h3>
        <div className="actions">
          <button type="button" onClick={onSyncApi} disabled={syncing}>
            {syncing ? "Generating..." : "Generate & Save"}
          </button>
          <button type="button" onClick={onCopy}>
            Copy
          </button>
          <button type="button" onClick={onReset}>
            Reset
          </button>
        </div>
      </div>
      <pre
        className="yaml-preview"
        dangerouslySetInnerHTML={{
          __html: html
        }}
      />
    </article>
  );
}
