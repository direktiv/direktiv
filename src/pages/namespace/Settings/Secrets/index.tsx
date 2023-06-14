import { FC, useEffect, useState } from "react";
import { Table, TableBody } from "~/design/Table";

import { Card } from "~/design/Card";
import Create from "./Create";
import CreateItemButton from "../compopnents/CreateItemButton";
import Delete from "./Delete";
import { Dialog } from "~/design/Dialog";
import EmptyList from "../compopnents/EmptyList";
import ItemRow from "../compopnents/ItemRow";
import { SecretSchemaType } from "~/api/secrets/schema";
import { SquareAsterisk } from "lucide-react";
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

        <CreateItemButton
          data-testid="secret-create"
          onClick={() => setCreateSecret(true)}
        />
      </div>

      <Card>
        {secrets.data?.secrets.results.length ? (
          <Table>
            <TableBody>
              {secrets.data?.secrets.results.map((item, i) => (
                <ItemRow item={item} key={i} onDelete={setDeleteSecret} />
              ))}
            </TableBody>
          </Table>
        ) : (
          <EmptyList>{t("pages.settings.secrets.list.empty")}</EmptyList>
        )}
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
