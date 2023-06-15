import { FC, useEffect, useState } from "react";
import { Table, TableBody } from "~/design/Table";

import { Boxes } from "lucide-react";
import { Card } from "~/design/Card";
import Create from "./Create";
import CreateItemButton from "../compopnents/CreateItemButton";
import Delete from "./Delete";
import { Dialog } from "~/design/Dialog";
import EmptyList from "../compopnents/EmptyList";
import ItemRow from "../compopnents/ItemRow";
import { RegistrySchemaType } from "~/api/registries/schema";
import { useDeleteRegistry } from "~/api/registries/mutate/deleteRegistry";
import { useRegistries } from "~/api/registries/query/get";
import { useTranslation } from "react-i18next";

const RegistriesList: FC = () => {
  const { t } = useTranslation();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteRegistry, setDeleteRegistry] = useState<RegistrySchemaType>();
  const [createRegistry, setCreateRegistry] = useState(false);

  const registries = useRegistries();

  const { mutate: deleteRegistryMutation } = useDeleteRegistry({
    onSuccess: () => {
      setDeleteRegistry(undefined);
      setDialogOpen(false);
    },
  });

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteRegistry(undefined);
      setCreateRegistry(false);
    }
  }, [dialogOpen]);

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <div className="mb-3 flex flex-row justify-between">
        <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <Boxes className="h-5" />
          {t("pages.settings.registries.list.title")}
        </h3>

        <CreateItemButton
          onClick={() => setCreateRegistry(true)}
          data-testid="registry-create"
        />
      </div>

      <Card>
        {registries.data?.registries.length ? (
          <Table>
            <TableBody>
              {registries.data?.registries.map((item, i) => (
                <ItemRow key={i} item={item} onDelete={setDeleteRegistry} />
              ))}
            </TableBody>
          </Table>
        ) : (
          <EmptyList>{t("pages.settings.registries.list.empty")}</EmptyList>
        )}
        {deleteRegistry && (
          <Delete
            name={deleteRegistry.name}
            onConfirm={() =>
              deleteRegistryMutation({ registry: deleteRegistry })
            }
          />
        )}
        {createRegistry && <Create onSuccess={() => setDialogOpen(false)} />}
      </Card>
    </Dialog>
  );
};

export default RegistriesList;
