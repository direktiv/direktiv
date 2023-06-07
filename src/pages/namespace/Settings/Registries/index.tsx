import { Boxes, PlusCircle, Trash } from "lucide-react";
import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useEffect, useState } from "react";
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Create from "./Create";
import Delete from "./Delete";
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

        <DialogTrigger
          asChild
          data-testid="registry-create"
          onClick={() => setCreateRegistry(true)}
        >
          <Button variant="ghost">
            <PlusCircle />
          </Button>
        </DialogTrigger>
      </div>

      <Card>
        <Table>
          <TableBody>
            {registries.data?.registries.map((item, i) => (
              <TableRow key={i}>
                <TableCell>{item.name}</TableCell>
                <TableCell className="w-0">
                  <DialogTrigger
                    asChild
                    data-testid="registry-delete"
                    onClick={() => {
                      setDeleteRegistry(item);
                    }}
                  >
                    <Button variant="ghost">
                      <Trash />
                    </Button>
                  </DialogTrigger>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
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
