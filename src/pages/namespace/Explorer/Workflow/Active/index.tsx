import { Bug, GitBranchPlus, GitMerge, Play, Save, Undo } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../../../../../design/Dropdown";
import { FC, useState } from "react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "../../../../../design/Popover";

import Button from "../../../../../design/Button";
import { Card } from "../../../../../design/Card";
import Editor from "../../../../../design/Editor";
import { RxChevronDown } from "react-icons/rx";
import moment from "moment";
import { pages } from "../../../../../util/router/pages";
import { useNodeContent } from "../../../../../api/tree/query/get";
import { useTheme } from "../../../../../util/store/theme";
import { useUpdateWorkflow } from "../../../../../api/tree/mutate/updateWorkflow";

const WorkflowOverviewPage: FC = () => {
  const { path } = pages.explorer.useParams();
  const { data } = useNodeContent({ path });
  const [error, setError] = useState<string | undefined>();

  const { mutate: updateWorkflow, isLoading } = useUpdateWorkflow({
    onError: (error) => {
      error && setError(error);
    },
  });

  const workflowData = data?.revision?.source && atob(data?.revision?.source);
  const [value, setValue] = useState<string | undefined>(workflowData);

  const handleEditorChange = (value: string | undefined) => {
    setValue(value);
  };

  const theme = useTheme();

  return (
    <div className="relative flex grow flex-col space-y-4 p-4">
      <Card className="grow p-4">
        <Editor
          value={workflowData}
          onChange={handleEditorChange}
          theme={theme ?? undefined}
        />
      </Card>
      <div className="flex flex-col items-center justify-end gap-4 sm:flex-row">
        <div className="grow text-sm text-gray-8 dark:text-gray-dark-8">
          {/* must use fromNow(true) because otherwise after saving, it sometimes shows Updated in a few seconds */}
          {data?.revision?.createdAt && (
            <>Updated {moment(data?.revision?.createdAt).fromNow(true)} ago</>
          )}
        </div>
        {error && (
          <Popover defaultOpen>
            <PopoverTrigger asChild>
              <Button variant="destructive">
                <Bug />
                There is one issue
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
            <Button variant="outline">
              <GitMerge />
              Revisions <RxChevronDown />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-56">
            <DropdownMenuItem>
              <GitBranchPlus className="mr-2 h-4 w-4" /> Make Revision
            </DropdownMenuItem>
            <DropdownMenuItem>
              <Undo className="mr-2 h-4 w-4" /> Revert
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <Button variant="outline">
          <Play />
          Run
        </Button>
        <Button
          variant="outline"
          disabled={isLoading}
          onClick={() => {
            if (value && path) {
              setError(undefined);
              updateWorkflow({
                path,
                fileContent: value,
              });
            }
          }}
        >
          <Save />
          Save
        </Button>
      </div>
    </div>
  );
};

export default WorkflowOverviewPage;
