import { decode } from "js-base64";
import { serializeServiceFile } from "../../Explorer/Service/ServiceEditor/utils";
import { useFile } from "~/api/files/query/file";

const Scale = ({ path, scale }: { path: string; scale: number }) => {
  const { data: serviceData, isSuccess } = useFile({
    path,
  });

  if (!isSuccess) return null;
  if (serviceData?.type === "directory") return null;

  const fileContent = decode(serviceData.data ?? "");
  const [serviceConfig] = serializeServiceFile(fileContent);

  return (
    <div>
      {scale} / {serviceConfig?.scale ?? "-"}
    </div>
  );
};

export default Scale;
