import Editor, { EditorLanguagesType, ExtraLibsType } from "~/design/Editor";
import { FC, useEffect, useRef } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import { Bug } from "lucide-react";
import { Card } from "~/design/Card";
import { editor } from "monaco-editor";
import useNavigationBlocker from "~/hooks/useNavigationBlocker";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooks/useUpdatedAt";

type EditorProps = {
  value: string;
  onValueChange: (value: string) => void;
  onSave: Parameters<typeof Editor>[0]["onSave"];
  hasUnsavedChanges?: boolean;
  updatedAt: string | undefined;
  error?: string;
  language: EditorLanguagesType;
  tsLibs?: ExtraLibsType;
  markers?: editor.IMarkerData[];
};

export const CodeEditor: FC<EditorProps> = ({
  value,
  onValueChange,
  onSave,
  hasUnsavedChanges = false,
  updatedAt,
  error,
  language,
  tsLibs = [],
  markers = [],
}) => {
  const { t } = useTranslation();
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);

  const theme = useTheme();
  const updatedAtInWords = useUpdatedAt(updatedAt);

  useNavigationBlocker(
    hasUnsavedChanges ? t("components.blocker.unsavedChangesWarning") : null
  );

  useEffect(() => {
    const model = editorRef.current?.getModel();
    if (!model) return;
    editor.setModelMarkers(model, "workflow-validation", markers);
  }, [markers]);

  return (
    <Card className="flex grow flex-col p-4">
      <div className="grow" data-testid="workflow-editor">
        <Editor
          value={value}
          onMount={(editor) => {
            editor.focus();
            editorRef.current = editor;
          }}
          onChange={(newData) => {
            onValueChange(newData ?? "");
          }}
          theme={theme ?? undefined}
          onSave={onSave}
          language={language}
          tsLibs={tsLibs}
        />
      </div>
      <div
        className="flex justify-between gap-2 pt-2 text-sm text-gray-8 dark:text-gray-dark-8"
        data-testid="workflow-txt-updated"
      >
        {updatedAt && !error && (
          <>
            {t("pages.explorer.workflow.updated", {
              relativeTime: updatedAtInWords,
            })}
          </>
        )}
        {error && (
          <Popover defaultOpen>
            <PopoverTrigger asChild>
              <span className="flex items-center gap-x-1 text-danger-11 dark:text-danger-dark-11">
                <Bug className="h-5" />
                {t("pages.explorer.workflow.editor.theresOneIssue")}
              </span>
            </PopoverTrigger>
            <PopoverContent asChild>
              <div className="flex p-4">
                <div className="grow">{error}</div>
              </div>
            </PopoverContent>
          </Popover>
        )}

        {hasUnsavedChanges && (
          <span className="text-center">
            {t("pages.explorer.workflow.editor.unsavedNote")}
          </span>
        )}
      </div>
    </Card>
  );
};
