import { ContextMenu, TableHeader } from "../components/PluginsTable";
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
  availablePlugins,
  outboundPluginTypes,
} from "../../../schema/plugins/outbound";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { EndpointFormSchemaType } from "../../../schema";
import { JsOutboundForm } from "./JsOutboundForm";
import { ModalWrapper } from "~/components/ModalWrapper";
import { OutboundPluginFormSchemaType } from "../../../schema/plugins/outbound/schema";
import { PluginSelector } from "../components/PluginSelector";
import { Plus } from "lucide-react";
import { getJsOutboundConfigAtIndex } from "../utils";
import { useTranslation } from "react-i18next";

type OutboundPluginFormProps = {
  form: UseFormReturn<EndpointFormSchemaType>;
};

export const OutboundPluginForm: FC<OutboundPluginFormProps> = ({ form }) => {
  const { t } = useTranslation();
  const { control } = form;
  const {
    append: addPlugin,
    remove: deletePlugin,
    move: movePlugin,
    update: editPlugin,
    fields,
  } = useFieldArray({
    control,
    name: "plugins.outbound",
  });
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editIndex, setEditIndex] = useState<number>();

  const [selectedPlugin, setSelectedPlugin] =
    useState<OutboundPluginFormSchemaType["type"]>();

  const pluginsCount = fields.length;
  const formId = "outboundPluginForm";

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
              "pages.explorer.endpoint.editor.form.plugins.outbound.table.headline",
              {
                count: pluginsCount,
              }
            )}
          >
            <DialogTrigger asChild>
              <Button icon variant="outline" size="sm">
                <Plus />
                {t(
                  "pages.explorer.endpoint.editor.form.plugins.outbound.table.addButton"
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
                      `pages.explorer.endpoint.editor.form.plugins.outbound.types.${type}`
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    <ContextMenu
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
                "pages.explorer.endpoint.editor.form.plugins.outbound.modal.headlineAdd"
              )
            : t(
                "pages.explorer.endpoint.editor.form.plugins.outbound.modal.headlineEdit"
              )
        }
      >
        <PluginSelector
          title={t(
            "pages.explorer.endpoint.editor.form.plugins.outbound.modal.label"
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
                  "pages.explorer.endpoint.editor.form.plugins.outbound.modal.placeholder"
                )}
              />
            </SelectTrigger>
            <SelectContent>
              {availablePlugins.map(({ name: pluginName }) => (
                <SelectItem key={pluginName} value={pluginName}>
                  {t(
                    `pages.explorer.endpoint.editor.form.plugins.outbound.types.${pluginName}`
                  )}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </PluginSelector>
        {selectedPlugin === outboundPluginTypes.jsOutbound.name && (
          <JsOutboundForm
            formId={formId}
            defaultConfig={getJsOutboundConfigAtIndex(fields, editIndex)}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              if (editIndex === undefined) {
                addPlugin(configuration);
              } else {
                editPlugin(editIndex, configuration);
              }
              setEditIndex(undefined);
            }}
          />
        )}
      </ModalWrapper>
    </Dialog>
  );
};
