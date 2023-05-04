import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "../../../../../design/Dropdown";
import { GitBranchPlus, GitMerge, Play, Save, Undo } from "lucide-react";

import Button from "../../../../../design/Button";
import { Card } from "../../../../../design/Card";
import Editor from "@monaco-editor/react";
import { FC } from "react";
import { RxChevronDown } from "react-icons/rx";
import { pages } from "../../../../../util/router/pages";
import { useNodeContent } from "../../../../../api/tree/query/get";
import { useTheme } from "../../../../../util/store/theme";

const WorkflowOverviewPage: FC = () => {
  const { path } = pages.explorer.useParams();
  const theme = useTheme();
  const { data } = useNodeContent({
    path,
  });

  return (
    <div className="flex grow flex-col space-y-4 p-4">
      <Card className="grow p-4">
        <Editor
          options={{
            scrollBeyondLastLine: false,
            cursorBlinking: "smooth",
          }}
          loading=""
          language="yaml"
          theme={theme === "dark" ? "vs-dark" : "vs-light"}
          value={data?.revision?.source && atob(data?.revision?.source)}
        />
      </Card>
      <div className="flex justify-end gap-4">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline">
              <GitMerge />
              Revisions <RxChevronDown />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-56">
            <DropdownMenuLabel>Choose Feature</DropdownMenuLabel>
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
        <Button variant="outline">
          <Save />
          Save
        </Button>
      </div>
    </div>
  );
};

export default WorkflowOverviewPage;
