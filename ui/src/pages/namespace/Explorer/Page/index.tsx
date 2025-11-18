import { DirektivPagesSchema, DirektivPagesType } from "./poc/schema";
import { decode, encode } from "js-base64";
import { jsonToYaml, yamlToJsonOrNull } from "../utils";

import { Card } from "~/design/Card";
import { FC } from "react";
import { NoPermissions } from "~/design/Table";
import PageEditor from "./poc/PageEditor";
import { PanelTop } from "lucide-react";
import { analyzePath } from "~/util/router/utils";
import { useFile } from "~/api/files/query/file";
import { useNamespace } from "~/util/store/namespace";
import { useParams } from "@tanstack/react-router";
import { useUpdateFile } from "~/api/files/mutate/updateFile";

const UIPage: FC = () => {
  const { _splat: path } = useParams({ strict: false });
  const namespace = useNamespace();
  const { segments } = analyzePath(path);
  const filename = segments[segments.length - 1];

  const {
    isAllowed,
    noPermissionMessage,
    data: file,
    isFetched: isPermissionCheckFetched,
    isPending,
  } = useFile({ path });

  const { mutate: updateFile } = useUpdateFile();

  if (isAllowed === false)
    return (
      <Card className="m-5 flex grow">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

  if (!namespace) return null;
  if (!isPermissionCheckFetched) return null;
  if (file?.type !== "page") return null;

  const parsedPage = DirektivPagesSchema.safeParse(
    yamlToJsonOrNull(decode(file.data))
  );

  if (!parsedPage.success) {
    console.error(parsedPage.error);
    throw new Error("File is not a valid page");
  }

  const handleSave = (page: DirektivPagesType) => {
    updateFile({
      path: file.path,
      payload: {
        data: encode(jsonToYaml(page)),
      },
    });
  };

  return (
    <>
      <div className="border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
        <div className="flex flex-col gap-5 max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <PanelTop className="h-5" />
            {filename?.relative}
          </h3>
        </div>
      </div>

      <PageEditor
        page={parsedPage.data}
        isPending={isPending}
        onSave={(page) => handleSave(page)}
      />
    </>
  );
};

export default UIPage;
