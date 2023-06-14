import { Bug, GitBranchPlus, Play, Save, Tag, Undo } from "lucide-react";
import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { FC, useEffect, useState } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import RunWorkflow from "../components/RunWorkflow";
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
  const workflowData = atob(data?.revision?.source ?? "");
  const updatedAt = useUpdatedAt(data.revision?.createdAt);

  const { mutate: updateWorkflow, isLoading } = useUpdateWorkflow({
    onError: (error) => {
      error && setError(error);
    },
  });

  const [value, setValue] = useState(workflowData);
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
    <div className="relative flex grow flex-col space-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
        <Tag className="h-5" />
        {t("pages.explorer.workflow.headline")}
      </h3>
      <Card className="grow p-4" data-testid="workflow-editor">
        <Editor
          value={workflowData}
          onMount={(editor) => {
            editor.focus();
          }}
          onChange={(newData) => {
            setValue(newData ?? "");
          }}
          theme={theme ?? undefined}
          onSave={onSave}
        />
      </Card>
      <div className="flex flex-col justify-end gap-4 md:flex-row md:items-center">
        <div
          data-testid="workflow-txt-updated"
          className="flex grow items-center justify-between gap-2 text-sm text-gray-8 dark:text-gray-dark-8"
        >
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
              <div className="flex p-4">
                <div className="grow">{error}</div>
              </div>
            </PopoverContent>
          </Popover>
        )}
        <DropdownMenu>
          <ButtonBar>
            <Button
              variant="outline"
              disabled={hasUnsavedChanged}
              onClick={() => {
                createRevision({
                  path,
                });
              }}
              className="grow"
              data-testid="workflow-editor-btn-make-revision"
            >
              <GitBranchPlus />
              {t("pages.explorer.workflow.editor.makeRevision")}
            </Button>
            <DropdownMenuTrigger asChild>
              <Button
                disabled={hasUnsavedChanged}
                variant="outline"
                data-testid="workflow-editor-btn-revision-drop"
              >
                <RxChevronDown />
              </Button>
            </DropdownMenuTrigger>
          </ButtonBar>
          <DropdownMenuContent className="w-60">
            <DropdownMenuItem
              onClick={() => {
                revertRevision({
                  path,
                });
              }}
              data-testid="workflow-editor-btn-revert-revision"
            >
              <Undo className="mr-2 h-4 w-4" />
              {t("pages.explorer.workflow.editor.revertToPrevious")}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <Dialog>
          <DialogTrigger asChild>
            <Button variant="outline">
              <Play />
              {t("pages.explorer.workflow.editor.runBtn")}
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-2xl">
            <RunWorkflow path={path} />
          </DialogContent>
        </Dialog>
        <Button
          variant="outline"
          disabled={isLoading}
          onClick={() => {
            onSave(value);
          }}
          data-testid="workflow-editor-btn-save"
        >
          <Save />
          {t("pages.explorer.workflow.editor.saveBtn")}
        </Button>
      </div>
    </div>
  );
};

export default WorkflowEditor;
