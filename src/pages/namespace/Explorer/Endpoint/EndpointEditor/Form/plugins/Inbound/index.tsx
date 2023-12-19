import { ChevronDown, ChevronUp, Edit, Plus, Trash } from "lucide-react";
import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";
import { ModalWrapper, PluginSelector } from "../components/Modal";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import { UseFormReturn, useFieldArray } from "react-hook-form";
import {
  getJsInboundConfigAtIndex,
  getRequestConvertConfigAtIndex,
} from "../utils";

import Button from "~/design/Button";
import { EndpointFormSchemaType } from "../../../schema";
import { InboundPluginFormSchemaType } from "../../../schema/plugins/inbound/schema";
import { JsInboundForm } from "./JsInboundForm";
import { RequestConvertForm } from "./RequestConvertForm";
import { inboundPluginTypes } from "../../../schema/plugins/inbound";
import { useTranslation } from "react-i18next";

type InboundPluginFormProps = {
  form: UseFormReturn<EndpointFormSchemaType>;
};

export const InboundPluginForm: FC<InboundPluginFormProps> = ({ form }) => {
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
    name: "plugins.inbound",
  });
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editIndex, setEditIndex] = useState<number>();

  const [selectedPlugin, setSelectedPlugin] =
    useState<InboundPluginFormSchemaType["type"]>();

  const { jsInbound, requestConvert } = inboundPluginTypes;

  const pluginsCount = fields.length;

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
        if (isOpen === false) setEditIndex(undefined);
        setDialogOpen(isOpen);
      }}
    >
      <div className="flex items-center gap-3">
        {pluginsCount} Inbound plugins
        <DialogTrigger asChild>
          <Button icon variant="outline">
            <Plus /> add inbound plugin
          </Button>
        </DialogTrigger>
      </div>
      {fields.map(({ id, type }, index, srcArray) => {
        const canMoveDown = index < srcArray.length - 1;
        const canMoveUp = index > 0;

        return (
          <div key={id} className="flex gap-2">
            {type}
            <Button
              variant="destructive"
              icon
              size="sm"
              onClick={() => {
                deletePlugin(index);
              }}
            >
              <Trash />
            </Button>
            <Button
              variant="outline"
              icon
              size="sm"
              disabled={!canMoveDown}
              onClick={() => {
                movePlugin(index, index + 1);
              }}
            >
              <ChevronDown />
            </Button>
            <Button
              variant="outline"
              icon
              size="sm"
              disabled={!canMoveUp}
              onClick={() => {
                movePlugin(index, index - 1);
              }}
            >
              <ChevronUp />
            </Button>
            <Button
              variant="outline"
              icon
              size="sm"
              onClick={() => {
                setSelectedPlugin(type);
                setDialogOpen(true);
                setEditIndex(index);
              }}
            >
              <Edit />
            </Button>
          </div>
        );
      })}

      <ModalWrapper
        title={
          editIndex === undefined
            ? t(
                "pages.explorer.endpoint.editor.form.plugins.inbound.headlineAdd"
              )
            : t(
                "pages.explorer.endpoint.editor.form.plugins.inbound.headlineEdit"
              )
        }
      >
        <PluginSelector
          title={t("pages.explorer.endpoint.editor.form.plugins.inbound.label")}
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
                  "pages.explorer.endpoint.editor.form.plugins.inbound.placeholder"
                )}
              />
            </SelectTrigger>
            <SelectContent>
              {Object.values(inboundPluginTypes).map((pluginType) => (
                <SelectItem key={pluginType} value={pluginType}>
                  {t(
                    `pages.explorer.endpoint.editor.form.plugins.inbound.types.${pluginType}`
                  )}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </PluginSelector>
        {selectedPlugin === requestConvert && (
          <RequestConvertForm
            defaultConfig={getRequestConvertConfigAtIndex(fields, editIndex)}
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
        {selectedPlugin === jsInbound && (
          <JsInboundForm
            defaultConfig={getJsInboundConfigAtIndex(fields, editIndex)}
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
