import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/design/Dialog";
import { Edit, Plus, Trash } from "lucide-react";
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
import { RequestConvertForm } from "./RequestConvertForm";
import { inboundPluginTypes } from "../../../schema/plugins/inbound";

type InboundPluginFormProps = {
  formControls: UseFormReturn<EndpointFormSchemaType>;
};

export const InboundPluginForm: FC<InboundPluginFormProps> = ({
  formControls,
}) => {
  const { control } = formControls;
  const {
    append: addPlugin,
    remove: deletePlugin,
    update: editPlugin,
    fields,
  } = useFieldArray({
    control,
    name: "plugins.inbound",
  });
  const [dialogOpen, setDialogOpen] = useState(false);
  const [edit, setEdit] = useState<number>();

  const [selectedPlugin, setSelectedPlugin] =
    useState<InboundPluginFormSchemaType["type"]>();

  const pluginsCount = fields.length;

  const defaultTest = fields.find((x, index) => {
    if (x.type === inboundPluginTypes.requestConvert && index === edit) {
      return true;
    }
    return false;
  });

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
        if (isOpen === false) setEdit(undefined);
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
      {fields.map(({ id, type }, index) => (
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
            onClick={() => {
              setSelectedPlugin(type);
              setDialogOpen(true);
              setEdit(index);
            }}
          >
            <Edit />
          </Button>
        </div>
      ))}
      EDIT {edit}
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {edit === undefined ? "add" : "edit"} Inbound Plugin
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
                  <SelectValue placeholder="please select a target plugin" />
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
              defaultConfig={defaultTest?.configuration}
              onSubmit={(configuration) => {
                setDialogOpen(false);
                console.log("🚀", edit);
                if (edit === undefined) {
                  addPlugin(configuration);
                } else {
                  editPlugin(edit, configuration);
                }
                setEdit(undefined);
              }}
            />
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
};
