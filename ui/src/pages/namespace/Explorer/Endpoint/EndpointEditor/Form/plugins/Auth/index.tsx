import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";
import { UseFormReturn, useFieldArray } from "react-hook-form";
import {
  WebhookAuthFormSchemaType,
  webhookAuthPluginNames,
} from "../../../schema/plugins/auth/webhookAuth";
import {
  authPluginTypes,
  useAvailablePlugins,
} from "../../../schema/plugins/auth";
import {
  getBasicAuthConfigAtIndex,
  getKeyAuthConfigAtIndex,
  getWebhookAuthConfigAtIndex,
} from "../utils";

import { AuthPluginFormSchemaType } from "../../../schema/plugins/auth/schema";
import { BasicAuthForm } from "./BasicAuthForm";
import { BasicAuthFormSchemaType } from "../../../schema/plugins/auth/basicAuth";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { EndpointFormSchemaType } from "../../../schema";
import { KeyAuthForm } from "./KeyAuthForm";
import { KeyAuthFormSchemaType } from "../../../schema/plugins/auth/keyAuth";
import { ListContextMenu } from "~/components/ListContextMenu";
import { ModalWrapper } from "~/components/ModalWrapper";
import { PluginSelector } from "../components/PluginSelector";
import { Plus } from "lucide-react";
import { TableHeader } from "../components/PluginsTable";
import { WebhookAuthForm } from "./WebhookAuthForm";
import { useTranslation } from "react-i18next";

type AuthPluginFormProps = {
  formControls: UseFormReturn<EndpointFormSchemaType>;
  onSave: (value: EndpointFormSchemaType) => void;
};

export const AuthPluginForm: FC<AuthPluginFormProps> = ({
  formControls,
  onSave,
}) => {
  const { t } = useTranslation();
  const availablePlugins = useAvailablePlugins();
  const { control, handleSubmit: handleParentSubmit } = formControls;
  const {
    append: addPlugin,
    remove: deletePlugin,
    move: movePlugin,
    update: editPlugin,
    fields,
  } = useFieldArray({
    control,
    name: "plugins.auth",
  });
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editIndex, setEditIndex] = useState<number>();

  const [selectedPlugin, setSelectedPlugin] =
    useState<AuthPluginFormSchemaType["type"]>();

  const pluginsCount = fields.length;
  const formId = "authPluginForm";

  type PluginConfigSchema =
    | BasicAuthFormSchemaType
    | KeyAuthFormSchemaType
    | WebhookAuthFormSchemaType;

  const handleSubmit = (configuration: PluginConfigSchema) => {
    setDialogOpen(false);
    if (editIndex === undefined) {
      addPlugin(configuration);
    } else {
      editPlugin(editIndex, configuration);
    }
    handleParentSubmit(onSave)();
    setEditIndex(undefined);
  };

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
        if (isOpen === false) setEditIndex(undefined);
        setDialogOpen(isOpen);
      }}
    >
      <Card noShadow>
        <Table>
          <TableHeader
            title={t(
              "pages.explorer.endpoint.editor.form.plugins.auth.table.headline",
              {
                count: pluginsCount,
              }
            )}
          >
            <DialogTrigger asChild>
              <Button icon variant="outline" size="sm">
                <Plus />
                {t(
                  "pages.explorer.endpoint.editor.form.plugins.auth.table.addButton"
                )}
              </Button>
            </DialogTrigger>
          </TableHeader>
          <TableBody>
            {fields.map(({ id, type }, index, srcArray) => {
              const canMoveDown = index < srcArray.length - 1;
              const canMoveUp = index > 0;
              const onMoveUp = canMoveUp
                ? () => {
                    movePlugin(index, index - 1);
                  }
                : undefined;
              const onMoveDown = canMoveDown
                ? () => {
                    movePlugin(index, index + 1);
                  }
                : undefined;
              const onDelete = () => {
                deletePlugin(index);
              };

              return (
                <TableRow
                  key={id}
                  className="cursor-pointer"
                  onClick={() => {
                    setSelectedPlugin(type);
                    setDialogOpen(true);
                    setEditIndex(index);
                  }}
                >
                  <TableCell>
                    {t(
                      `pages.explorer.endpoint.editor.form.plugins.auth.types.${type}`
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    <ListContextMenu
                      onDelete={onDelete}
                      onMoveDown={onMoveDown}
                      onMoveUp={onMoveUp}
                    />
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </Card>

      <ModalWrapper
        formId={formId}
        showSaveBtn={!!selectedPlugin}
        title={
          editIndex === undefined
            ? t(
                "pages.explorer.endpoint.editor.form.plugins.auth.modal.headlineAdd"
              )
            : t(
                "pages.explorer.endpoint.editor.form.plugins.auth.modal.headlineEdit"
              )
        }
        onCancel={() => {
          setDialogOpen(false);
        }}
      >
        <PluginSelector
          title={t(
            "pages.explorer.endpoint.editor.form.plugins.auth.modal.label"
          )}
        >
          <Select
            onValueChange={(e) => {
              setSelectedPlugin(e as typeof selectedPlugin);
            }}
            value={selectedPlugin}
          >
            <SelectTrigger variant="outline" className="grow">
              <SelectValue
                placeholder={t(
                  "pages.explorer.endpoint.editor.form.plugins.auth.modal.placeholder"
                )}
              />
            </SelectTrigger>
            <SelectContent>
              {availablePlugins.map(({ name: pluginName }) => (
                <SelectItem key={pluginName} value={pluginName}>
                  {t(
                    `pages.explorer.endpoint.editor.form.plugins.auth.types.${pluginName}`
                  )}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </PluginSelector>
        {selectedPlugin === authPluginTypes.basicAuth.name && (
          <BasicAuthForm
            formId={formId}
            defaultConfig={getBasicAuthConfigAtIndex(fields, editIndex)}
            onSubmit={handleSubmit}
          />
        )}
        {selectedPlugin === authPluginTypes.keyAuth.name && (
          <KeyAuthForm
            formId={formId}
            defaultConfig={getKeyAuthConfigAtIndex(fields, editIndex)}
            onSubmit={handleSubmit}
          />
        )}
        {webhookAuthPluginNames.map(
          (name) =>
            selectedPlugin === name && (
              <WebhookAuthForm
                key={name}
                type={name}
                formId={formId}
                defaultConfig={getWebhookAuthConfigAtIndex(
                  name,
                  fields,
                  editIndex
                )}
                onSubmit={handleSubmit}
              />
            )
        )}
      </ModalWrapper>
    </Dialog>
  );
};
