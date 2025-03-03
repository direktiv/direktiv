import {
  Table,
  TableBody,
  TableCell,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import Alert from "~/design/Alert";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { Link } from "@tanstack/react-router";
import { jsonToYaml } from "../../Explorer/utils";
import { useInfo } from "~/api/gateway/query/getInfo";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const InfoPage = () => {
  const { t } = useTranslation();
  const { data } = useInfo();
  const theme = useTheme();
  const info = data?.data;
  const { spec, errors, file_path: filePath } = info || {};
  const { title, version, description = "" } = spec?.info || {};

  const specToYaml = spec ? jsonToYaml(spec) : "";

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <div className="flex flex-col gap-4 sm:flex-row w-full">
        <Card className=" lg:h-[calc(100vh-15.5rem)] lg:overflow-y-scroll w-1/2">
          <Table className="border-gray-5 dark:border-gray-dark-5">
            <TableBody>
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableHeaderCell>
                  {t("pages.gateway.info.columns.title")}
                </TableHeaderCell>
                <TableCell>{title}</TableCell>
              </TableRow>
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableHeaderCell>
                  {t("pages.gateway.info.columns.version")}
                </TableHeaderCell>
                <TableCell>{version}</TableCell>
              </TableRow>
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableHeaderCell>
                  {t("pages.gateway.info.columns.description")}
                </TableHeaderCell>
                <TableCell>{description}</TableCell>
              </TableRow>
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableHeaderCell>
                  {t("pages.gateway.info.columns.file")}
                </TableHeaderCell>
                <TableCell>
                  {filePath === "virtual" || !filePath ? (
                    <span>
                      {filePath ||
                        t(
                          "pages.explorer.tree.openapiSpecification.unknownFilePath"
                        )}
                    </span>
                  ) : (
                    <Link
                      className="whitespace-normal break-all hover:underline"
                      to="/n/$namespace/explorer/openapiSpecification/$"
                      from="/n/$namespace"
                      params={{ _splat: filePath }}
                    >
                      {filePath}
                    </Link>
                  )}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>

          {errors?.length ? (
            <Alert variant="error" className="m-2">
              <h3>{t("pages.gateway.info.columns.errors")}</h3>
              <p>
                <ul className="list-disc pl-4">
                  {errors.map((error, index) => (
                    <li key={index}>{error}</li>
                  ))}
                </ul>
              </p>
            </Alert>
          ) : null}
        </Card>

        <Card className="flex grow p-4 max-lg:h-[500px] w-1/2">
          <Editor
            value={specToYaml}
            theme={theme ?? undefined}
            options={{
              readOnly: true,
            }}
          />
        </Card>
      </div>
    </div>
  );
};

export default InfoPage;
