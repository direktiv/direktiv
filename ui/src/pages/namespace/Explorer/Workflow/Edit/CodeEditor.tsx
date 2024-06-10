import Editor, { EditorLanguagesType } from "~/design/Editor";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import { Bug } from "lucide-react";
import { Card } from "~/design/Card";
import { FC } from "react";
import useNavigationBlocker from "~/hooks/useNavigationBlocker";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooks/useUpdatedAt";

type EditorProps = {
  value: string;
  onValueChange: (value: string) => void;
  onSave: Parameters<typeof Editor>[0]["onSave"];
  hasUnsavedChanges: boolean;
  updatedAt: string | undefined;
  error: string | undefined;
  language: EditorLanguagesType;
  isTsWorkflow?: boolean;
};

export const CodeEditor: FC<EditorProps> = ({
  value,
  onValueChange,
  onSave,
  hasUnsavedChanges,
  updatedAt,
  error,
  language,
  isTsWorkflow = false,
}) => {
  const { t } = useTranslation();

  const theme = useTheme();
  const updatedAtInWords = useUpdatedAt(updatedAt);

  useNavigationBlocker(
    hasUnsavedChanges ? t("components.blocker.unsavedChangesWarning") : null
  );

  return (
    <Card className="flex grow flex-col p-4">
      <div className="grow" data-testid="workflow-editor">
        <Editor
          value={value}
          onMount={(editor) => {
            editor.focus();
          }}
          onChange={(newData) => {
            onValueChange(newData ?? "");
          }}
          theme={theme ?? undefined}
          onSave={onSave}
          language={language}
          isTsWorkflow={isTsWorkflow}
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
