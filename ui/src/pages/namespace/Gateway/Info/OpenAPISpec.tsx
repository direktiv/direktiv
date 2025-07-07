import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { jsonToYaml } from "../../Explorer/utils";
import { useInfo } from "~/api/gateway/query/getInfo";
import { useTheme } from "~/util/store/theme";

const OpenAPISpec = () => {
  const { data } = useInfo();
  const theme = useTheme();
  const info = data?.data;
  const { spec } = info || {};

  const specToYaml = spec ? jsonToYaml(spec) : "";

  return (
    <Card className="flex h-96 w-full grow p-4 lg:h-[calc(100vh-11.8rem)]">
      <Editor
        value={specToYaml}
        theme={theme ?? undefined}
        options={{
          readOnly: true,
        }}
      />
    </Card>
  );
};

export default OpenAPISpec;
