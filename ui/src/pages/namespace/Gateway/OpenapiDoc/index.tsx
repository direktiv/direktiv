import { jsonToYaml, yamlToJsonOrNull } from "../../Explorer/utils";

import { Card } from "~/design/Card";
import { RapiDoc } from "~/design/RapiDoc";
import { ScrollText } from "lucide-react";
import { useInfo } from "~/api/gateway/query/getInfo";
import { useNamespace } from "~/util/store/namespace";
import { useTheme } from "~/util/store/theme";

const OpenapiDocPage = () => {
  const { data } = useInfo();
  const theme = useTheme();
  const namespace = useNamespace();
  const info = data?.data;
  const { spec, errors } = info || {};

  // Clone the JSON spec to avoid mutating the original.
  const updatedSpec = JSON.parse(JSON.stringify(JsonSpec));
  updateRefs(updatedSpec, baseFileUrl);

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <div className="flex flex-col gap-4 sm:flex-row">
        <h3 className="flex grow items-center gap-x-2 pb-1 font-bold">
          <ScrollText className="h-5" />
          OpenAPI Documentation
        </h3>
      </div>
      <div className="flex flex-col gap-4 sm:flex-row w-full">
        <Card className="size-full">
          <RapiDoc spec={updatedSpec} />
        </Card>
      </div>
    </div>
  );
};

export default OpenapiDocPage;
