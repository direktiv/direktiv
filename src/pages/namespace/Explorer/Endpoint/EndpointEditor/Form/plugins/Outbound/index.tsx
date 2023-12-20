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

import Button from "~/design/Button";
import { EndpointFormSchemaType } from "../../../schema";
import { JsOutboundForm } from "./JsOutboundForm";
import { OutboundPluginFormSchemaType } from "../../../schema/plugins/outbound/schema";
import { getJsOutboundConfigAtIndex } from "../utils";
import { outboundPluginTypes } from "../../../schema/plugins/outbound";
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

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
        if (isOpen === false) setEditIndex(undefined);
        setDialogOpen(isOpen);
      }}
    >
      <div className="flex items-center gap-3">
        {pluginsCount} Outbound plugins
        <DialogTrigger asChild>
          <Button icon variant="outline">
            <Plus /> add outbound plugin
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
                "pages.explorer.endpoint.editor.form.plugins.outbound.headlineAdd"
              )
            : t(
                "pages.explorer.endpoint.editor.form.plugins.outbound.headlineEdit"
              )
        }
      >
        <PluginSelector
          title={t(
            "pages.explorer.endpoint.editor.form.plugins.outbound.label"
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
                  "pages.explorer.endpoint.editor.form.plugins.outbound.placeholder"
                )}
              />
            </SelectTrigger>
            <SelectContent>
              {Object.values(outboundPluginTypes).map((pluginType) => (
                <SelectItem key={pluginType} value={pluginType}>
                  {pluginType}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </PluginSelector>
        {selectedPlugin === outboundPluginTypes.jsOutbound && (
          <JsOutboundForm
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
