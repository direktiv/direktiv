import { Bug, GitBranchPlus, GitMerge, Play, Save, Undo } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { FC, useEffect, useState } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { RxChevronDown } from "react-icons/rx";
import { useCreateRevision } from "~/api/tree/mutate/createRevision";
import { useNodeContent } from "~/api/tree/query/node";
import { useRevertRevision } from "~/api/tree/mutate/revertRevision";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { useUpdateWorkflow } from "~/api/tree/mutate/updateWorkflow";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

// get type of useNodeContent return value
type NodeContentType = ReturnType<typeof useNodeContent>["data"];

const WorkflowEditor: FC<{
  data: NonNullable<NodeContentType>;
  path: string;
}> = ({ data, path }) => {
  const { t } = useTranslation();
  const [error, setError] = useState<string | undefined>();
  const [hasUnsavedChanged, setHasUnsavedChanged] = useState(false);
  const workflowData = data.revision?.source && atob(data?.revision?.source);
  const updatedAt = useUpdatedAt(data.revision?.createdAt);

  const { mutate: updateWorkflow, isLoading } = useUpdateWorkflow({
    onError: (error) => {
      error && setError(error);
    },
  });

  const [value, setValue] = useState<string | undefined>(workflowData);
  const theme = useTheme();

  const { mutate: createRevision } = useCreateRevision();
  const { mutate: revertRevision } = useRevertRevision();

  useEffect(() => {
    setHasUnsavedChanged(workflowData !== value);
  }, [value, workflowData]);

  const onSave = (toSave: string | undefined) => {
    if (toSave) {
      setError(undefined);
      updateWorkflow({
        path,
        fileContent: toSave,
      });
    }
  };

  return (
    <div className="relative flex grow flex-col space-y-4 p-4">
      <Card className="grow p-4">
        <Editor
          value={workflowData}
          onChange={(newData) => {
            setValue(newData);
          }}
          theme={theme ?? undefined}
          onSave={onSave}
        />
      </Card>
      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <div className="flex grow items-center justify-between gap-2 text-sm text-gray-8 dark:text-gray-dark-8">
          {data.revision?.createdAt && (
            <>
              {t("pages.explorer.workflow.updated", {
                relativeTime: updatedAt,
              })}
            </>
          )}
          {hasUnsavedChanged && (
            <span className="text-center">
              {t("pages.explorer.workflow.editor.unsavedNote")}
            </span>
          )}
        </div>
        {error && (
          <Popover defaultOpen>
            <PopoverTrigger asChild>
              <Button variant="destructive">
                <Bug />
                {t("pages.explorer.workflow.editor.theresOneIssue")}
              </Button>
            </PopoverTrigger>
            <PopoverContent asChild>
              <div className="flex">
                <div className="grow">{error}</div>
              </div>
            </PopoverContent>
          </Popover>
        )}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" disabled={hasUnsavedChanged}>
              <GitMerge />
              {t("pages.explorer.workflow.editor.revisionsBtn")}
              <RxChevronDown />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-60">
            <DropdownMenuItem
              onClick={() => {
                createRevision({
                  path,
                });
              }}
            >
              <GitBranchPlus className="mr-2 h-4 w-4" />
              {t("pages.explorer.workflow.editor.makeRevision")}
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => {
                revertRevision({
                  path,
                });
              }}
            >
              <Undo className="mr-2 h-4 w-4" />
              {t("pages.explorer.workflow.editor.revertToPrevious")}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <Button variant="outline">
          <Play />
          {t("pages.explorer.workflow.editor.runBtn")}
        </Button>
        <Button
          variant="outline"
          disabled={isLoading}
          onClick={() => {
            onSave(value);
          }}
        >
          <Save />
          {t("pages.explorer.workflow.editor.saveBtn")}
        </Button>
      </div>
    </div>
  );
};

export default WorkflowEditor;
