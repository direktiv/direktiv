import { Table as TableSchema, TableType } from "../../schema/blocks/table";

import { BlockEditFormProps } from "..";
import { DialogFooter } from "../components/Footer";
import { DialogHeader } from "../components/Header";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type TableEditFormProps = BlockEditFormProps<TableType>;

export const Table = ({
  action,
  block: propBlock,
  path,
  onSubmit,
}: TableEditFormProps) => {
  const { t } = useTranslation();
  const form = useForm<TableType>({
    resolver: zodResolver(TableSchema),
    defaultValues: propBlock,
  });

  return (
    <>
      <DialogHeader action={action} path={path} type={propBlock.type} />
      <div className="text-gray-10 dark:text-gray-10">
        {t("direktivPage.blockEditor.blockForms.table.description")}
      </div>

      <DialogFooter onSubmit={() => onSubmit(form.getValues())} />
    </>
  );
};
