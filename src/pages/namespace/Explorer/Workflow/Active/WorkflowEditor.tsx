import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { FC, useEffect, useState } from "react";
import { GitBranchPlus, Play, Save, Tag, Undo } from "lucide-react";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { CodeEditor } from "./CodeEditor";
import { Diagram } from "./Diagram";
import { EditorLayoutSwitcher } from "~/componentsNext/EditorLayoutSwitcher";
import RunWorkflow from "../components/RunWorkflow";
import { RxChevronDown } from "react-icons/rx";
import { WorkspaceLayout } from "~/componentsNext/WorkspaceLayout";
import { useCreateRevision } from "~/api/tree/mutate/createRevision";
import { useEditorLayout } from "~/util/store/editor";
import { useNodeContent } from "~/api/tree/query/node";
import { useRevertRevision } from "~/api/tree/mutate/revertRevision";
import { useTranslation } from "react-i18next";
import { useUpdateWorkflow } from "~/api/tree/mutate/updateWorkflow";

export type NodeContentType = ReturnType<typeof useNodeContent>["data"];

const WorkflowEditor: FC<{
  data: NonNullable<NodeContentType>;
  path: string;
}> = ({ data, path }) => {
  const currentLayout = useEditorLayout();
  const { t } = useTranslation();
  const [error, setError] = useState<string | undefined>();
  const [hasUnsavedChanged, setHasUnsavedChanged] = useState(false);

  const workflowData = atob(data?.revision?.source ?? "");

  const { mutate: updateWorkflow, isLoading } = useUpdateWorkflow({
    onError: (error) => {
      error && setError(error);
    },
  });

  const [value, setValue] = useState(workflowData);

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
      <h3 className="flex items-center gap-x-2 font-bold">
        <Tag className="h-5" />
        {t("pages.explorer.workflow.headline")}
      </h3>
      <WorkspaceLayout
        layout={currentLayout}
        diagramComponent={
          <Diagram workflowData={workflowData} layout={currentLayout} />
        }
        editorComponent={
          <CodeEditor
            value={workflowData}
            setValue={setValue}
            createdAt={data.revision?.createdAt}
            error={error}
            hasUnsavedChanged={hasUnsavedChanged}
            onSave={onSave}
          />
        }
      />

      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <EditorLayoutSwitcher />
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
            <Button variant="outline" data-testid="workflow-editor-btn-run">
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
