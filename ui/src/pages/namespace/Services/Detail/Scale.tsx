import { decode } from "js-base64";
import { serializeServiceFile } from "../../Explorer/Service/ServiceEditor/utils";
import { useNodeContent } from "~/api/tree/query/node";

const Scale = ({ path, scale }: { path: string; scale: number }) => {
  const { data: serviceData, isSuccess } = useNodeContent({
    path,
  });

  if (!isSuccess) return null;

  const fileContent = decode(serviceData.source ?? "");
  const [serviceConfig] = serializeServiceFile(fileContent);

  return (
    <div>
      {scale} / {serviceConfig?.scale ?? "-"}
    </div>
  );
};

export default Scale;
