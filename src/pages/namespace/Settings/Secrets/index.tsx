import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useEffect, useState } from "react";
import { PlusCircle, SquareAsterisk } from "lucide-react";
import { Table, TableBody } from "~/design/Table";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Create from "./Create";
import Delete from "./Delete";
import ItemRow from "../compopnents/ItemRow";
import { SecretSchemaType } from "~/api/secrets/schema";
import { useDeleteSecret } from "~/api/secrets/mutate/deleteSecret";
import { useSecrets } from "~/api/secrets/query/get";
import { useTranslation } from "react-i18next";

const SecretsList: FC = () => {
  const { t } = useTranslation();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteSecret, setDeleteSecret] = useState<SecretSchemaType>();
  const [createSecret, setCreateSecret] = useState(false);

  const secrets = useSecrets();

  const { mutate: deleteSecretMutation } = useDeleteSecret({
    onSuccess: () => {
      setDeleteSecret(undefined);
      setDialogOpen(false);
    },
  });

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteSecret(undefined);
      setCreateSecret(false);
    }
  }, [dialogOpen]);

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <div className="mb-3 flex flex-row justify-between">
        <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <SquareAsterisk className="h-5" />
          {t("pages.settings.secrets.list.title")}
        </h3>

        <DialogTrigger
          asChild
          data-testid="secret-create"
          onClick={() => setCreateSecret(true)}
        >
          <Button variant="ghost">
            <PlusCircle />
          </Button>
        </DialogTrigger>
      </div>

      <Card>
        <Table>
          <TableBody>
            {secrets.data?.secrets.results.map((item, i) => (
              <ItemRow item={item} key={i} onDelete={setDeleteSecret} />
            ))}
          </TableBody>
        </Table>
        {deleteSecret && (
          <Delete
            name={deleteSecret.name}
            onConfirm={() => deleteSecretMutation({ secret: deleteSecret })}
          />
        )}
        {createSecret && <Create onSuccess={() => setDialogOpen(false)} />}
      </Card>
    </Dialog>
  );
};

export default SecretsList;
