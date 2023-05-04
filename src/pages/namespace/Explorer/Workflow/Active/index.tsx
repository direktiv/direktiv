import Badge from "../../../../../design/Badge";
import { Card } from "../../../../../design/Card";
import Editor from "@monaco-editor/react";
import { FC } from "react";
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
      <Card className="p-4">
        <Badge>{data?.revision?.hash.slice(0, 8)}</Badge>
      </Card>
      <Editor
        options={{
          scrollBeyondLastLine: false,
          cursorBlinking: "smooth",
        }}
        language="yaml"
        theme={theme === "dark" ? "vs-dark" : "vs-light"}
        value={data?.revision?.source && atob(data?.revision?.source)}
      />
    </div>
  );
};

export default WorkflowOverviewPage;
