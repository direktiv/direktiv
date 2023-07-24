import { Dispatch, FC, SetStateAction } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import { Bug } from "lucide-react";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

type EditorProps = {
  value: string;
  setValue: Dispatch<SetStateAction<string>>;
  onSave: Parameters<typeof Editor>[0]["onSave"];
  hasUnsavedChanged: boolean;
  createdAt: string | undefined;
  error: string | undefined;
};

export const CodeEditor: FC<EditorProps> = ({
  value,
  setValue,
  onSave,
  hasUnsavedChanged,
  createdAt,
  error,
}) => {
  const { t } = useTranslation();
  const theme = useTheme();
  const updatedAt = useUpdatedAt(createdAt);

  return (
    <Card className="flex grow flex-col p-4" data-testid="workflow-editor">
      <div className="grow">
        <Editor
          value={value}
          onMount={(editor) => {
            editor.focus();
          }}
          onChange={(newData) => {
            setValue(newData ?? "");
          }}
          theme={theme ?? undefined}
          onSave={onSave}
        />
      </div>
      <div
        className="flex justify-between gap-2 pt-2 text-sm text-gray-8 dark:text-gray-dark-8"
        data-testid="workflow-txt-updated"
      >
        {createdAt && !error && (
          <>
            {t("pages.explorer.workflow.updated", {
              relativeTime: updatedAt,
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

        {hasUnsavedChanged && (
          <span className="text-center">
            {t("pages.explorer.workflow.editor.unsavedNote")}
          </span>
        )}
      </div>
    </Card>
  );
};
