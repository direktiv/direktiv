import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { BroadcastsSchemaType } from "~/api/broadcasts/schema";
import { Card } from "~/design/Card";
import { Checkbox } from "~/design/Checkbox";
import { FC } from "react";
import { Radio } from "lucide-react";
import { twMergeClsx } from "~/util/helpers";
import { useBroadcasts } from "~/api/broadcasts/query/useBroadcasts";
import { useTranslation } from "react-i18next";
import { useUpdateBroadcasts } from "~/api/broadcasts/mutate/updateBroadcasts";

const leftColClasses = "col-span-2";
const labelCellClasses = "place-self-center px-2";
const switchCellClasses = "place-self-center px-2";
const rowClasses = "grid grid-cols-5 gap-0 lg:pr-8 xl:pr-12";

const Broadcasts: FC = () => {
  const { t } = useTranslation();
  const { data } = useBroadcasts();
  const { mutate } = useUpdateBroadcasts();

  if (!data?.broadcast) return null;

  const broadcasts = data.broadcast;

  const toggleBroadcast = (key: keyof BroadcastsSchemaType) =>
    mutate({
      payload: {
        broadcast: {
          [key]: !broadcasts[key],
        },
      },
    });

  return (
    <>
      <div className="mb-3 flex flex-row justify-between">
        <h3 className="flex items-center gap-x-2 pb-2 pt-1 font-bold">
          <Radio className="h-5" />
          {t("pages.settings.broadcasts.title")}
        </h3>
      </div>

      <Card>
        <Table>
          <TableHead>
            <TableRow
              className={twMergeClsx(
                rowClasses,
                "hover:bg-inherit dark:hover:bg-inherit"
              )}
            >
              <TableHeaderCell className={leftColClasses}></TableHeaderCell>
              <TableHeaderCell className={labelCellClasses}>
                {t("pages.settings.broadcasts.columns.success")}
              </TableHeaderCell>
              <TableHeaderCell className={labelCellClasses}>
                {t("pages.settings.broadcasts.columns.start")}
              </TableHeaderCell>
              <TableHeaderCell className={labelCellClasses}>
                {t("pages.settings.broadcasts.columns.fail")}
              </TableHeaderCell>
            </TableRow>
          </TableHead>

          <TableBody>
            <TableRow className={rowClasses}>
              <TableCell className={leftColClasses}>
                {t("pages.settings.broadcasts.rows.instance")}
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.instance.success"
                  checked={broadcasts["instance.success"]}
                  onClick={() => toggleBroadcast("instance.success")}
                />
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.instance.started"
                  checked={broadcasts["instance.started"]}
                  onClick={() => toggleBroadcast("instance.started")}
                />
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.instance.failed"
                  checked={broadcasts["instance.failed"]}
                  onClick={() => toggleBroadcast("instance.failed")}
                />
              </TableCell>
            </TableRow>
          </TableBody>

          <TableHead>
            <TableRow
              className={twMergeClsx(
                rowClasses,
                "hover:bg-inherit dark:hover:bg-inherit"
              )}
            >
              <TableHeaderCell className={leftColClasses}></TableHeaderCell>
              <TableHeaderCell className={labelCellClasses}>
                {t("pages.settings.broadcasts.columns.create")}
              </TableHeaderCell>
              <TableHeaderCell className={labelCellClasses}>
                {t("pages.settings.broadcasts.columns.update")}
              </TableHeaderCell>
              <TableHeaderCell className={labelCellClasses}>
                {t("pages.settings.broadcasts.columns.delete")}
              </TableHeaderCell>
            </TableRow>
          </TableHead>

          <TableBody>
            <TableRow className={rowClasses}>
              <TableCell className={leftColClasses}>
                {t("pages.settings.broadcasts.rows.directory")}
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.directory.create"
                  checked={broadcasts["directory.create"]}
                  onClick={() => toggleBroadcast("directory.create")}
                />
              </TableCell>
              <TableCell>{/* not implemented: directory.update */}</TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.directory.delete"
                  checked={broadcasts["directory.delete"]}
                  onClick={() => toggleBroadcast("directory.delete")}
                />
              </TableCell>
            </TableRow>
            <TableRow className={rowClasses}>
              <TableCell className={leftColClasses}>
                {t("pages.settings.broadcasts.rows.workflow")}
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.workflow.create"
                  checked={broadcasts["workflow.create"]}
                  onClick={() => toggleBroadcast("workflow.create")}
                />
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.workflow.update"
                  checked={broadcasts["workflow.update"]}
                  onClick={() => toggleBroadcast("workflow.update")}
                />
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.workflow.delete"
                  checked={broadcasts["workflow.delete"]}
                  onClick={() => toggleBroadcast("workflow.delete")}
                />
              </TableCell>
            </TableRow>
            <TableRow className={rowClasses}>
              <TableCell className={leftColClasses}>
                {t("pages.settings.broadcasts.rows.instanceVariable")}
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.instance.variable.create"
                  checked={broadcasts["instance.variable.create"]}
                  onClick={() => toggleBroadcast("instance.variable.create")}
                />
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.instance.variable.update"
                  checked={broadcasts["instance.variable.update"]}
                  onClick={() => toggleBroadcast("instance.variable.update")}
                />
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.instance.variable.delete"
                  checked={broadcasts["instance.variable.delete"]}
                  onClick={() => toggleBroadcast("instance.variable.delete")}
                />
              </TableCell>
            </TableRow>
            <TableRow className={rowClasses}>
              <TableCell className={leftColClasses}>
                {t("pages.settings.broadcasts.rows.namespaceVariable")}
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.namespace.variable.create"
                  checked={broadcasts["namespace.variable.create"]}
                  onClick={() => toggleBroadcast("namespace.variable.create")}
                />
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.namespace.variable.update"
                  checked={broadcasts["namespace.variable.update"]}
                  onClick={() => toggleBroadcast("namespace.variable.update")}
                />
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.namespace.variable.delete"
                  checked={broadcasts["namespace.variable.delete"]}
                  onClick={() => toggleBroadcast("namespace.variable.delete")}
                />
              </TableCell>
            </TableRow>
            <TableRow className={rowClasses}>
              <TableCell className={leftColClasses}>
                {t("pages.settings.broadcasts.rows.workflowVariable")}
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.workflow.variable.create"
                  checked={broadcasts["workflow.variable.create"]}
                  onClick={() => toggleBroadcast("workflow.variable.create")}
                />
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.workflow.variable.update"
                  checked={broadcasts["workflow.variable.update"]}
                  onClick={() => toggleBroadcast("workflow.variable.update")}
                />
              </TableCell>
              <TableCell className={switchCellClasses}>
                <Checkbox
                  data-testid="check.workflow.variable.delete"
                  checked={broadcasts["workflow.variable.delete"]}
                  onClick={() => toggleBroadcast("workflow.variable.delete")}
                />
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </Card>
    </>
  );
};

export default Broadcasts;
