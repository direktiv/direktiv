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
  inboundPluginTypes,
} from "../../../schema/plugins/inbound";
import {
  getAclConfigAtIndex,
  getEventFilterConfigAtIndex,
  getHeaderManipulationConfigAtIndex,
  getJsInboundConfigAtIndex,
  getRequestConvertConfigAtIndex,
} from "../utils";

import { AclForm } from "./AclForm";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { EndpointFormSchemaType } from "../../../schema";
import { EventFilterForm } from "./EventFilterForm";
import { HeaderManipulationForm } from "./HeaderManipulationForm";
import { InboundPluginFormSchemaType } from "../../../schema/plugins/inbound/schema";
import { JsInboundForm } from "./JsInboundForm";
import { ListContextMenu } from "~/components/ListContextMenu";
import { ModalWrapper } from "~/components/ModalWrapper";
import { PluginSelector } from "../components/PluginSelector";
import { Plus } from "lucide-react";
import { RequestConvertForm } from "./RequestConvertForm";
import { TableHeader } from "../components/PluginsTable";
import { useTranslation } from "react-i18next";

type InboundPluginFormProps = {
  form: UseFormReturn<EndpointFormSchemaType>;
  onSave: (value: EndpointFormSchemaType) => void;
};

export const InboundPluginForm: FC<InboundPluginFormProps> = ({
  form,
  onSave,
}) => {
  const { t } = useTranslation();
  const { control, handleSubmit: parentSubmit } = form;
  const {
    append: addPlugin,
    remove: deletePlugin,
    move: movePlugin,
    update: editPlugin,
    fields,
  } = useFieldArray({
    control,
    name: "plugins.inbound",
  });
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editIndex, setEditIndex] = useState<number>();

  const [selectedPlugin, setSelectedPlugin] =
    useState<InboundPluginFormSchemaType["type"]>();

  const { jsInbound, requestConvert, acl, eventFilter, headerManipulation } =
    inboundPluginTypes;

  const handleSubmit = (configuration: InboundPluginFormSchemaType) => {
    setDialogOpen(false);
    if (editIndex === undefined) {
      addPlugin(configuration);
    } else {
      editPlugin(editIndex, configuration);
    }
    parentSubmit(onSave)();
    setEditIndex(undefined);
  };

  const pluginsCount = fields.length;
  const formId = "inboundPluginForm";

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
              "pages.explorer.endpoint.editor.form.plugins.inbound.table.headline",
              {
                count: pluginsCount,
              }
            )}
          >
            <DialogTrigger asChild>
              <Button icon variant="outline" size="sm">
                <Plus />
                {t(
                  "pages.explorer.endpoint.editor.form.plugins.inbound.table.addButton"
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
                      `pages.explorer.endpoint.editor.form.plugins.inbound.types.${type}`
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
                "pages.explorer.endpoint.editor.form.plugins.inbound.modal.headlineAdd"
              )
            : t(
                "pages.explorer.endpoint.editor.form.plugins.inbound.modal.headlineEdit"
              )
        }
        onCancel={() => {
          setDialogOpen(false);
        }}
      >
        <PluginSelector
          title={t(
            "pages.explorer.endpoint.editor.form.plugins.inbound.modal.label"
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
                  "pages.explorer.endpoint.editor.form.plugins.inbound.modal.placeholder"
                )}
              />
            </SelectTrigger>
            <SelectContent>
              {availablePlugins.map(({ name: pluginName }) => (
                <SelectItem key={pluginName} value={pluginName}>
                  {t(
                    `pages.explorer.endpoint.editor.form.plugins.inbound.types.${pluginName}`
                  )}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </PluginSelector>
        {selectedPlugin === requestConvert.name && (
          <RequestConvertForm
            formId={formId}
            defaultConfig={getRequestConvertConfigAtIndex(fields, editIndex)}
            onSubmit={handleSubmit}
          />
        )}
        {selectedPlugin === jsInbound.name && (
          <JsInboundForm
            formId={formId}
            defaultConfig={getJsInboundConfigAtIndex(fields, editIndex)}
            onSubmit={handleSubmit}
          />
        )}
        {selectedPlugin === acl.name && (
          <AclForm
            formId={formId}
            defaultConfig={getAclConfigAtIndex(fields, editIndex)}
            onSubmit={handleSubmit}
          />
        )}

        {selectedPlugin === headerManipulation.name && (
          <HeaderManipulationForm
            formId={formId}
            defaultConfig={getHeaderManipulationConfigAtIndex(
              fields,
              editIndex
            )}
            onSubmit={handleSubmit}
          />
        )}
        {selectedPlugin === eventFilter.name && (
          <EventFilterForm
            formId={formId}
            defaultConfig={getEventFilterConfigAtIndex(fields, editIndex)}
            onSubmit={handleSubmit}
          />
        )}
      </ModalWrapper>
    </Dialog>
  );
};
