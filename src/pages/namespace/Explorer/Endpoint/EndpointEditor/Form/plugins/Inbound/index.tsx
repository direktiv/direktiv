import { ChevronDown, ChevronUp, Edit, Plus, Trash } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/design/Dialog";
import { FC, useState } from "react";
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
import { InboundPluginFormSchemaType } from "../../../schema/plugins/inbound/schema";
import { JsInboundForm } from "./JsInboundForm";
import { JsInboundFormSchemaType } from "../../../schema/plugins/inbound/jsInbound";
import { RequestConvertForm } from "./RequestConvertForm";
import { RequestConvertFormSchemaType } from "../../../schema/plugins/inbound/requestConvert";
import { inboundPluginTypes } from "../../../schema/plugins/inbound";

type InboundPluginFormProps = {
  formControls: UseFormReturn<EndpointFormSchemaType>;
};

// TODO: may create a factory for this, ot introduce a generic
const readRequestConvertConfig = (
  fields: InboundPluginFormSchemaType[] | undefined,
  index: number | undefined
): RequestConvertFormSchemaType["configuration"] | undefined => {
  const plugin = index ? fields?.[index] : undefined;
  return plugin?.type === inboundPluginTypes.requestConvert
    ? plugin.configuration
    : undefined;
};

const readJsInboundConfig = (
  fields: InboundPluginFormSchemaType[] | undefined,
  index: number | undefined
): JsInboundFormSchemaType["configuration"] | undefined => {
  const plugin = index ? fields?.[index] : undefined;
  return plugin?.type === inboundPluginTypes.jsInbound
    ? plugin.configuration
    : undefined;
};

export const InboundPluginForm: FC<InboundPluginFormProps> = ({
  formControls,
}) => {
  const { control } = formControls;
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
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {editIndex === undefined ? "add" : "edit"} Inbound Plugin
          </DialogTitle>
        </DialogHeader>
        <div className="my-3 flex flex-col gap-y-5">
          <div className="flex flex-col gap-y-5">
            <fieldset className="flex items-center gap-5">
              <label className="w-[150px] overflow-hidden text-right text-sm">
                select a inbound plugin
              </label>
              <Select
                onValueChange={(e) => {
                  setSelectedPlugin(e as typeof selectedPlugin);
                }}
                value={selectedPlugin}
              >
                <SelectTrigger variant="outline">
                  <SelectValue placeholder="please select a inbound plugin" />
                </SelectTrigger>
                <SelectContent>
                  {Object.values(inboundPluginTypes).map((pluginType) => (
                    <SelectItem key={pluginType} value={pluginType}>
                      {pluginType}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </fieldset>
          </div>
          {selectedPlugin === inboundPluginTypes.requestConvert && (
            <RequestConvertForm
              defaultConfig={readRequestConvertConfig(fields, editIndex)}
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

          {selectedPlugin === inboundPluginTypes.jsInbound && (
            <JsInboundForm
              defaultConfig={readJsInboundConfig(fields, editIndex)}
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
        </div>
      </DialogContent>
    </Dialog>
  );
};
